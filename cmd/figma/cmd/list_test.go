package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
)

func TestWildcardToRegex(t *testing.T) {
	tests := []struct {
		pattern string
		want    string
	}{
		{"hello", "^hello$"},
		{"*.txt", "^.*\\.txt$"},
		{"file*", "^file.*$"},
		{"*file", "^.*file$"},
		{"*file*", "^.*file.*$"},
		{"file?.txt", "^file.\\.txt$"},
		{"?", "^.$"},
		{"??", "^..$"},
		{"file[abc].txt", "^file[abc]\\.txt$"},
		{"file[a-z].txt", "^file[a-z]\\.txt$"},
		{"file.txt", "^file\\.txt$"},
		{"file(1).txt", "^file\\(1\\)\\.txt$"},
		{"file{2}.txt", "^file\\{2\\}\\.txt$"},
		{"file^$", "^file\\^\\$$"},
		{"*[0-9]?.txt", "^.*[0-9].\\.txt$"},
		{"", "^$"},
	}
	for _, tt := range tests {
		got := wildcardToRegex(tt.pattern)
		if got != tt.want {
			t.Errorf("wildcardToRegex(%q) = %q, want %q", tt.pattern, got, tt.want)
		}
		// Ensure generated regex compiles
		_, err := regexp.Compile(got)
		if err != nil {
			t.Errorf("wildcardToRegex(%q) produced invalid regex %q: %v", tt.pattern, got, err)
		}
	}
}

// Helper to create a node with given properties
func mknode(id, name, typ string) *api.Node {
	return &api.Node{
		ID:   id,
		Name: name,
		Type: typ,
	}
}

