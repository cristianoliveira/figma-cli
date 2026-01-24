package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/joho/godotenv"
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
		AuthToken:      "",
		OutputFormat:  "text",
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

// Merge applies CLI flag values to the configuration.
// This should be called after Load with values parsed from CLI flags.
func (cfg *Config) Merge(flags *CLIFlags) {
	if flags == nil {
		return
	}
	if flags.Token != "" {
		cfg.Token = flags.Token
	}
	if flags.OutputFormat != "" {
		cfg.OutputFormat = flags.OutputFormat
	}
	if flags.ExportDir != "" {
		cfg.ExportDir = flags.ExportDir
	}
	if flags.APITimeout > 0 {
		cfg.API.Timeout = flags.APITimeout
	}
	if flags.APIMaxRetries >= 0 {
		cfg.API.MaxRetries = flags.APIMaxRetries
	}
	if flags.APIBaseURL != "" {
		cfg.API.BaseURL = flags.APIBaseURL
	}
	if flags.APIDebug {
		cfg.API.Debug = true
	}
	if flags.APITier > 0 {
		cfg.API.Tier = flags.APITier
	}
	if flags.APISeatType != "" {
		cfg.API.SeatType = flags.APISeatType
	}
}

// Validate checks that configuration values are valid (format, ranges, etc.).
// Does not check for presence of required fields like token.
func (cfg *Config) Validate() error {
	if cfg.OutputFormat != "" && cfg.OutputFormat != "json" && cfg.OutputFormat != "yaml" && cfg.OutputFormat != "text" {
		return fmt.Errorf("invalid output format %q, must be one of: json, yaml, text", cfg.OutputFormat)
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

// mergeConfig merges non-zero values from src into dst.
func mergeConfig(dst, src *Config) {
	if src.Token != "" {
		dst.Token = src.Token
	}
	if src.TokenType != "" {
		dst.TokenType = src.TokenType
	}
	if src.OAuthClientID != "" {
		dst.OAuthClientID = src.OAuthClientID
	}
	if src.OAuthScopes != "" {
		dst.OAuthScopes = src.OAuthScopes
	}
	if src.OutputFormat != "" {
		dst.OutputFormat = src.OutputFormat
	}
	if src.ExportDir != "" {
		dst.ExportDir = src.ExportDir
	}
	if src.API.Timeout > 0 {
		dst.API.Timeout = src.API.Timeout
	}
	if src.API.MaxRetries > 0 {
		dst.API.MaxRetries = src.API.MaxRetries
	}
	if src.API.BaseURL != "" {
		dst.API.BaseURL = src.API.BaseURL
	}
	if src.API.Debug {
		dst.API.Debug = true
	}
	if src.API.Tier > 0 {
		dst.API.Tier = src.API.Tier
	}
	if src.API.SeatType != "" {
		dst.API.SeatType = src.API.SeatType
	}
	// Merge logging config
	mergeLoggingConfig(&dst.Logging, &src.Logging)
}

// mergeLoggingConfig merges non-zero values from src into dst.
func mergeLoggingConfig(dst, src *LoggingConfig) {
	if src.Level != "" {
		dst.Level = src.Level
	}
	if src.Format != "" {
		dst.Format = src.Format
	}
	if src.File != "" {
		dst.File = src.File
	}
	if src.MaxSize > 0 {
		dst.MaxSize = src.MaxSize
	}
	if src.MaxBackups > 0 {
		dst.MaxBackups = src.MaxBackups
	}
	if src.MaxAge > 0 {
		dst.MaxAge = src.MaxAge
	}
	if src.Compress {
		dst.Compress = true
	}
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

// applyEnvMap applies configuration from a map of key-value strings.
func applyEnvMap(cfg *Config, envMap map[string]string) {
	// Token
	if val := envMap["FIGMA_ACCESS_TOKEN"]; val != "" {
		cfg.Token = val
	}
	// Token type
	if val := envMap["FIGMA_TOKEN_TYPE"]; val != "" {
		cfg.TokenType = val
	}
	// OAuth client ID
	if val := envMap["FIGMA_OAUTH_CLIENT_ID"]; val != "" {
		cfg.OAuthClientID = val
	}
	// OAuth scopes
	if val := envMap["FIGMA_OAUTH_SCOPES"]; val != "" {
		cfg.OAuthScopes = val
	}
	// Output format
	if val := envMap["OUTPUT_FORMAT"]; val != "" {
		cfg.OutputFormat = val
	}
	// Export directory
	if val := envMap["EXPORT_DIR"]; val != "" {
		cfg.ExportDir = val
	}
	// API timeout
	if val := envMap["API_TIMEOUT"]; val != "" {
		if duration, err := time.ParseDuration(val); err == nil && duration > 0 {
			cfg.API.Timeout = duration
		}
	}
	// API max retries
	if val := envMap["API_MAX_RETRIES"]; val != "" {
		if retries, err := parseInt(val); err == nil && retries >= 0 {
			cfg.API.MaxRetries = retries
		}
	}
	// API base URL
	if val := envMap["API_BASE_URL"]; val != "" {
		cfg.API.BaseURL = val
	}
	// API debug
	if val := envMap["API_DEBUG"]; val != "" {
		cfg.API.Debug = val == "true" || val == "1"
	}
	// API tier
	if val := envMap["API_TIER"]; val != "" {
		if tier, err := parseInt(val); err == nil && tier > 0 {
			cfg.API.Tier = tier
		}
	}
	// API seat type
	if val := envMap["API_SEAT_TYPE"]; val != "" {
		cfg.API.SeatType = val
	}
	// Logging level
	if val := envMap["LOG_LEVEL"]; val != "" {
		cfg.Logging.Level = val
	}
	// Logging format
	if val := envMap["LOG_FORMAT"]; val != "" {
		cfg.Logging.Format = val
	}
	// Logging file
	if val := envMap["LOG_FILE"]; val != "" {
		cfg.Logging.File = val
	}
	// Logging max size
	if val := envMap["LOG_MAX_SIZE"]; val != "" {
		if size, err := parseInt(val); err == nil && size > 0 {
			cfg.Logging.MaxSize = size
		}
	}
	// Logging max backups
	if val := envMap["LOG_MAX_BACKUPS"]; val != "" {
		if backups, err := parseInt(val); err == nil && backups >= 0 {
			cfg.Logging.MaxBackups = backups
		}
	}
	// Logging max age
	if val := envMap["LOG_MAX_AGE"]; val != "" {
		if age, err := parseInt(val); err == nil && age >= 0 {
			cfg.Logging.MaxAge = age
		}
	}
	// Logging compress
	if val := envMap["LOG_COMPRESS"]; val != "" {
		cfg.Logging.Compress = val == "true" || val == "1"
	}
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
