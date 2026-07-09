package cli

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrinter_HonorsJSONFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.PersistentFlags().Bool("json", false, "")
	require.NoError(t, cmd.PersistentFlags().Set("json", "true"))

	p := NewPrinter(cmd)

	assert.NotNil(t, p)
}

func TestNewPrinter_DefaultsToFalse(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.PersistentFlags().Bool("json", false, "")

	p := NewPrinter(cmd)

	assert.NotNil(t, p)
}

func TestRunSimpleFetch_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	t.Setenv("FIGMA_ACCESS_TOKEN", "test-token")
	cmd := newJSONCmd()

	err := RunSimpleFetch(cmd, []string{"file123"},
		func(fileID string) (string, error) { return server.URL, nil },
		"test",
	)

	require.NoError(t, err)
}

func TestRunSimpleFetch_BadInput(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "test-token")
	cmd := newJSONCmd()

	err := RunSimpleFetch(cmd, []string{"https://www.figma.com/community/x"},
		func(string) (string, error) { return "", nil },
		"test",
	)

	require.Error(t, err)
}

func TestRunSimpleFetch_MissingToken(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "")
	cmd := newJSONCmd()

	err := RunSimpleFetch(cmd, []string{"file123"},
		func(string) (string, error) { return "https://example.com", nil },
		"test",
	)

	require.Error(t, err)
}

func TestRunSimpleFetch_BuildURLError(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "test-token")
	cmd := newJSONCmd()
	wantErr := errors.New("boom")

	err := RunSimpleFetch(cmd, []string{"file123"},
		func(string) (string, error) { return "", wantErr },
		"test",
	)

	require.ErrorIs(t, err, wantErr)
}

func TestRunSimpleFetch_FetchError(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "test-token")
	cmd := newJSONCmd()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	err := RunSimpleFetch(cmd, []string{"file123"},
		func(string) (string, error) { return server.URL, nil },
		"test",
	)

	require.Error(t, err)
}

func TestRunSimpleFetch_PassesFileID(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "test-token")
	var seenFileID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	cmd := newJSONCmd()

	err := RunSimpleFetch(cmd, []string{"grnVU2vAihHXwYgHryu2xE"},
		func(fileID string) (string, error) {
			seenFileID = fileID
			return server.URL, nil
		},
		"test",
	)

	require.NoError(t, err)
	assert.Equal(t, "grnVU2vAihHXwYgHryu2xE", seenFileID)
}

func newJSONCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.PersistentFlags().Bool("json", false, "")
	return cmd
}
