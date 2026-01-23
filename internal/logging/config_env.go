package logging

import (
	"os"
	"strconv"
	"strings"
)

const (
	envLogLevel      = "LOG_LEVEL"
	envLogFormat     = "LOG_FORMAT"
	envLogFile       = "LOG_FILE"
	envLogMaxSize    = "LOG_MAX_SIZE"
	envLogMaxBackups = "LOG_MAX_BACKUPS"
	envLogMaxAge     = "LOG_MAX_AGE"
	envLogCompress   = "LOG_COMPRESS"
)

// ApplyEnv applies environment variables to the configuration.
func ApplyEnv(cfg Config) Config {
	if level := os.Getenv(envLogLevel); level != "" {
		cfg.Level = strings.ToLower(level)
	}
	if format := os.Getenv(envLogFormat); format != "" {
		cfg.Format = strings.ToLower(format)
	}
	if file := os.Getenv(envLogFile); file != "" {
		cfg.File = file
	}
	if maxSize := os.Getenv(envLogMaxSize); maxSize != "" {
		if val, err := strconv.Atoi(maxSize); err == nil && val > 0 {
			cfg.MaxSize = val
		}
	}
	if maxBackups := os.Getenv(envLogMaxBackups); maxBackups != "" {
		if val, err := strconv.Atoi(maxBackups); err == nil && val >= 0 {
			cfg.MaxBackups = val
		}
	}
	if maxAge := os.Getenv(envLogMaxAge); maxAge != "" {
		if val, err := strconv.Atoi(maxAge); err == nil && val >= 0 {
			cfg.MaxAge = val
		}
	}
	if compress := os.Getenv(envLogCompress); compress != "" {
		cfg.Compress = strings.ToLower(compress) == "true" || compress == "1"
	}
	return cfg
}
