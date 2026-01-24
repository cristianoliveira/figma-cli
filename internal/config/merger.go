package config

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
