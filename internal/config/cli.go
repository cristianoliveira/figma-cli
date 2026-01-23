package config

import "time"

// CLIFlags represents command-line flag values that can override configuration.
type CLIFlags struct {
	Token        string
	OutputFormat string
	ExportDir    string

	APITimeout    time.Duration
	APIMaxRetries int
	APIBaseURL    string
	APIDebug      bool
}
