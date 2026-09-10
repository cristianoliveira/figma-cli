package cmd

import (
	"net/http"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
)

// Deps carries the explicit dependencies that leaf commands need. The struct
// stays narrow on purpose: one factory function per external capability, no
// shared mutable state. Tests construct Deps with fakes; the production
// composition root in cmd/figma/main.go wires the real collaborators.
type Deps struct {
	LoadClient       func() (*figma.Client, error)
	DownloadClient   *http.Client
	FetchVariables   func(*figma.Client, string) (map[string]any, error)
	ResolveExec      func() (string, error)
	GetEnv           func(string) string
	StdoutEnvPrinter func() (string, error)
}

// defaultResolveExecutable resolves the current binary path. It is overridable
// in tests so production readiness text can be asserted deterministically.
func defaultResolveExecutable() (string, error) {
	return cli.CurrentExecutablePath()
}

// defaultGetEnv reads from os.Getenv. Tests may inject an env stub via Deps.
func defaultGetEnv(key string) string { return os.Getenv(key) }

// NewProductionDeps builds the Deps that cmd/figma/main.go passes to the root
// factory. It centralises every environment-backed collaborator so the main
// function stays a flat composition root.
func NewProductionDeps() Deps {
	return Deps{
		LoadClient:     cli.LoadClient,
		DownloadClient: nil,
		FetchVariables: figma.FetchVariables,
		ResolveExec:    defaultResolveExecutable,
		GetEnv:         defaultGetEnv,
	}
}

// DepsForLoadClient is a tiny helper for tests that only need to stub the
// Figma client loader. Production callers should use NewProductionDeps;
// tests can wrap their fake loadClient in this helper without having to
// construct the full Deps struct themselves.
func DepsForLoadClient(loadClient func() (*figma.Client, error)) Deps {
	return Deps{
		LoadClient:     loadClient,
		FetchVariables: figma.FetchVariables,
		ResolveExec:    defaultResolveExecutable,
		GetEnv:         defaultGetEnv,
	}
}
