package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Token != "" {
		t.Errorf("expected empty token, got %q", cfg.Token)
	}
	if cfg.OutputFormat != "text" {
		t.Errorf("expected output format text, got %q", cfg.OutputFormat)
	}
	if cfg.ExportDir != "." {
		t.Errorf("expected export dir ., got %q", cfg.ExportDir)
	}
	if cfg.API.Timeout != 30*time.Second {
		t.Errorf("expected API timeout 30s, got %v", cfg.API.Timeout)
	}
	if cfg.API.MaxRetries != 3 {
		t.Errorf("expected API max retries 3, got %d", cfg.API.MaxRetries)
	}
	if cfg.API.BaseURL != "https://api.figma.com/v1" {
		t.Errorf("expected API base URL https://api.figma.com/v1, got %q", cfg.API.BaseURL)
	}
	if cfg.API.Debug {
		t.Error("expected API debug false")
	}
	// Logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("expected log level info, got %q", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("expected log format text, got %q", cfg.Logging.Format)
	}
	if cfg.Logging.File != "" {
		t.Errorf("expected log file empty, got %q", cfg.Logging.File)
	}
	if cfg.Logging.MaxSize != 10 {
		t.Errorf("expected log max size 10, got %d", cfg.Logging.MaxSize)
	}
	if cfg.Logging.MaxBackups != 5 {
		t.Errorf("expected log max backups 5, got %d", cfg.Logging.MaxBackups)
	}
	if cfg.Logging.MaxAge != 30 {
		t.Errorf("expected log max age 30, got %d", cfg.Logging.MaxAge)
	}
	if !cfg.Logging.Compress {
		t.Error("expected log compress true")
	}
}

func TestValidate(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Token = "test" // set token to pass required validation
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid default config should pass validation: %v", err)
	}

	// Invalid output format
	cfg2 := cfg
	cfg2.OutputFormat = "invalid"
	if err := cfg2.Validate(); err == nil {
		t.Error("expected validation error for invalid output format")
	}

	// Empty export dir
	cfg3 := cfg
	cfg3.ExportDir = ""
	if err := cfg3.Validate(); err == nil {
		t.Error("expected validation error for empty export dir")
	}

	// Invalid API timeout
	cfg4 := cfg
	cfg4.API.Timeout = -1 * time.Second
	if err := cfg4.Validate(); err == nil {
		t.Error("expected validation error for negative API timeout")
	}

	// Invalid API max retries
	cfg5 := cfg
	cfg5.API.MaxRetries = -1
	if err := cfg5.Validate(); err == nil {
		t.Error("expected validation error for negative API max retries")
	}

	// Invalid log level
	cfg6 := cfg
	cfg6.Logging.Level = "invalid"
	if err := cfg6.Validate(); err == nil {
		t.Error("expected validation error for invalid log level")
	}

	// Invalid log format
	cfg7 := cfg
	cfg7.Logging.Format = "invalid"
	if err := cfg7.Validate(); err == nil {
		t.Error("expected validation error for invalid log format")
	}
}

func TestLoadNoFiles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// Ensure no .env file exists in test directory
	_ = os.Unsetenv("FIGMA_ACCESS_TOKEN")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed with no config files: %v", err)
	}
	if cfg.Token != "" {
		t.Errorf("expected empty token, got %q", cfg.Token)
	}
}

func TestLoadEnvFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// Create temporary .env file
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `FIGMA_ACCESS_TOKEN=env_token
OUTPUT_FORMAT=json
EXPORT_DIR=/tmp/export
API_TIMEOUT=60s
API_MAX_RETRIES=5
API_BASE_URL=https://test.figma.com
API_DEBUG=true
LOG_LEVEL=debug
LOG_FORMAT=json
LOG_FILE=/tmp/figma.log
LOG_MAX_SIZE=20
LOG_MAX_BACKUPS=10
LOG_MAX_AGE=60
LOG_COMPRESS=false`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Change working directory to temp dir
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Token != "env_token" {
		t.Errorf("expected token env_token, got %q", cfg.Token)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected output format json, got %q", cfg.OutputFormat)
	}
	if cfg.ExportDir != "/tmp/export" {
		t.Errorf("expected export dir /tmp/export, got %q", cfg.ExportDir)
	}
	if cfg.API.Timeout != 60*time.Second {
		t.Errorf("expected API timeout 60s, got %v", cfg.API.Timeout)
	}
	if cfg.API.MaxRetries != 5 {
		t.Errorf("expected API max retries 5, got %d", cfg.API.MaxRetries)
	}
	if cfg.API.BaseURL != "https://test.figma.com" {
		t.Errorf("expected API base URL https://test.figma.com, got %q", cfg.API.BaseURL)
	}
	if !cfg.API.Debug {
		t.Error("expected API debug true")
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected log level debug, got %q", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected log format json, got %q", cfg.Logging.Format)
	}
	if cfg.Logging.File != "/tmp/figma.log" {
		t.Errorf("expected log file /tmp/figma.log, got %q", cfg.Logging.File)
	}
	if cfg.Logging.MaxSize != 20 {
		t.Errorf("expected log max size 20, got %d", cfg.Logging.MaxSize)
	}
	if cfg.Logging.MaxBackups != 10 {
		t.Errorf("expected log max backups 10, got %d", cfg.Logging.MaxBackups)
	}
	if cfg.Logging.MaxAge != 60 {
		t.Errorf("expected log max age 60, got %d", cfg.Logging.MaxAge)
	}
	if cfg.Logging.Compress {
		t.Error("expected log compress false")
	}
}

func TestMergeCLIFlags(t *testing.T) {
	cfg := DefaultConfig()
	flags := &CLIFlags{
		Token:         "flag_token",
		OutputFormat:  "yaml",
		ExportDir:     "/flag/export",
		APITimeout:    10 * time.Second,
		APIMaxRetries: 1,
		APIBaseURL:    "https://flag.figma.com",
		APIDebug:      true,
	}
	cfg.Merge(flags)
	if cfg.Token != "flag_token" {
		t.Errorf("expected token flag_token, got %q", cfg.Token)
	}
	if cfg.OutputFormat != "yaml" {
		t.Errorf("expected output format yaml, got %q", cfg.OutputFormat)
	}
	if cfg.ExportDir != "/flag/export" {
		t.Errorf("expected export dir /flag/export, got %q", cfg.ExportDir)
	}
	if cfg.API.Timeout != 10*time.Second {
		t.Errorf("expected API timeout 10s, got %v", cfg.API.Timeout)
	}
	if cfg.API.MaxRetries != 1 {
		t.Errorf("expected API max retries 1, got %d", cfg.API.MaxRetries)
	}
	if cfg.API.BaseURL != "https://flag.figma.com" {
		t.Errorf("expected API base URL https://flag.figma.com, got %q", cfg.API.BaseURL)
	}
	if !cfg.API.Debug {
		t.Error("expected API debug true")
	}
}
