package assets

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSource records candidates and their order.
type fakeSource struct {
	candidates []extract.Asset
	err        error
	calls      int
	gotCtx     context.Context
}

func (f *fakeSource) Candidates(ctx context.Context) ([]extract.Asset, error) {
	f.calls++
	f.gotCtx = ctx
	if f.err != nil {
		return nil, f.err
	}
	return f.candidates, nil
}

// fakeURLSource records every URL request in call order.
type fakeURLSource struct {
	urls   map[string]string // key: "nodeID|format"
	errFor map[string]error
	order  []string
	gotCtx []context.Context
}

func (f *fakeURLSource) ExportURL(ctx context.Context, nodeID, format string) (string, error) {
	key := nodeID + "|" + format
	f.order = append(f.order, key)
	f.gotCtx = append(f.gotCtx, ctx)
	if err, ok := f.errFor[key]; ok {
		return "", err
	}
	if u, ok := f.urls[key]; ok {
		return u, nil
	}
	return "", fmt.Errorf("no URL for %s", key)
}

// fakeSink records every write in call order.
type fakeSink struct {
	written map[string]string // path -> url
	order   []string
	errFor  map[string]error
	gotCtx  []context.Context
}

func (f *fakeSink) Write(ctx context.Context, path, url string) error {
	if f.written == nil {
		f.written = map[string]string{}
	}
	f.written[path] = url
	f.order = append(f.order, path)
	f.gotCtx = append(f.gotCtx, ctx)
	if err, ok := f.errFor[path]; ok {
		return err
	}
	return nil
}

func sampleAssets() []extract.Asset {
	return []extract.Asset{
		{ID: "1:1", Name: "Login Button", Kind: "instance", Format: "svg"},
		{ID: "1:2", Name: "Photo", Kind: "image", Format: "png"},
		{ID: "1:3", Name: "Icon", Kind: "vector", Format: "svg"},
	}
}

func newFullRequest(src AssetSource, urls *fakeURLSource, sink *fakeSink) ApplicationRequest {
	return ApplicationRequest{
		Source:          src,
		URLSource:       urls,
		Sink:            sink,
		Kind:            "all",
		Format:          "auto",
		Filename:        func(a extract.Asset) string { return strings.ToLower(a.Name) },
		OutputDirectory: "out",
	}
}

func TestApplicationFilteringKinds(t *testing.T) {
	tests := []struct {
		name    string
		kind    string
		wantIDs []string
	}{
		{name: "all", kind: "all", wantIDs: []string{"1:1", "1:2", "1:3"}},
		{name: "icon includes instance and vector", kind: "icon", wantIDs: []string{"1:1", "1:3"}},
		{name: "image", kind: "image", wantIDs: []string{"1:2"}},
		{name: "instance", kind: "instance", wantIDs: []string{"1:1"}},
		{name: "vector", kind: "vector", wantIDs: []string{"1:3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := &fakeSource{candidates: sampleAssets()}
			urls := &fakeURLSource{urls: map[string]string{
				"1:1|svg": "https://cdn/a.svg",
				"1:2|png": "https://cdn/a.png",
				"1:3|svg": "https://cdn/b.svg",
			}}
			sink := &fakeSink{}
			req := newFullRequest(src, urls, sink)
			req.Kind = tt.kind

			manifest, err := NewApplication().Run(context.Background(), req)
			require.NoError(t, err)
			require.Len(t, manifest.Items, len(tt.wantIDs))
			for i, id := range tt.wantIDs {
				assert.Equal(t, id, manifest.Items[i].NodeID)
			}
		})
	}
}

func TestApplicationIconExcludesImage(t *testing.T) {
	src := &fakeSource{candidates: sampleAssets()}
	urls := &fakeURLSource{urls: map[string]string{
		"1:1|svg": "https://cdn/a.svg",
		"1:3|svg": "https://cdn/b.svg",
	}}
	sink := &fakeSink{}
	req := newFullRequest(src, urls, sink)
	req.Kind = "icon"

	manifest, err := NewApplication().Run(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, manifest.Items, 2)
	assert.Equal(t, "1:1", manifest.Items[0].NodeID)
	assert.Equal(t, "1:3", manifest.Items[1].NodeID)
	assert.Equal(t, 2, manifest.Succeeded)
}

