package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is a structured logger interface that supports context and leveled logging.
type Logger interface {
	// Debug logs a debug message with optional fields.
	Debug(ctx context.Context, msg string, fields ...Field)
	// Info logs an info message with optional fields.
	Info(ctx context.Context, msg string, fields ...Field)
	// Warn logs a warning message with optional fields.
	Warn(ctx context.Context, msg string, fields ...Field)
	// Error logs an error message with optional fields.
	Error(ctx context.Context, msg string, fields ...Field)
	// Fatal logs a fatal message and exits the program.
	Fatal(ctx context.Context, msg string, fields ...Field)
	// With returns a new logger with additional fields.
	With(fields ...Field) Logger
	// Sync flushes any buffered log entries.
	Sync() error
}

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value interface{}
}

// NewField creates a new Field.
func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Config holds logging configuration.
type Config struct {
	// Level is the minimum log level (debug, info, warn, error).
	Level string
	// Format is the log format (json, text).
	Format string
	// File is the path to the log file. If empty, logs to stderr.
	File string
	// MaxSize is the maximum size in megabytes of the log file before rotation.
	MaxSize int
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int
	// MaxAge is the maximum number of days to retain old log files.
	MaxAge int
	// Compress determines if rotated log files should be compressed.
	Compress bool
}

// DefaultConfig returns a default logging configuration.
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "text",
		File:       "",
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}
}

// NewLogger creates a new Logger based on the configuration.
func NewLogger(cfg Config) (Logger, error) {
	// Create encoder config for zap
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Set log level
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	// Create encoder
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create write syncers
	var writeSyncer zapcore.WriteSyncer
	if cfg.File != "" {
		// File output with rotation
		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		writeSyncer = zapcore.AddSync(lumberjackLogger)
	} else {
		// Standard error output
		writeSyncer = zapcore.AddSync(os.Stderr)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Create zap logger with caller skipping
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	// Wrap zap logger
	return &zapLoggerWrapper{
		logger: zapLogger.Sugar(),
		config: cfg,
	}, nil
}

// zapLoggerWrapper wraps zap.SugaredLogger to implement Logger interface.
type zapLoggerWrapper struct {
	logger *zap.SugaredLogger
	config Config
}

func (z *zapLoggerWrapper) Debug(ctx context.Context, msg string, fields ...Field) {
	z.logWithContext(ctx, z.logger.Debugw, msg, fields...)
}

func (z *zapLoggerWrapper) Info(ctx context.Context, msg string, fields ...Field) {
	z.logWithContext(ctx, z.logger.Infow, msg, fields...)
}

func (z *zapLoggerWrapper) Warn(ctx context.Context, msg string, fields ...Field) {
	z.logWithContext(ctx, z.logger.Warnw, msg, fields...)
}

func (z *zapLoggerWrapper) Error(ctx context.Context, msg string, fields ...Field) {
	z.logWithContext(ctx, z.logger.Errorw, msg, fields...)
}

func (z *zapLoggerWrapper) Fatal(ctx context.Context, msg string, fields ...Field) {
	z.logWithContext(ctx, z.logger.Fatalw, msg, fields...)
}

func (z *zapLoggerWrapper) With(fields ...Field) Logger {
	zapFields := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		zapFields = append(zapFields, f.Key, f.Value)
	}
	return &zapLoggerWrapper{
		logger: z.logger.With(zapFields...),
		config: z.config,
	}
}

func (z *zapLoggerWrapper) Sync() error {
	return z.logger.Sync()
}

// logWithContext adds request ID from context and calls the zap logging function.
func (z *zapLoggerWrapper) logWithContext(
	ctx context.Context,
	logFunc func(msg string, keysAndValues ...interface{}),
	msg string,
	fields ...Field,
) {
	allFields := make([]interface{}, 0, len(fields)*2+2) // +2 for request ID

	// Add request ID from context if present
	if requestID, ok := RequestIDFromContext(ctx); ok {
		allFields = append(allFields, "request_id", requestID)
	}

	// Add user-provided fields
	for _, f := range fields {
		allFields = append(allFields, f.Key, f.Value)
	}

	logFunc(msg, allFields...)
}

// NewNopLogger returns a no-op logger that discards all logs.
func NewNopLogger() Logger {
	return &nopLogger{}
}

type nopLogger struct{}

func (n *nopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}
func (n *nopLogger) Info(ctx context.Context, msg string, fields ...Field)  {}
func (n *nopLogger) Warn(ctx context.Context, msg string, fields ...Field)  {}
func (n *nopLogger) Error(ctx context.Context, msg string, fields ...Field) {}
func (n *nopLogger) Fatal(ctx context.Context, msg string, fields ...Field) {}
func (n *nopLogger) With(fields ...Field) Logger                            { return n }
func (n *nopLogger) Sync() error                                            { return nil }