func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		name          string
		node          *api.Node
		typeFilter    string
		namePattern   string
		caseSensitive bool
		want          bool
	}{
		// No filters -> matches
		{
			name: "no filters",
			node: mknode("1", "Button", "FRAME"),
			want: true,
		},
		// Type filter match
		{
			name:       "type match",
			node:       mknode("1", "Button", "FRAME"),
			typeFilter: "FRAME",
			want:       true,
		},
		{
			name:       "type mismatch",
			node:       mknode("1", "Button", "FRAME"),
			typeFilter: "COMPONENT",
			want:       false,
		},
		// Name pattern match with wildcard
		{
			name:        "name pattern * match",
			node:        mknode("1", "SubmitButton", "FRAME"),
			namePattern: "*Button",
			want:        true,
		},
		{
			name:        "name pattern * mismatch",
			node:        mknode("1", "SubmitButton", "FRAME"),
			namePattern: "Button*",
			want:        false,
		},
		{
			name:        "name pattern ? match",
			node:        mknode("1", "A", "FRAME"),
			namePattern: "?",
			want:        true,
		},
		{
			name:        "name pattern ? mismatch",
			node:        mknode("1", "AB", "FRAME"),
			namePattern: "?",
			want:        false,
		},
		// Case sensitivity
		{
			name:          "case-insensitive match",
			node:          mknode("1", "Button", "FRAME"),
			namePattern:   "button",
			caseSensitive: false,
			want:          true,
		},
		{
			name:          "case-sensitive mismatch",
			node:          mknode("1", "Button", "FRAME"),
			namePattern:   "button",
			caseSensitive: true,
			want:          false,
		},
		// Combined filters
		{
			name:        "type and name match",
			node:        mknode("1", "SubmitButton", "FRAME"),
			typeFilter:  "FRAME",
			namePattern: "*Button",
			want:        true,
		},
		{
			name:        "type match name mismatch",
			node:        mknode("1", "SubmitButton", "FRAME"),
			typeFilter:  "FRAME",
			namePattern: "Button*",
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ListOptions{
				TypeFilter:    tt.typeFilter,
				NamePattern:   tt.namePattern,
				CaseSensitive: tt.caseSensitive,
			}
			var nameRegex *regexp.Regexp
			if opts.NamePattern != "" {
				pattern := wildcardToRegex(opts.NamePattern)
				if !opts.CaseSensitive {
					pattern = "(?i)" + pattern
				}
				r, err := regexp.Compile(pattern)
				if err != nil {
					// Invalid pattern, treat as no matches (as per collectNodes)
					nameRegex = nil
				} else {
					nameRegex = r
				}
			}
			got := matchesFilter(tt.node, opts, nameRegex)
			if got != tt.want {
				t.Errorf("matchesFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectNodes(t *testing.T) {
	// Build a tree with various node types and depths
	root := &api.Node{
		ID:   "0",
		Name: "Root",
		Type: "DOCUMENT",
		Children: []api.Node{
			{
				ID:   "1",
				Name: "FrameA",
				Type: "FRAME",
				Children: []api.Node{
					{
						ID:   "2",
						Name: "Button",
						Type: "RECTANGLE",
					},
					{
						ID:   "3",
						Name: "Label",
						Type: "TEXT",
					},
				},
			},
			{
				ID:   "4",
				Name: "FrameB",
				Type: "FRAME",
				Children: []api.Node{
					{
						ID:   "5",
						Name: "Input",
						Type: "COMPONENT",
					},
					{
						ID:   "6",
						Name: "Header",
						Type: "TEXT",
					},
				},
			},
			{
				ID:   "7",
				Name: "Footer",
				Type: "FRAME",
			},
		},
	}

	tests := []struct {
		name    string
		opts    *ListOptions
		wantIDs []string
	}{
		// No filters - returns all nodes (depth unlimited, MaxDepth -1, Recursive true)
		{
			name:    "no filters",
			opts:    &ListOptions{MaxDepth: -1, Recursive: true},
			wantIDs: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
		},
		// Type filter
		{
			name:    "filter FRAME",
			opts:    &ListOptions{TypeFilter: "FRAME", MaxDepth: -1, Recursive: true},
			wantIDs: []string{"1", "4", "7"},
		},
		{
			name:    "filter TEXT",
			opts:    &ListOptions{TypeFilter: "TEXT", MaxDepth: -1, Recursive: true},
			wantIDs: []string{"3", "6"},
		},
		// Name pattern with wildcard
		{
			name:    "name pattern Frame*",
			opts:    &ListOptions{NamePattern: "Frame*", MaxDepth: -1, Recursive: true},
			wantIDs: []string{"1", "4"},
		},
		{
			name:    "name pattern *er",
			opts:    &ListOptions{NamePattern: "*er", MaxDepth: -1, Recursive: true},
			wantIDs: []string{"6", "7"}, // Header, Footer
		},
		// Combined type and name
		{
			name:    "type FRAME and name Frame*",
			opts:    &ListOptions{TypeFilter: "FRAME", NamePattern: "Frame*", MaxDepth: -1, Recursive: true},
			wantIDs: []string{"1", "4"},
		},
		// Case-insensitive name pattern
		{
			name:    "case-insensitive pattern",
			opts:    &ListOptions{NamePattern: "frame*", CaseSensitive: false, MaxDepth: -1, Recursive: true},
			wantIDs: []string{"1", "4"},
		},
		// Case-sensitive mismatch
		{
			name:    "case-sensitive mismatch",
			opts:    &ListOptions{NamePattern: "frame*", CaseSensitive: true, MaxDepth: -1, Recursive: true},
			wantIDs: []string{},
		},
		// Recursive false (only immediate children)
		{
			name:    "non-recursive",
			opts:    &ListOptions{Recursive: false},
			wantIDs: []string{"0", "1", "4", "7"},
		},
		// Depth limit
		{
			name:    "depth 1",
			opts:    &ListOptions{MaxDepth: 1, Recursive: true},
			wantIDs: []string{"0", "1", "4", "7"},
		},
		{
			name:    "depth 2",
			opts:    &ListOptions{MaxDepth: 2, Recursive: true},
			wantIDs: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
		},
		// Invalid name pattern (should return empty)
		{
			name:    "invalid pattern",
			opts:    &ListOptions{NamePattern: "[", MaxDepth: -1, Recursive: true},
			wantIDs: []string{},
		},
		// Limit and offset are NOT applied by collectNodes; they are applied in runList.
		// So we omit those tests here.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := collectNodes(root, "testfile", tt.opts)
			var gotIDs []string
			for _, n := range nodes {
				gotIDs = append(gotIDs, n.ID)
			}
			if len(gotIDs) != len(tt.wantIDs) {
				t.Errorf("collectNodes() returned %d results, want %d", len(gotIDs), len(tt.wantIDs))
				t.Logf("got IDs: %v", gotIDs)
				t.Logf("want IDs: %v", tt.wantIDs)
				return
			}
			for i, id := range tt.wantIDs {
				if gotIDs[i] != id {
					t.Errorf("result %d: got ID %s, want %s", i, gotIDs[i], id)
				}
			}
		})
	}
}

func TestRunList(t *testing.T) {
	// Create a test server that returns the fixture
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Request: %s %s", r.Method, r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/v1/files/") && r.Method == http.MethodGet && !strings.Contains(r.URL.Path, "/nodes") {
			// Load fixture
			fixturePath := filepath.Join("../../../testdata", "fixtures", "file_response.json")
			data, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}
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

	// Create list command and set context
	cmd := listCmd
	cmd.SetContext(ctx)

	// Set flags
	cmd.Flags().Set("type", "")
	cmd.Flags().Set("name", "")
	cmd.Flags().Set("case-sensitive", "false")
	cmd.Flags().Set("limit", "0")
	cmd.Flags().Set("offset", "0")
	cmd.Flags().Set("recursive", "true")
	cmd.Flags().Set("depth", "-1")

	// Execute with a file key that matches the fixture (the fixture uses no specific key)
	// The fixture file key is not used; any key will work because server responds to any file key.
	err := runList(cmd, []string{"ABC123"})
	if err != nil {
		t.Errorf("runList failed: %v", err)
	}
}
