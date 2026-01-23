# Structured Logging Infrastructure

This project uses structured logging powered by [zap](https://github.com/uber-go/zap) and [lumberjack](https://github.com/natefinch/lumberjack) for log rotation.

## Features

- **Structured logging**: JSON or text format with key‑value fields
- **Request‑ID tracing**: Automatic request‑ID generation and propagation via context
- **Configurable levels**: debug, info, warn, error
- **Configurable outputs**: stderr (default) or file with rotation
- **API client logging**: Middleware that logs all API requests/responses with timing and error tracking
- **Environment‑based configuration**: No code changes required for most settings

## Usage

### Basic Logging

Import the logging package:

```go
import "github.com/cristianoliveira/figma-cli/internal/logging"
```

Use the default logger (automatically configured from environment):

```go
ctx := context.Background()
logging.Info(ctx, "message")
```

Add structured fields:

```go
logging.Info(ctx, "user logged in",
    logging.String("user_id", "123"),
    logging.Int("attempt", 3),
    logging.Err(err),
)
```

### Creating a Logger Instance

For more control, create a logger with a specific configuration:

```go
cfg := logging.DefaultConfig()
cfg.Level = "debug"
cfg.Format = "json"
cfg.File = "/var/log/figma-cli.log"

logger, err := logging.NewLogger(cfg)
if err != nil {
    // handle error
}
defer logger.Sync()

logger.Info(ctx, "custom logger")
```

### Request IDs

Generate a request ID and attach it to a context:

```go
ctx = logging.NewContextWithRequestID(ctx)
```

All logs written with that context will automatically include a `request_id` field.

Extract the request ID from a context:

```go
id := logging.RequestID(ctx)
```

### API Client Logging

Wrap any `api.Client` implementation with logging:

```go
import "github.com/cristianoliveira/figma-cli/internal/api"

var rawClient api.Client // your concrete client
loggingClient := api.NewLoggingClient(rawClient, logger)
```

The wrapper logs each request at debug level, records the duration, and logs errors at error level.

## Configuration

Logging is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Minimum log level: `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `text` | Output format: `json` or `text` |
| `LOG_FILE` | (empty) | If set, logs are written to this file (rotation enabled) |
| `LOG_MAX_SIZE` | `10` | Max log file size in megabytes before rotation |
| `LOG_MAX_BACKUPS` | `5` | Max number of old log files to keep |
| `LOG_MAX_AGE` | `30` | Max age of a log file in days |
| `LOG_COMPRESS` | `true` | Whether to compress rotated logs (`true`/`false`) |

Example:

```bash
export LOG_LEVEL=debug
export LOG_FORMAT=json
export LOG_FILE=./logs/figma-cli.log
export LOG_MAX_SIZE=100
./figma-cli parse "https://figma.com/..."
```

## File Rotation

When `LOG_FILE` is set, logs are automatically rotated by lumberjack:

- Logs are written to the specified file
- When the file reaches `LOG_MAX_SIZE` megabytes, it is rotated
- Up to `LOG_MAX_BACKUPS` old files are kept
- Files older than `LOG_MAX_AGE` days are deleted
- Rotated logs can be compressed (if `LOG_COMPRESS=true`)

## Performance

The logger uses **zap**, a high‑performance logging library. All logging calls are non‑blocking and have minimal overhead.

## Testing

Use `logging.NewNopLogger()` to obtain a no‑op logger for tests.

## Example Output

**Text format (development):**
```
2026-01-23T10:24:25.852+0100	info	logging/logger.go:156	message with fields	{"key": "value", "count": 42}
```

**JSON format (production):**
```json
{"level":"info","timestamp":"2026-01-23T10:24:25.852+0100","caller":"logging/logger.go:156","message":"message with request id","request_id":"test-request-123"}
```