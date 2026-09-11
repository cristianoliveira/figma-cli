package artifact

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/operr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileWriterWritesExactBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out", "artifact.txt")
	err := NewFileWriter().Write(context.Background(), path, []byte("hello"))
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

func TestFileWriterCreatesMissingParentDirectories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c", "artifact.txt")
	err := NewFileWriter().Write(context.Background(), path, []byte("x"))
	require.NoError(t, err)
	assert.FileExists(t, path)
}

func TestFileWriterAppliesFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "artifact.txt")
	err := NewFileWriter().Write(context.Background(), path, []byte("x"))
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	// 0644 (permissions only, ignoring type bits).
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

func TestFileWriterOverwritesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "artifact.txt")
	require.NoError(t, NewFileWriter().Write(context.Background(), path, []byte("first")))

	require.NoError(t, NewFileWriter().Write(context.Background(), path, []byte("second")))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "second", string(data))
}

func TestFileWriterEmptyBytesStillWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.txt")
	err := NewFileWriter().Write(context.Background(), path, nil)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Empty(t, data)
}

func TestFileWriterPermissionErrorClassifiesArtifactAccess(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o644))
	path := filepath.Join(blocker, "child", "out.txt")

	err := NewFileWriter().Write(context.Background(), path, []byte("data"))

	require.Error(t, err)
	var classified *operr.ClassifiedError
	require.ErrorAs(t, err, &classified)
	assert.Equal(t, operr.CategoryArtifactAccess, classified.Category)
	assert.NotContains(t, classified.Error(), blocker, "private path must not leak")

	var pathErr *os.PathError
	assert.True(t, errors.As(err, &pathErr))
}

func TestFileWriterHonoursCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	path := filepath.Join(t.TempDir(), "out.txt")
	err := NewFileWriter().Write(ctx, path, []byte("data"))

	require.Error(t, err)
	assert.False(t, fileExists(t, path), "no write should occur on cancelled context")
}

func TestClassifyFileErrorRetainsCause(t *testing.T) {
	orig := &os.PathError{Op: "open", Path: "/private/x", Err: os.ErrNotExist}
	err := ClassifyFileError(orig)

	var classified *operr.ClassifiedError
	require.ErrorAs(t, err, &classified)
	assert.Equal(t, operr.CategoryArtifactAccess, classified.Category)
	assert.True(t, errors.Is(err, os.ErrNotExist))
	assert.NotContains(t, classified.Error(), "/private/x")
}

func fileExists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	return err == nil
}
