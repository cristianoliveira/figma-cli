// Package env provides environment variable utilities for the Figma CLI.
package env

import (
	"os"

	"github.com/cristianoliveira/figma-cli/internal/operr"
)

// GetFigmaToken returns the FIGMA_ACCESS_TOKEN environment variable.
// If the variable is not set or is empty, it returns a neutral
// authentication error wrapping ErrTokenNotSet as the cause.
func GetFigmaToken() (string, error) {
	token := os.Getenv("FIGMA_ACCESS_TOKEN")
	if token == "" {
		return "", operr.New(operr.CategoryAuthentication,
			"Figma authentication is not configured.",
			"Set FIGMA_ACCESS_TOKEN and retry.",
			&ErrTokenNotSet{})
	}
	return token, nil
}

// ErrTokenNotSet is returned when FIGMA_ACCESS_TOKEN is not set.
type ErrTokenNotSet struct{}

func (e *ErrTokenNotSet) Error() string {
	return "FIGMA_ACCESS_TOKEN environment variable not set"
}
