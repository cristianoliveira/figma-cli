package cmd

import (
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/spf13/cobra"
)

const debugLevel = "debug"
const jsonFormat = "json"

func TestAdjustConfig(t *testing.T) {
	t.Run("debug flag true", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("debug", false, "")
		cmd.Flags().String("log-level", "", "")
		cfg := &config.Config{}
		_ = cmd.Flags().Set("debug", "true")
		if err := adjustConfig(cfg, cmd); err != nil {
			t.Errorf("adjustConfig with debug flag returned error: %v", err)
		}
		if cfg.Logging.Level != debugLevel {
			t.Errorf("expected Logging.Level = %s, got %s", debugLevel, cfg.Logging.Level)
		}
		if !cfg.API.Debug {
			t.Errorf("expected API.Debug = true, got %v", cfg.API.Debug)
		}
	})

	t.Run("log-level flag", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("debug", false, "")
		cmd.Flags().String("log-level", "", "")
		cfg := &config.Config{}
		_ = cmd.Flags().Set("log-level", "info")
		if err := adjustConfig(cfg, cmd); err != nil {
			t.Errorf("adjustConfig with log-level flag returned error: %v", err)
		}
		if cfg.Logging.Level != "info" {
			t.Errorf("expected Logging.Level = info, got %s", cfg.Logging.Level)
		}
	})

	t.Run("missing debug flag returns error", func(t *testing.T) {
		cmd := &cobra.Command{}
		// Do not define debug flag
		cfg := &config.Config{}
		err := adjustConfig(cfg, cmd)
		if err == nil {
			t.Error("expected error when debug flag not defined, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to get debug flag") {
			t.Errorf("expected error about debug flag, got: %v", err)
		}
	})

	t.Run("missing log-level flag returns error", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("debug", false, "") // define debug but not log-level
		cfg := &config.Config{}
		err := adjustConfig(cfg, cmd)
		if err == nil {
			t.Error("expected error when log-level flag not defined, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to get log-level flag") {
			t.Errorf("expected error about log-level flag, got: %v", err)
		}
	})
}

func TestGetCLIFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("token", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("export-dir", "", "")
	cmd.Flags().Duration("api-timeout", 0, "")
	cmd.Flags().Int("api-max-retries", 0, "")
	cmd.Flags().String("api-base-url", "", "")
	cmd.Flags().Bool("api-debug", false, "")
	cmd.Flags().Bool("debug", false, "")
	cmd.Flags().String("log-level", "", "")
	cmd.Flags().String("log-format", "", "")
	cmd.Flags().String("log-file", "", "")
	cmd.Flags().Int("log-max-size", 0, "")
	cmd.Flags().Int("log-max-backups", 0, "")
	cmd.Flags().Int("log-max-age", 0, "")
	cmd.Flags().Bool("log-compress", false, "")

	// Set some values
	_ = cmd.Flags().Set("token", "test-token")
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("export-dir", "/tmp")
	_ = cmd.Flags().Set("api-timeout", "30s")
	_ = cmd.Flags().Set("api-max-retries", "3")
	_ = cmd.Flags().Set("api-base-url", "https://test.com")
	_ = cmd.Flags().Set("api-debug", "true")
	_ = cmd.Flags().Set("debug", "true")
	_ = cmd.Flags().Set("log-level", "debug")
	_ = cmd.Flags().Set("log-format", "json")
	_ = cmd.Flags().Set("log-file", "/tmp/log")
	_ = cmd.Flags().Set("log-max-size", "100")
	_ = cmd.Flags().Set("log-max-backups", "5")
	_ = cmd.Flags().Set("log-max-age", "30")
	_ = cmd.Flags().Set("log-compress", "true")

	flags, err := getCLIFlags(cmd)
	if err != nil {
		t.Fatalf("getCLIFlags returned error: %v", err)
	}
	if flags.Token != "test-token" {
		t.Errorf("expected Token = test-token, got %s", flags.Token)
	}
	if flags.OutputFormat != "json" {
		t.Errorf("expected OutputFormat = json, got %s", flags.OutputFormat)
	}
	if flags.ExportDir != "/tmp" {
		t.Errorf("expected ExportDir = /tmp, got %s", flags.ExportDir)
	}
	if flags.APIDebug != true {
		t.Errorf("expected APIDebug = true, got %v", flags.APIDebug)
	}
	if flags.LogLevel != debugLevel {
		t.Errorf("expected LogLevel = %s, got %s", debugLevel, flags.LogLevel)
	}
	if flags.LogFormat != jsonFormat {
		t.Errorf("expected LogFormat = %s, got %s", jsonFormat, flags.LogFormat)
	}
	if flags.LogFile != "/tmp/log" {
		t.Errorf("expected LogFile = /tmp/log, got %s", flags.LogFile)
	}
	if flags.LogMaxSize != 100 {
		t.Errorf("expected LogMaxSize = 100, got %d", flags.LogMaxSize)
	}
	if flags.LogMaxBackups != 5 {
		t.Errorf("expected LogMaxBackups = 5, got %d", flags.LogMaxBackups)
	}
	if flags.LogMaxAge != 30 {
		t.Errorf("expected LogMaxAge = 30, got %d", flags.LogMaxAge)
	}
	if flags.LogCompress != true {
		t.Errorf("expected LogCompress = true, got %v", flags.LogCompress)
	}
}
