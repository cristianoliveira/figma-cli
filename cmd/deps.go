package cmd

import (
	"net/http"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/assets"
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/inspect"
)

// InspectServiceFactory builds an inspect application service. Returning
// a factory (rather than caching the service) keeps Deps free of mutable
// state and lets tests substitute a service without going through the
// loadClient path.
type InspectServiceFactory func() (InspectService, error)

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
	InspectService   InspectServiceFactory
	AssetApplication func() AssetApplication
}

// defaultResolveExecutable resolves the current binary path. It is overridable
// in tests so production readiness text can be asserted deterministically.
func defaultResolveExecutable() (string, error) {
	return cli.CurrentExecutablePath()
}

// defaultGetEnv reads from os.Getenv. Tests may inject an env stub via Deps.
func defaultGetEnv(key string) string { return os.Getenv(key) }

// defaultInspectServiceFactory builds an inspect service bound to the
// production Figma client. It is the production wiring for TASK-0003.
func defaultInspectServiceFactory(loadClient func() (*figma.Client, error)) InspectServiceFactory {
	return func() (InspectService, error) {
		client, err := loadClient()
		if err != nil {
			return nil, err
		}
		adapter := figma.NewInspectAdapter(client)
		return newInspectServiceFromAdapter(adapter, adapter), nil
	}
}

// newInspectServiceFromAdapter composes an inspect service from the
// NodeFetcher / VariableFetcher pair returned by the figma adapter. It
// lives in cmd so the wiring test can swap either port independently.
func newInspectServiceFromAdapter(nodes inspect.NodeFetcher, vars inspect.VariableFetcher) InspectService {
	return inspect.New(nodes, vars)
}

// NewProductionDeps builds the Deps that cmd/figma/main.go passes to the root
// factory. It centralises every environment-backed collaborator so the main
// function stays a flat composition root.
func NewProductionDeps() Deps {
	loadClient := cli.LoadClient
	return Deps{
		LoadClient:       loadClient,
		DownloadClient:   nil,
		FetchVariables:   figma.FetchVariables,
		ResolveExec:      defaultResolveExecutable,
		GetEnv:           defaultGetEnv,
		InspectService:   defaultInspectServiceFactory(loadClient),
		AssetApplication: defaultAssetApplication,
	}
}

// defaultAssetApplication returns the production asset application
// instance. It is the wiring for TASK-0005: the command builds the
// edge adapters and hands them, plus the request, to this pure
// application service.
func defaultAssetApplication() AssetApplication {
	return assets.NewApplication()
}

// DepsForLoadClient is a tiny helper for tests that only need to stub the
// Figma client loader. Production callers should use NewProductionDeps;
// tests can wrap their fake loadClient in this helper without having to
// construct the full Deps struct themselves.
func DepsForLoadClient(loadClient func() (*figma.Client, error)) Deps {
	return Deps{
		LoadClient:       loadClient,
		FetchVariables:   figma.FetchVariables,
		ResolveExec:      defaultResolveExecutable,
		GetEnv:           defaultGetEnv,
		AssetApplication: defaultAssetApplication,
	}
}
