package config

import (
	"time"

	"github.com/cristianoliveira/figma-cli/internal/logging"
)

const (
	// OutputFormatText is the plain text output format.
	OutputFormatText = "text"
	// OutputFormatJSON is the JSON output format.
	OutputFormatJSON = "json"
	// OutputFormatYAML is the YAML output format.
	OutputFormatYAML = "yaml"
)

// Config holds all configuration settings for the Figma CLI.
type Config struct {
	// Figma API token (required for API calls)
	Token string `json:"token" env:"FIGMA_ACCESS_TOKEN"`

	// TokenType indicates the type of token: "pat" or "oauth".
	// Defaults to "pat" for backward compatibility.
	TokenType string `json:"token_type" env:"FIGMA_TOKEN_TYPE"`

	// OAuthClientID is the client ID for OAuth authentication.
	// If empty, a default client ID may be used (not recommended for production).
	OAuthClientID string `json:"oauth_client_id" env:"FIGMA_OAUTH_CLIENT_ID"`

	// OAuthScopes is a comma-separated list of OAuth scopes to request.
	// Defaults to "file_content:read".
	OAuthScopes string `json:"oauth_scopes" env:"FIGMA_OAUTH_SCOPES"`

	// AuthToken stores serialized authentication token (JSON). Used for OAuth tokens with refresh capability.
	AuthToken string `json:"auth_token,omitempty"`

	// Output format for command results (json, yaml, text)
	OutputFormat string `json:"output_format" env:"OUTPUT_FORMAT"`

	// Export directory for downloaded assets
	ExportDir string `json:"export_dir" env:"EXPORT_DIR"`

	// API settings
	API APISettings `json:"api"`

	// Logging settings
	Logging LoggingConfig `json:"logging"`
}

// APISettings holds configuration for API client behavior.
type APISettings struct {
	// Timeout for HTTP requests
	Timeout time.Duration `json:"timeout" env:"API_TIMEOUT"`
	// Maximum number of retries for failed requests
	MaxRetries int `json:"max_retries" env:"API_MAX_RETRIES"`
	// Base URL for Figma API (can be overridden for testing)
	BaseURL string `json:"base_url" env:"API_BASE_URL"`
	// Enable debug logging for API requests
	Debug bool `json:"debug" env:"API_DEBUG"`
	// API Tier (1, 2, 3, 4) for rate limiting
	Tier int `json:"tier" env:"API_TIER"`
	// Seat type (view_collab, dev_full) for rate limiting
	SeatType string `json:"seat_type" env:"API_SEAT_TYPE"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	// Level is the minimum log level (debug, info, warn, error).
	Level string `json:"level" env:"LOG_LEVEL"`
	// Format is the log format (json, text).
	Format string `json:"format" env:"LOG_FORMAT"`
	// File is the path to the log file. If empty, logs to stderr.
	File string `json:"file" env:"LOG_FILE"`
	// MaxSize is the maximum size in megabytes of the log file before rotation.
	MaxSize int `json:"max_size" env:"LOG_MAX_SIZE"`
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `json:"max_backups" env:"LOG_MAX_BACKUPS"`
	// MaxAge is the maximum number of days to retain old log files.
	MaxAge int `json:"max_age" env:"LOG_MAX_AGE"`
	// Compress determines if rotated log files should be compressed.
	Compress bool `json:"compress" env:"LOG_COMPRESS"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Token:         "",
		TokenType:     "pat",
		OAuthClientID: "",
		OAuthScopes:   "file_content:read",
		AuthToken:     "",
		OutputFormat:  OutputFormatText,
		ExportDir:     ".",
		API: APISettings{
			Timeout:    30 * time.Second,
			MaxRetries: 3,
			BaseURL:    "https://api.figma.com/v1",
			Debug:      false,
			Tier:       1,
			SeatType:   "dev_full",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			File:       "",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		},
	}
}

// LoggingConfig converts the config's logging settings to a logging.Config.
func (cfg *Config) LoggingConfig() logging.Config {
	return logging.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		File:       cfg.Logging.File,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	}
}
