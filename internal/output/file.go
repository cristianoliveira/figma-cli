package output

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cristianoliveira/figma-cli/internal/operr"
)

// CreateFile creates path and any missing parent directories.
func CreateFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, classifyFileError(err)
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, classifyFileError(err)
	}
	return file, nil
}

// WriteFile writes data to path after creating any missing parent directories.
func WriteFile(path string, data []byte, permission fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return classifyFileError(err)
	}
	if err := os.WriteFile(path, data, permission); err != nil {
		return classifyFileError(err)
	}
	return nil
}

// classifyFileError translates a filesystem failure into a neutral
// artifact-access error, retaining the original *os.PathError as the
// cause. Private paths and raw OS text are never exposed.
func classifyFileError(err error) error {
	return operr.New(operr.CategoryArtifactAccess,
		"Could not access a required file.",
		"Check the file path and permissions, then retry.",
		err)
}
