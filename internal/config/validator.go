package config

import "fmt"

// Validate checks that configuration values are valid (format, ranges, etc.).
// Does not check for presence of required fields like token.
func (cfg *Config) Validate() error {
	if cfg.OutputFormat != "" && cfg.OutputFormat != OutputFormatJSON && cfg.OutputFormat != OutputFormatYAML && cfg.OutputFormat != OutputFormatText {
		return fmt.Errorf("invalid output format %q, must be one of: %s, %s, %s", cfg.OutputFormat, OutputFormatJSON, OutputFormatYAML, OutputFormatText)
	}
	if cfg.ExportDir == "" {
		return fmt.Errorf("export directory cannot be empty")
	}
	if cfg.TokenType != "pat" && cfg.TokenType != "oauth" {
		return fmt.Errorf("invalid token type %q, must be either 'pat' or 'oauth'", cfg.TokenType)
	}
	if cfg.API.Timeout <= 0 {
		return fmt.Errorf("API timeout must be positive")
	}
	if cfg.API.MaxRetries < 0 {
		return fmt.Errorf("API max retries cannot be negative")
	}
	if cfg.API.Tier < 1 || cfg.API.Tier > 4 {
		return fmt.Errorf("API tier must be between 1 and 4")
	}
	if cfg.API.SeatType != "" && cfg.API.SeatType != "view_collab" && cfg.API.SeatType != "dev_full" {
		return fmt.Errorf("API seat type must be either 'view_collab' or 'dev_full'")
	}
	// Validate logging config
	if cfg.Logging.Level != "" && cfg.Logging.Level != "debug" && cfg.Logging.Level != "info" && cfg.Logging.Level != "warn" && cfg.Logging.Level != "error" {
		return fmt.Errorf("invalid log level %q, must be one of: debug, info, warn, error", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "" && cfg.Logging.Format != "json" && cfg.Logging.Format != "text" {
		return fmt.Errorf("invalid log format %q, must be one of: json, text", cfg.Logging.Format)
	}
	if cfg.Logging.MaxSize < 0 {
		return fmt.Errorf("log max size cannot be negative")
	}
	if cfg.Logging.MaxBackups < 0 {
		return fmt.Errorf("log max backups cannot be negative")
	}
	if cfg.Logging.MaxAge < 0 {
		return fmt.Errorf("log max age cannot be negative")
	}
	return nil
}

// ValidateRequired checks that required fields for API operations are present.
func (cfg *Config) ValidateRequired() error {
	if cfg.Token == "" {
		return fmt.Errorf("token is required (set FIGMA_ACCESS_TOKEN environment variable or use --token flag)")
	}
	return cfg.Validate()
}
