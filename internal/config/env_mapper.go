package config

import (
	"fmt"
	"time"
)

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
