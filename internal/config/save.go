package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigPath returns the path to the JSON config file.
// Returns an error if home directory cannot be determined.
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "figma", "config.json"), nil
}

// Save writes the configuration to the JSON config file.
// The file is saved at ~/.config/figma/config.json.
// If the config file already exists, it will be overwritten.
// If the directory doesn't exist, it will be created.
func (cfg *Config) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveWithDefaults writes the configuration to the JSON config file,
// but only includes fields that differ from default values.
// This keeps the config file minimal.
func (cfg *Config) SaveWithDefaults() error {
	// TODO: implement diff-based saving
	return cfg.Save()
}
