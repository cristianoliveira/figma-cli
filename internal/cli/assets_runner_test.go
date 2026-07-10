package cli

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type assetRoundTripper func(*http.Request) (*http.Response, error)

func (f assetRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestExportAssetsFetchesFiltersAndDownloads(t *testing.T) {
	httpClient := &http.Client{Transport: assetRoundTripper(func(request *http.Request) (*http.Response, error) {
		body := "svg"
		switch request.URL.Path {
		case "/v1/files/FILE/nodes":
			body = `{"nodes":{"1:2":{"document":{"id":"1:2","name":"Screen","type":"FRAME","children":[{"id":"2:3","name":"Close","type":"VECTOR"},{"id":"4:5","name":"Photo","fills":[{"type":"IMAGE"}]}]}}}}`
		case "/v1/images/FILE":
			assert.Equal(t, "2:3", request.URL.Query().Get("ids"))
			body = `{"images":{"2:3":"https://cdn.example/close.svg"}}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}

	manifest, err := ExportAssets(AssetExportRequest{
		Client:          &figma.Client{HTTP: httpClient},
		FileID:          "FILE",
		NodeIDs:         []string{"1:2"},
		OutputDirectory: t.TempDir(),
		Kind:            "icon",
		Format:          "auto",
		NameFilter:      "close",
		Filename:        func(asset extract.Asset) string { return strings.ToLower(asset.Name) },
	})

	require.NoError(t, err)
	require.Len(t, manifest.Items, 1)
	assert.Equal(t, "2:3", manifest.Items[0].NodeID)
	assert.FileExists(t, manifest.Items[0].Path)
	assert.Equal(t, 1, manifest.Succeeded)
}
