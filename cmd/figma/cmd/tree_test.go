package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
)

// buildTestTree creates a simple node hierarchy for testing.
func buildTestTree() *api.Node {
	return &api.Node{
		ID:   "0:0",
		Name: "Document",
		Type: "DOCUMENT",
		Children: []api.Node{
			{
				ID:   "1:1",
				Name: "Page 1",
				Type: "CANVAS",
				Children: []api.Node{
					{
						ID:   "2:1",
						Name: "Frame 1",
						Type: "FRAME",
						Children: []api.Node{
							{
								ID:   "3:1",
								Name: "Text Layer",
								Type: "TEXT",
							},
							{
								ID:   "3:2",
								Name: "Rectangle",
								Type: "RECTANGLE",
							},
						},
					},
					{
						ID:   "2:2",
						Name: "Frame 2",
						Type: "FRAME",
						Children: []api.Node{
							{
								ID:   "3:3",
								Name: "Component",
								Type: "COMPONENT",
							},
						},
					},
				},
			},
		},
	}
}

func TestFilterTree(t *testing.T) {
	root := buildTestTree()

	tests := []struct {
		name       string
		typeFilter string
		maxDepth   int
		wantIDs    []string // IDs of nodes expected in filtered tree (depth-first)
	}{
		{
			name:       "no filter",
			typeFilter: "",
			maxDepth:   -1,
			wantIDs:    []string{"0:0", "1:1", "2:1", "3:1", "3:2", "2:2", "3:3"},
		},
		{
			name:       "filter FRAME",
			typeFilter: "FRAME",
			maxDepth:   -1,
			wantIDs:    []string{"0:0", "1:1", "2:1", "2:2"}, // includes ancestors and matching nodes
		},
		{
			name:       "filter TEXT",
			typeFilter: "TEXT",
			maxDepth:   -1,
			wantIDs:    []string{"0:0", "1:1", "2:1", "3:1"},
		},
		{
			name:       "max depth 0",
			typeFilter: "",
			maxDepth:   0,
			wantIDs:    []string{"0:0"},
		},
		{
			name:       "max depth 1",
			typeFilter: "",
			maxDepth:   1,
			wantIDs:    []string{"0:0", "1:1"},
		},
		{
			name:       "max depth 2 with FRAME filter",
			typeFilter: "FRAME",
			maxDepth:   2,
			wantIDs:    []string{"0:0", "1:1", "2:1", "2:2"},
		},
		{
			name:       "no match",
			typeFilter: "ELLIPSE",
			maxDepth:   -1,
			wantIDs:    []string{}, // filteredRoot should be nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterTree(root, tt.typeFilter, tt.maxDepth)
			if len(tt.wantIDs) == 0 {
				if filtered != nil {
					t.Errorf("filterTree() expected nil, got node with ID %s", filtered.ID)
				}
				return
			}
			if filtered == nil {
				t.Fatal("filterTree() returned nil, expected nodes")
			}
			// Collect IDs via depth-first traversal
			var gotIDs []string
			var dfs func(node *api.Node)
			dfs = func(node *api.Node) {
				if node == nil {
					return
				}
				gotIDs = append(gotIDs, node.ID)
				for i := range node.Children {
					dfs(&node.Children[i])
				}
			}
			dfs(filtered)
			if len(gotIDs) != len(tt.wantIDs) {
				t.Errorf("filterTree() got %d IDs %v, want %d IDs %v", len(gotIDs), gotIDs, len(tt.wantIDs), tt.wantIDs)
				return
			}
			for i := range gotIDs {
				if gotIDs[i] != tt.wantIDs[i] {
					t.Errorf("filterTree() ID mismatch at index %d: got %s, want %s", i, gotIDs[i], tt.wantIDs[i])
				}
			}
		})
	}
}

func TestToTreeNode(t *testing.T) {
	root := buildTestTree()
	tn := toTreeNode(root)
	if tn == nil {
		t.Fatal("toTreeNode returned nil")
	}
	if tn.ID != "0:0" {
		t.Errorf("toTreeNode root ID = %s, want 0:0", tn.ID)
	}
	if len(tn.Children) != 1 {
		t.Errorf("toTreeNode root children count = %d, want 1", len(tn.Children))
	}
	// Quick check that structure is preserved
	page := tn.Children[0]
	if page.ID != "1:1" {
		t.Errorf("page ID = %s, want 1:1", page.ID)
	}
	if len(page.Children) != 2 {
		t.Errorf("page children count = %d, want 2", len(page.Children))
	}
}

func TestRunTree(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "file_response.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Skipf("Could not read fixture: %v", err)
	}

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Request: %s %s", r.Method, r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/v1/files/") && r.Method == http.MethodGet && !strings.Contains(r.URL.Path, "/nodes") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
		// For node endpoint (not needed for this test)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"err": "Not Found"}`))
	}))
	defer ts.Close()

	// Create config with test server URL, using defaults
	cfg := config.DefaultConfig()
	cfg.API.BaseURL = ts.URL + "/v1"
	cfg.API.Timeout = 30 * time.Second
	cfg.API.Tier = 1
	cfg.API.SeatType = "dev_full"
	cfg.Token = "testtoken"
	cfg.TokenType = "pat"
	cfg.OutputFormat = config.OutputFormatText

	// Create a no-op logger
	logger := logging.NewNopLogger()

	// Create command context
	ctx := context.Background()
	ctx = context.WithValue(ctx, configKey{}, &cfg)
	ctx = context.WithValue(ctx, loggerKey{}, logger)

	// Create tree command and set context
	cmd := treeCmd
	cmd.SetContext(ctx)

	// Set flags (defaults)
	cmd.Flags().Set("node-id", "")
	cmd.Flags().Set("type", "")
	cmd.Flags().Set("max-depth", "-1")
	cmd.Flags().Set("output", "")

	// Execute with a file key that matches the fixture (the fixture uses no specific key)
	err = runTree(cmd, []string{"ABC123"})
	if err != nil {
		t.Errorf("runTree failed: %v", err)
	}
}

func TestRunTreeWithFilters(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "file_response.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Skipf("Could not read fixture: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/files/") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	cfg := config.DefaultConfig()
	cfg.API.BaseURL = ts.URL + "/v1"
	cfg.API.Timeout = 30 * time.Second
	cfg.Token = "testtoken"
	cfg.TokenType = "pat"
	cfg.OutputFormat = config.OutputFormatText

	logger := logging.NewNopLogger()

	ctx := context.Background()
	ctx = context.WithValue(ctx, configKey{}, &cfg)
	ctx = context.WithValue(ctx, loggerKey{}, logger)

	cmd := treeCmd
	cmd.SetContext(ctx)

	// Test with type filter
	cmd.Flags().Set("type", "FRAME")
	err = runTree(cmd, []string{"ABC123"})
	if err != nil {
		t.Errorf("runTree with type filter failed: %v", err)
	}
	// Reset flags for next test
	cmd.Flags().Set("type", "")
	cmd.Flags().Set("max-depth", "2")
	err = runTree(cmd, []string{"ABC123"})
	if err != nil {
		t.Errorf("runTree with max-depth failed: %v", err)
	}
}