func TestApplicationNameFilterIsCaseInsensitive(t *testing.T) {
	src := &fakeSource{candidates: sampleAssets()}
	urls := &fakeURLSource{urls: map[string]string{"1:2|png": "https://cdn/a.png"}}
	sink := &fakeSink{}
	req := newFullRequest(src, urls, sink)
	req.NameFilter = "PHOTO"

	manifest, err := NewApplication().Run(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, manifest.Items, 1)
	assert.Equal(t, "1:2", manifest.Items[0].NodeID)
}

func TestApplicationFormatOverride(t *testing.T) {
	src := &fakeSource{candidates: sampleAssets()}
	urls := &fakeURLSource{urls: map[string]string{
		"1:1|png": "https://cdn/a.png",
		"1:2|png": "https://cdn/b.png",
		"1:3|png": "https://cdn/c.png",
	}}
	sink := &fakeSink{}
	req := newFullRequest(src, urls, sink)
	req.Format = "png"

	manifest, err := NewApplication().Run(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, manifest.Items, 3)
	for _, item := range manifest.Items {
		assert.Equal(t, "png", item.Format)
	}
}

func TestApplicationURLFailureRecordsItemErrorAndContinues(t *testing.T) {
	src := &fakeSource{candidates: sampleAssets()}
	urls := &fakeURLSource{
		urls: map[string]string{
			"1:1|svg": "https://cdn/a.svg",
			"1:3|svg": "https://cdn/c.svg",
		},
		errFor: map[string]error{
			"1:2|png": errors.New("export URL failed"),
		},
	}
	sink := &fakeSink{}
	req := newFullRequest(src, urls, sink)

	manifest, err := NewApplication().Run(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, manifest.Items, 3)
	assert.Equal(t, "1:2", manifest.Items[1].NodeID)
	assert.Empty(t, manifest.Items[1].Path)
	assert.Equal(t, "export URL failed", manifest.Items[1].Error)
	assert.Equal(t, 1, manifest.Failed)
	assert.Equal(t, 2, manifest.Succeeded)
}

func TestApplicationSinkFailureRecordsItemErrorAndContinues(t *testing.T) {
	src := &fakeSource{candidates: sampleAssets()}
	urls := &fakeURLSource{urls: map[string]string{
		"1:1|svg": "https://cdn/a.svg",
		"1:2|png": "https://cdn/b.png",
		"1:3|svg": "https://cdn/c.svg",
	}}
	sink := &fakeSink{errFor: map[string]error{
		"out/login button.svg": errors.New("write failed"),
	}}
	req := newFullRequest(src, urls, sink)

	manifest, err := NewApplication().Run(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, manifest.Items, 3)
	assert.Equal(t, "1:1", manifest.Items[0].NodeID)
	assert.Empty(t, manifest.Items[0].Path)
	assert.Equal(t, "write failed", manifest.Items[0].Error)
	assert.Equal(t, 1, manifest.Failed)
	assert.Equal(t, 2, manifest.Succeeded)
}

func TestApplicationFilenameCollisionPolicy(t *testing.T) {
	// Two assets with the same name should receive deterministic
	// suffixes in request order.
	candidates := []extract.Asset{
		{ID: "1:1", Name: "Icon", Kind: "vector", Format: "svg"},
		{ID: "1:2", Name: "Icon", Kind: "vector", Format: "svg"},
		{ID: "1:3", Name: "Icon", Kind: "vector", Format: "svg"},
	}
	src := &fakeSource{candidates: candidates}
	urls := &fakeURLSource{urls: map[string]string{
		"1:1|svg": "https://cdn/a.svg",
		"1:2|svg": "https://cdn/b.svg",
		"1:3|svg": "https://cdn/c.svg",
	}}
	sink := &fakeSink{}
	req := newFullRequest(src, urls, sink)
	req.Filename = func(a extract.Asset) string { return "icon" }
	req.OutputDirectory = ""

	manifest, err := NewApplication().Run(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, manifest.Items, 3)
	assert.Equal(t, "icon.svg", manifest.Items[0].Path)
	assert.Equal(t, "icon-2.svg", manifest.Items[1].Path)
	assert.Equal(t, "icon-3.svg", manifest.Items[2].Path)
}

