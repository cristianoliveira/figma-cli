// Package api provides a client for the Figma API.
//
//go:generate go run ./gen -input client.go -output logging_client_gen.go -template gen/templates/logging.tmpl
package api

import (
	"github.com/cristianoliveira/figma-cli/internal/logging"
)

// LoggingClient wraps a Client and logs all requests and responses.
type LoggingClient struct {
	client Client
	logger logging.Logger
}

// NewLoggingClient creates a new LoggingClient.
func NewLoggingClient(client Client, logger logging.Logger) *LoggingClient {
	return &LoggingClient{
		client: client,
		logger: logger,
	}
}
