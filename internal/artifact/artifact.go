// Package artifact owns durable artifact persistence, exposed through a
// narrow consumer-owned port. Rendering (internal/output) and durable
// storage change for different reasons; this package owns only the
// write-bytes-to-a-path responsibility.
package artifact

import "context"

// Writer persists a byte payload to a destination path. Implementations
// own parent-directory creation, permissions, open/write/close, and OS
// error translation. It is deliberately narrow: no reads, downloads,
// rendering, or arbitrary I/O operations.
type Writer interface {
	Write(ctx context.Context, path string, data []byte) error
}

// WriterFunc adapts a function to the Writer interface.
type WriterFunc func(ctx context.Context, path string, data []byte) error

func (f WriterFunc) Write(ctx context.Context, path string, data []byte) error {
	return f(ctx, path, data)
}
