package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Load loads configuration from multiple sources with precedence:
// 1. Default values
// 2. .env file in current directory (dotenv)
// 3. JSON config file at ~/.config/figma/config.json
// 4. Environment variables (override file config)
// 5. CLI flags (should be applied after Load via Merge)
//
// Returns the merged configuration and any error encountered.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Load .env file if present
	if err := loadDotEnv(&cfg); err != nil {
		return nil, fmt.Errorf("loading .env: %w", err)
	}

	// Load JSON config file if present
	if err := loadJSONConfig(&cfg); err != nil {
		return nil, fmt.Errorf("loading JSON config: %w", err)
	}

	// Apply environment variables (override file config)
	if err := applyEnv(&cfg); err != nil {
		return nil, fmt.Errorf("applying environment: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &cfg, nil
}

// loadDotEnv loads configuration from .env file in current directory.
// Values are applied directly to config, not set as environment variables.
func loadDotEnv(cfg *Config) error {
	envMap, err := godotenv.Read()
	if err != nil {
		if os.IsNotExist(err) {
			return nil // .env file not found, skip
		}
		return err
	}
	// Apply .env values (lowest priority)
	applyEnvMap(cfg, envMap)
	return nil
}

// loadJSONConfig loads configuration from ~/.config/figma/config.json.
func loadJSONConfig(cfg *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home directory, skip JSON config
		return nil
	}

	configPath := filepath.Join(homeDir, ".config", "figma", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, skip
		}
		return err
	}

	// Parse JSON into a temporary config to avoid overriding defaults unintentionally
	var fileConfig Config
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}

	// Merge file config into current config, respecting non-zero values
	mergeConfig(cfg, &fileConfig)
	return nil
}

// applyEnv applies environment variables to config.
func applyEnv(cfg *Config) error {
	// Build map from environment
	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		key, val, ok := strings.Cut(env, "=")
		if ok {
			envMap[key] = val
		}
	}
	applyEnvMap(cfg, envMap)
	return nil
}
