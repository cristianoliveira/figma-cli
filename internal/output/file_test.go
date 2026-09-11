package output

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/operr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFilePermissionErrorClassifiesArtifactAccess(t *testing.T) {
	// Write into a path whose parent is a file, forcing a filesystem error.
	blocker := filepath.Join(t.TempDir(), "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o644))
	path := filepath.Join(blocker, "child", "out.txt")

	err := WriteFile(path, []byte("data"), 0o644)

	require.Error(t, err)
	var classified *operr.ClassifiedError
	require.ErrorAs(t, err, &classified)
	assert.Equal(t, operr.CategoryArtifactAccess, classified.Category)
	assert.Equal(t, "Could not access a required file.", classified.Message)
	assert.NotContains(t, classified.Error(), blocker, "private path must not leak")

	// Cause retention: the underlying path error is still reachable.
	var pathErr *os.PathError
	assert.True(t, errors.As(err, &pathErr))
}

func TestWriteFileSuccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "out.txt")
	require.NoError(t, WriteFile(path, []byte("data"), 0o644))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "data", string(data))
}
