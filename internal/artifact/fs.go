package artifact

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cristianoliveira/figma-cli/internal/operr"
)

const (
	defaultFileMode = 0o644
	defaultDirMode  = 0o755
)

// FileWriter is the filesystem adapter for the Writer port. It creates
// missing parent directories recursively with a fixed directory mode
// and writes the payload with a fixed file mode. Parent creation,
// open/write/close, and permission failures are translated to a neutral
// artifact-access error that retains the underlying cause.
type FileWriter struct {
	// FileMode and DirMode override the defaults when non-zero. They are
	// intentionally fixed at construction (defaults 0644/0755); callers
	// do not select arbitrary permissions per write.
	FileMode os.FileMode
	DirMode  os.FileMode
}

// NewFileWriter returns a FileWriter using the default modes.
func NewFileWriter() FileWriter {
	return FileWriter{FileMode: defaultFileMode, DirMode: defaultDirMode}
}

// Write persists data to path, creating parent directories as needed.
// Existing files are overwritten (os.WriteFile truncates). Context
// cancellation is honored before any I/O.
func (w FileWriter) Write(ctx context.Context, path string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fileMode := w.FileMode
	if fileMode == 0 {
		fileMode = defaultFileMode
	}
	dirMode := w.DirMode
	if dirMode == 0 {
		dirMode = defaultDirMode
	}

	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return ClassifyFileError(err)
	}
	if err := os.WriteFile(path, data, fileMode); err != nil {
		return ClassifyFileError(err)
	}
	return nil
}

// ClassifyFileError translates a filesystem failure into a neutral
// artifact-access error, retaining the underlying cause. Private paths
// and raw OS text are never exposed in the safe message.
func ClassifyFileError(err error) error {
	return operr.New(operr.CategoryArtifactAccess,
		"Could not access a required file.",
		"Check the file path and permissions, then retry.",
		err)
}
