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

func TestRunProjectsFiles_Success(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "project_files_response.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	var expectedFiles []*api.File
	if err := json.Unmarshal(data, &expectedFiles); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		// Expect path /v1/projects/proj123/files (base URL includes /v1)
		if !strings.HasPrefix(r.URL.Path, "/v1/projects/") {
			t.Errorf("expected path starting with /v1/projects/, got %s", r.URL.Path)
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

	// Call runProjectsFiles directly with project ID
	err = runProjectsFiles(cmd, []string{"proj123"})
	if err != nil {
		t.Fatalf("runProjectsFiles failed: %v", err)
	}

	output := outputBuf.String()
	// Verify output contains expected file names and keys
	for _, f := range expectedFiles {
		if !strings.Contains(output, f.Name) {
			t.Errorf("output missing file name %q", f.Name)
		}
		if !strings.Contains(output, f.Key) {
			t.Errorf("output missing file key %q", f.Key)
		}
	}
}

func TestRunProjectsFiles_ErrorHandling(t *testing.T) {
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

	err := runProjectsFiles(cmd, []string{"proj123"})
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if !strings.Contains(err.Error(), "failed to fetch project files") {
		t.Errorf("expected error about fetching project files, got: %v", err)
	}
}

func TestFileOutputConversion(t *testing.T) {
	// Test conversion from api.File to FileOutput
	apiFile := &api.File{
		Key:          "file123",
		Name:         "Test File",
		LastModified: "2024-01-01T00:00:00Z",
		Version:      "1234567890",
		ThumbnailURL: "https://example.com/thumb.jpg",
	}
	output := FileOutput{
		Key:          apiFile.Key,
		Name:         apiFile.Name,
		LastModified: apiFile.LastModified,
		Version:      apiFile.Version,
		ThumbnailURL: apiFile.ThumbnailURL,
	}
	if output.Key != apiFile.Key {
		t.Errorf("expected Key %s, got %s", apiFile.Key, output.Key)
	}
	if output.Name != apiFile.Name {
		t.Errorf("expected Name %s, got %s", apiFile.Name, output.Name)
	}
	if output.LastModified != apiFile.LastModified {
		t.Errorf("expected LastModified %s, got %s", apiFile.LastModified, output.LastModified)
	}
	if output.Version != apiFile.Version {
		t.Errorf("expected Version %s, got %s", apiFile.Version, output.Version)
	}
	if output.ThumbnailURL != apiFile.ThumbnailURL {
		t.Errorf("expected ThumbnailURL %s, got %s", apiFile.ThumbnailURL, output.ThumbnailURL)
	}
}

func TestRunProjectsFiles_OutputFormats(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "project_files_response.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	var expectedFiles []*api.File
	if err := json.Unmarshal(data, &expectedFiles); err != nil {
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

			err := runProjectsFiles(cmd, []string{"proj123"})
			if err != nil {
				t.Fatalf("runProjectsFiles failed with format %s: %v", format, err)
			}

			output := outputBuf.String()
			// Basic validation per format
			switch format {
			case formatJSON:
				var files []FileOutput
				if err := json.Unmarshal([]byte(output), &files); err != nil {
					t.Errorf("failed to unmarshal JSON output: %v", err)
				}
				if len(files) != len(expectedFiles) {
					t.Errorf("expected %d files, got %d", len(expectedFiles), len(files))
				}
			case formatYAML:
				var files []FileOutput
				if err := yaml.Unmarshal([]byte(output), &files); err != nil {
					t.Errorf("failed to unmarshal YAML output: %v", err)
				}
				if len(files) != len(expectedFiles) {
					t.Errorf("expected %d files, got %d", len(expectedFiles), len(files))
				}
			case formatText:
				// Ensure each file name appears
				for _, f := range expectedFiles {
					if !strings.Contains(output, f.Name) {
						t.Errorf("text output missing file name %q", f.Name)
					}
				}
			}
		})
	}
}

func TestRunProjectsFiles_NotFound(t *testing.T) {
	// Test server returning 404
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Project not found"}`))
	}))
	defer ts.Close()

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

	logger := logging.NewNopLogger()
	cmd := &cobra.Command{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, configKey{}, cfg)
	ctx = context.WithValue(ctx, loggerKey{}, logger)
	cmd.SetContext(ctx)

	outputBuf := &strings.Builder{}
	cmd.SetOut(outputBuf)
	cmd.SetErr(outputBuf)

	err := runProjectsFiles(cmd, []string{"invalid-project"})
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if !strings.Contains(err.Error(), "failed to fetch project files") {
		t.Errorf("expected error about fetching project files, got: %v", err)
	}
}
