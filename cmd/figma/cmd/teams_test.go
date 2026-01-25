package cmd

import (
	"context"
	"encoding/json"
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
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func TestRunTeamsProjects_Success(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "team_projects_response.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	var expectedProjects []*api.Project
	if err := json.Unmarshal(data, &expectedProjects); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		// Expect path /v1/teams/team123/projects (base URL includes /v1)
		if !strings.HasPrefix(r.URL.Path, "/v1/teams/") {
			t.Errorf("expected path starting with /v1/teams/, got %s", r.URL.Path)
		}
		// Check authorization header (PAT token)
		if r.Header.Get("X-Figma-Token") != "test-token" {
			t.Errorf("expected X-Figma-Token header 'test-token', got %s", r.Header.Get("X-Figma-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	// Create config with test server base URL (must include /v1 because API client expects it)
	cfg := &config.Config{
		Token:     "test-token",
		TokenType: "pat",
		API: config.APISettings{
			BaseURL:    ts.URL + "/v1",
			Timeout:    30 * time.Second,
			MaxRetries: 0,
			Debug:      false,
			Tier:       1,
			SeatType:   "dev_full",
		},
		OutputFormat: config.OutputFormatText,
		ExportDir:    ".",
	}

	// Create logger
	logger := logging.NewNopLogger()

	// Create a mock cobra command with context containing config and logger
	cmd := &cobra.Command{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, configKey{}, cfg)
	ctx = context.WithValue(ctx, loggerKey{}, logger)
	cmd.SetContext(ctx)

	// Capture output
	outputBuf := &strings.Builder{}
	cmd.SetOut(outputBuf)
	cmd.SetErr(outputBuf)

	// Temporarily replace newAPIClient to use our config (it will use the base URL)
	// The actual newAPIClient will use the cfg.API.BaseURL which points to our test server.
	// This is fine; we just need to call runTeamsProjects with the team ID.
	// We'll call the function directly with our command and args.
	err = runTeamsProjects(cmd, []string{"team123"})
	if err != nil {
		t.Fatalf("runTeamsProjects failed: %v", err)
	}

	output := outputBuf.String()
	// Verify output contains expected project names
	for _, p := range expectedProjects {
		if !strings.Contains(output, p.Name) {
			t.Errorf("output missing project name %q", p.Name)
		}
		if !strings.Contains(output, p.ID) {
			t.Errorf("output missing project ID %q", p.ID)
		}
	}
}

func TestRunTeamsProjects_ErrorHandling(t *testing.T) {
	// Test server returning error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer ts.Close()

	cfg := &config.Config{
		Token:     "invalid-token",
		TokenType: "pat",
		API: config.APISettings{
			BaseURL:    ts.URL + "/v1",
			Timeout:    30 * time.Second,
			MaxRetries: 0,
			Debug:      false,
			Tier:       1,
			SeatType:   "dev_full",
		},
		OutputFormat: config.OutputFormatText,
		ExportDir:    ".",
	}

	logger := logging.NewNopLogger()
	cmd := &cobra.Command{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, configKey{}, cfg)
	ctx = context.WithValue(ctx, loggerKey{}, logger)
	cmd.SetContext(ctx)

	outputBuf := &strings.Builder{}
	cmd.SetOut(outputBuf)
	cmd.SetErr(outputBuf)

	err := runTeamsProjects(cmd, []string{"team123"})
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if !strings.Contains(err.Error(), "failed to fetch team projects") {
		t.Errorf("expected error about fetching projects, got: %v", err)
	}
}

func TestProjectOutputConversion(t *testing.T) {
	// Test conversion from api.Project to ProjectOutput
	apiProject := &api.Project{
		ID:         "proj_123",
		Name:       "Test Project",
		CreatedAt:  "2024-01-01T00:00:00Z",
		ModifiedAt: "2024-01-02T00:00:00Z",
	}
	output := ProjectOutput{
		ID:         apiProject.ID,
		Name:       apiProject.Name,
		CreatedAt:  apiProject.CreatedAt,
		ModifiedAt: apiProject.ModifiedAt,
	}
	if output.ID != apiProject.ID {
		t.Errorf("expected ID %s, got %s", apiProject.ID, output.ID)
	}
	if output.Name != apiProject.Name {
		t.Errorf("expected Name %s, got %s", apiProject.Name, output.Name)
	}
	if output.CreatedAt != apiProject.CreatedAt {
		t.Errorf("expected CreatedAt %s, got %s", apiProject.CreatedAt, output.CreatedAt)
	}
	if output.ModifiedAt != apiProject.ModifiedAt {
		t.Errorf("expected ModifiedAt %s, got %s", apiProject.ModifiedAt, output.ModifiedAt)
	}
}

func TestRunTeamsProjects_OutputFormats(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "team_projects_response.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	var expectedProjects []*api.Project
	if err := json.Unmarshal(data, &expectedProjects); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	const (
		formatJSON = "json"
		formatYAML = "yaml"
		formatText = "text"
	)
	outputFormats := []string{formatJSON, formatYAML, formatText}
	for _, format := range outputFormats {
		t.Run(format, func(t *testing.T) {
			cfg := &config.Config{
				Token:     "test-token",
				TokenType: "pat",
				API: config.APISettings{
					BaseURL:    ts.URL + "/v1",
					Timeout:    30 * time.Second,
					MaxRetries: 0,
					Debug:      false,
					Tier:       1,
					SeatType:   "dev_full",
				},
				OutputFormat: format,
				ExportDir:    ".",
			}

			logger := logging.NewNopLogger()
			cmd := &cobra.Command{}
			ctx := context.Background()
			ctx = context.WithValue(ctx, configKey{}, cfg)
			ctx = context.WithValue(ctx, loggerKey{}, logger)
			cmd.SetContext(ctx)

			outputBuf := &strings.Builder{}
			cmd.SetOut(outputBuf)
			cmd.SetErr(outputBuf)

			err := runTeamsProjects(cmd, []string{"team123"})
			if err != nil {
				t.Fatalf("runTeamsProjects failed with format %s: %v", format, err)
			}

			output := outputBuf.String()
			// Basic validation per format
			switch format {
			case formatJSON:
				var projects []ProjectOutput
				if err := json.Unmarshal([]byte(output), &projects); err != nil {
					t.Errorf("failed to unmarshal JSON output: %v", err)
				}
				if len(projects) != len(expectedProjects) {
					t.Errorf("expected %d projects, got %d", len(expectedProjects), len(projects))
				}
			case formatYAML:
				var projects []ProjectOutput
				if err := yaml.Unmarshal([]byte(output), &projects); err != nil {
					t.Errorf("failed to unmarshal YAML output: %v", err)
				}
				if len(projects) != len(expectedProjects) {
					t.Errorf("expected %d projects, got %d", len(expectedProjects), len(projects))
				}
			case formatText:
				// Ensure each project name appears
				for _, p := range expectedProjects {
					if !strings.Contains(output, p.Name) {
						t.Errorf("text output missing project name %q", p.Name)
					}
				}
			}
		})
	}
}
