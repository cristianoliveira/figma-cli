package cmd

import (
	"bytes"
	"net/http"

	"github.com/spf13/cobra"
)

type commandResult struct {
	Stdout string
	Stderr string
	Err    error
}

func executeCommand(command *cobra.Command, args ...string) commandResult {
	root := newRootCommand(command)
	root.SilenceErrors = true
	root.SilenceUsage = true
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(append([]string{command.Name()}, args...))
	err := root.Execute()
	return commandResult{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
