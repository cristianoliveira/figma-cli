package comments

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestFetchMapsCommentsAndBuildsWebURLs(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		assert.True(t, strings.HasSuffix(request.URL.Path, "/comments"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`{"comments":[{
				"id":"comment-1","message":"Adjust spacing","created_at":"2026-01-01T00:00:00Z","file_key":"abc",
				"client_meta":{"node_id":"1:2","node_offset":{"x":0,"y":0}},"reactions":[],
				"user":{"handle":"Ada","id":"1","img_url":""}
			}]}`)),
		}, nil
	})}}

	outputs, err := Fetch(client, "abc")

	require.NoError(t, err)
	require.Len(t, outputs, 1)
	assert.Equal(t, "comment-1", outputs[0].ID)
	assert.Equal(t, "https://www.figma.com/design/abc?m=dev&node-id=1-2#comment-1", outputs[0].URL)
}
