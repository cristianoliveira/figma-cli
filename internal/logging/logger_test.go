package logging

import (
	"context"
	"os"
	"testing"
)

func TestNewLogger(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Level = "debug"
	cfg.Format = "text"

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	ctx := context.Background()
	logger.Debug(ctx, "debug message")
	logger.Info(ctx, "info message")
	logger.Warn(ctx, "warn message")
	logger.Error(ctx, "error message")
}

func TestLoggerWithRequestID(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Level = "debug"
	cfg.Format = "json"

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	ctx := NewContextWithRequestID(context.Background(), "test-request-123")
	logger.Info(ctx, "message with request id")
}

func TestLoggerWithFields(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Level = "info"
	cfg.Format = "text"

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	ctx := context.Background()
	logger.Info(ctx, "message with fields",
		String("key", "value"),
		Int("count", 42),
		Err(os.ErrNotExist),
	)
}

func TestApplyEnv(t *testing.T) {
	_ = os.Setenv("LOG_LEVEL", "debug")
	_ = os.Setenv("LOG_FORMAT", "json")
	_ = os.Setenv("LOG_FILE", "/tmp/test.log")
	defer func() {
		_ = os.Unsetenv("LOG_LEVEL")
		_ = os.Unsetenv("LOG_FORMAT")
		_ = os.Unsetenv("LOG_FILE")
	}()

	cfg := DefaultConfig()
	cfg = ApplyEnv(cfg)

	if cfg.Level != "debug" {
		t.Errorf("expected level debug, got %s", cfg.Level)
	}
	if cfg.Format != "json" {
		t.Errorf("expected format json, got %s", cfg.Format)
	}
	if cfg.File != "/tmp/test.log" {
		t.Errorf("expected file /tmp/test.log, got %s", cfg.File)
	}
}

func TestDefaultLogger(t *testing.T) {
	// Ensure default logger can be created
	logger := Default()
	if logger == nil {
		t.Fatal("default logger is nil")
	}

	ctx := context.Background()
	logger.Info(ctx, "test default logger")
}
