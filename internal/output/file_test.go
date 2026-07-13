package output

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFileCreatesMissingParentDirectories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "nested", "result.txt")

	err := WriteFile(path, []byte("result"), 0o600)

	require.NoError(t, err)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "result", string(content))
}

func TestCreateFileCreatesMissingParentDirectories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "nested", "result.txt")

	file, err := CreateFile(path)
	require.NoError(t, err)
	_, err = io.WriteString(file, "result")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "result", string(content))
}
