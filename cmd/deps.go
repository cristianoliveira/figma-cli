package cmd

import (
	"context"
	"net/http"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/artifact"
	"github.com/cristianoliveira/figma-cli/internal/assets"
	"github.com/cristianoliveira/figma-cli/internal/assetsedge"
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/inspect"
	"github.com/cristianoliveira/figma-cli/internal/tokens"
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
	TokenService     func() (TokenService, error)
	ArtifactWriter   artifact.Writer
	DownloadAsset    func(ctx context.Context, httpClient *http.Client, path, url string) error
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
	writer := artifact.NewFileWriter()
	return Deps{
		LoadClient:       loadClient,
		DownloadClient:   nil,
		FetchVariables:   figma.FetchVariables,
		ResolveExec:      defaultResolveExecutable,
		GetEnv:           defaultGetEnv,
		InspectService:   defaultInspectServiceFactory(loadClient),
		AssetApplication: defaultAssetApplication,
		TokenService:     defaultTokenServiceFactory(loadClient, writer),
		ArtifactWriter:   writer,
		DownloadAsset:    defaultDownloadAsset,
	}
}

// defaultTokenServiceFactory builds a token service bound to the Figma
// client and the supplied artifact sink. It is the wiring for TASK-0008
// (source policy in internal/tokens) composed with the TASK-0010
// artifact persistence port; the sink is the same Deps.ArtifactWriter
// used by other workflows.
func defaultTokenServiceFactory(loadClient func() (*figma.Client, error), sink tokens.ArtifactSink) func() (TokenService, error) {
	return func() (TokenService, error) {
		client, err := loadClient()
		if err != nil {
			return nil, err
		}
		adapter := figma.NewTokensAdapters(client)
		return tokens.New(
			tokens.Policy{
				Variables: tokens.VariablesSource{Fetch: adapter.VariablesFetcher()},
				Styles:    tokens.StylesSource{FetchStyles: adapter.StylesFetcher(), FetchNodes: adapter.NodesFetcher()},
				Scan:      tokens.ScanSource{FetchDocument: adapter.DocumentFetcher()},
			},
			tokens.FormatterFunc(extract.FormatTokens),
			sink,
		), nil
	}
}

// defaultDownloadAsset is the URL-aware asset download used by the
// export command. It wraps assetsedge.DownloadFile so the command never
// calls the concrete HTTP/filesystem helper directly.
func defaultDownloadAsset(ctx context.Context, httpClient *http.Client, path, url string) error {
	_ = ctx
	return assetsedge.DownloadFile(httpClient, path, url)
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
	writer := artifact.NewFileWriter()
	return Deps{
		LoadClient:       loadClient,
		FetchVariables:   figma.FetchVariables,
		ResolveExec:      defaultResolveExecutable,
		GetEnv:           defaultGetEnv,
		AssetApplication: defaultAssetApplication,
		TokenService:     defaultTokenServiceFactory(loadClient, writer),
		ArtifactWriter:   writer,
		DownloadAsset:    defaultDownloadAsset,
	}
}