func TestApplicationRequestOrderPreserved(t *testing.T) {
	src := &fakeSource{candidates: sampleAssets()}
	urls := &fakeURLSource{urls: map[string]string{
		"1:1|svg": "https://cdn/a.svg",
		"1:2|png": "https://cdn/b.png",
		"1:3|svg": "https://cdn/c.svg",
	}}
	sink := &fakeSink{}
	req := newFullRequest(src, urls, sink)

	manifest, err := NewApplication().Run(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, []string{"1:1|svg", "1:2|png", "1:3|svg"}, urls.order)
	assert.Equal(t, []string{"out/login button.svg", "out/photo.png", "out/icon.svg"}, sink.order)
	assert.Equal(t, []string{"1:1", "1:2", "1:3"}, []string{
		manifest.Items[0].NodeID, manifest.Items[1].NodeID, manifest.Items[2].NodeID,
	})
}

func TestApplicationEmptyRequestNoIOAndEmptyManifest(t *testing.T) {
	src := &fakeSource{candidates: nil}
	urls := &fakeURLSource{}
	sink := &fakeSink{}

	manifest, err := NewApplication().Run(context.Background(), ApplicationRequest{
		Source:    src,
		URLSource: urls,
		Sink:      sink,
	})

	require.NoError(t, err)
	require.NotNil(t, manifest.Items)
	assert.Empty(t, manifest.Items)
	assert.Zero(t, manifest.Succeeded)
	assert.Zero(t, manifest.Failed)
	assert.Equal(t, 1, src.calls, "source is called once to discover candidates")
	assert.Empty(t, urls.order, "no URL resolution for empty candidates")
	assert.Empty(t, sink.order, "no sink write for empty candidates")
}

func TestApplicationNilSourceReturnsEmptyManifestNoError(t *testing.T) {
	manifest, err := NewApplication().Run(context.Background(), ApplicationRequest{})
	require.NoError(t, err)
	require.NotNil(t, manifest.Items)
	assert.Empty(t, manifest.Items)
	assert.Zero(t, manifest.Succeeded)
	assert.Zero(t, manifest.Failed)
}

func TestApplicationMissingPortsErrors(t *testing.T) {
	src := &fakeSource{candidates: sampleAssets()}
	_, err := NewApplication().Run(context.Background(), ApplicationRequest{
		Source: src,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMissingPort))
}

func TestApplicationContextReachesAllPorts(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "marker")

	src := &fakeSource{candidates: sampleAssets()}
	urls := &fakeURLSource{urls: map[string]string{
		"1:1|svg": "https://cdn/a.svg",
		"1:2|png": "https://cdn/b.png",
		"1:3|svg": "https://cdn/c.svg",
	}}
	sink := &fakeSink{}
	req := newFullRequest(src, urls, sink)

	_, err := NewApplication().Run(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, "marker", src.gotCtx.Value(ctxKey{}))
	for _, got := range urls.gotCtx {
		assert.Equal(t, "marker", got.Value(ctxKey{}))
	}
	for _, got := range sink.gotCtx {
		assert.Equal(t, "marker", got.Value(ctxKey{}))
	}
}

func TestApplicationSourceErrorIsTopLevel(t *testing.T) {
	src := &fakeSource{err: errors.New("document fetch failed")}
	urls := &fakeURLSource{}
	sink := &fakeSink{}

	manifest, err := NewApplication().Run(context.Background(), ApplicationRequest{
		Source:    src,
		URLSource: urls,
		Sink:      sink,
	})

	require.Error(t, err)
	assert.Equal(t, "document fetch failed", err.Error())
	assert.Empty(t, manifest.Items)
	assert.Empty(t, urls.order, "no URL resolution on source failure")
	assert.Empty(t, sink.order, "no sink write on source failure")
}

func TestApplicationPathUsesOutputDirectory(t *testing.T) {
	src := &fakeSource{candidates: []extract.Asset{{ID: "1:1", Name: "A", Kind: "vector", Format: "svg"}}}
	urls := &fakeURLSource{urls: map[string]string{"1:1|svg": "https://cdn/a.svg"}}
	sink := &fakeSink{}
	req := newFullRequest(src, urls, sink)
	req.OutputDirectory = filepath.Join("deep", "nested", "dir")

	manifest, err := NewApplication().Run(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, manifest.Items, 1)
	assert.Equal(t, filepath.Join("deep", "nested", "dir", "a.svg"), manifest.Items[0].Path)
}
