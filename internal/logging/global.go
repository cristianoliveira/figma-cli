package logging

import (
	"context"
	"sync"
)

var (
	defaultLoggerMu   sync.RWMutex
	defaultLogger     Logger
	defaultLoggerOnce sync.Once
)

// Default returns the default logger.
// If not initialized, returns a logger with default configuration.
func Default() Logger {
	defaultLoggerOnce.Do(func() {
		cfg := DefaultConfig()
		// Apply environment variables if present
		cfg = ApplyEnv(cfg)
		var err error
		logger, err := NewLogger(cfg)
		if err != nil {
			// Fallback to a basic logger
			logger = NewNopLogger()
		}
		defaultLoggerMu.Lock()
		defaultLogger = logger
		defaultLoggerMu.Unlock()
	})
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

// SetDefault sets the default logger.
func SetDefault(logger Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	defaultLogger = logger
}

// Debug logs a debug message using the default logger.
func Debug(ctx context.Context, msg string, fields ...Field) {
	Default().Debug(ctx, msg, fields...)
}

// Info logs an info message using the default logger.
func Info(ctx context.Context, msg string, fields ...Field) {
	Default().Info(ctx, msg, fields...)
}

// Warn logs a warning message using the default logger.
func Warn(ctx context.Context, msg string, fields ...Field) {
	Default().Warn(ctx, msg, fields...)
}

// Error logs an error message using the default logger.
func Error(ctx context.Context, msg string, fields ...Field) {
	Default().Error(ctx, msg, fields...)
}

// Fatal logs a fatal message using the default logger.
func Fatal(ctx context.Context, msg string, fields ...Field) {
	Default().Fatal(ctx, msg, fields...)
}
