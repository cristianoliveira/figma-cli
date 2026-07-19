// Package cli holds shared command-line runtime helpers — error handling and
// dependency wiring used across every cobra command in cmd/. Keeping them here
// lets cmd/ contain only command definitions.
package cli

import (
	"errors"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewPrinter builds an output.Printer bound to stdout, reading the global
// --json flag from cmd. Centralising this keeps the flag lookup in one place
// and stops every command from re-reading the same persistent flag.
func NewPrinter(cmd *cobra.Command) *output.Printer {
	asJSON, _ := cmd.Flags().GetBool("json")
	return output.New(cmd.OutOrStdout(), asJSON)
}

// ExitCodeError carries a process exit code without a diagnostic message.
// Commands return it to request a silent non-zero exit (for example
// `diff text --quiet` exits 1 when no changes match). Returning it instead of
// calling os.Exit keeps the behaviour testable in-process.
type ExitCodeError struct {
	Code int
}

func (e *ExitCodeError) Error() string { return "" }

// UsageError marks invalid command syntax or arguments. Process entrypoints map
// it to exit code 2 while preserving the original actionable diagnostic.
type UsageError struct {
	err error
}

func NewUsageError(err error) error {
	if err == nil {
		return nil
	}
	return &UsageError{err: err}
}

func (e *UsageError) Error() string { return e.err.Error() }
func (e *UsageError) Unwrap() error { return e.err }

// ExitCode maps command errors to the CLI process contract.
func ExitCode(err error) int {
	var explicit *ExitCodeError
	if errors.As(err, &explicit) {
		return explicit.Code
	}
	var usage *UsageError
	if errors.As(err, &usage) {
		return 2
	}
	return 1
}

// MarkUsageErrors classifies Cobra's positional-argument failures.
// Flag handlers must mark their errors while preserving command-specific help.
// RunE errors remain operational unless command code explicitly marks them.
func MarkUsageErrors(command *cobra.Command) {
	if command.Args != nil {
		validateArgs := command.Args
		command.Args = func(cmd *cobra.Command, args []string) error {
			return NewUsageError(validateArgs(cmd, args))
		}
	}
	for _, child := range command.Commands() {
		MarkUsageErrors(child)
	}
}

// LoadClient builds a Figma API client from the configured access token.
// Centralizing construction means HTTP config (timeouts, base URL, retries)
// has exactly one place to change.
func LoadClient() (*figma.Client, error) {
	token, err := env.GetFigmaToken()
	if err != nil {
		return nil, err
	}
	return figma.NewClient(token), nil
}
