# Logger Package

This package provides a structured, context-aware logger for Go services, built on top of Go's `log/slog` package with JSON output.

## Features (Implemented)

- **Structured Logging**: Uses JSON format for logs, making them easy to parse and analyze.
- **Log Levels**: Supports Debug, Info, Warning, Error, and Fatal log levels.
- **Context Support**: All log functions accept a `context.Context` to include request-scoped data.
- **Operation ID Tracking**: Automatically includes an `operation_id` from context (if available) in each log entry for traceability.
- **Custom Fields**: Supports adding custom key-value fields to log entries.
- **Error Logging**: Provides a `LogError` function that logs error details, stack trace, and error code (integrates with `syserr` package).
- **Configurable Initialization**: Allows configuration of log level, output destination, source information, and attribute replacement via the `Init` function and `Config` struct.
- **Source Information**: Optionally includes file, line, and function name in logs (via `AddSource`).
- **Thread-Safe Initialization**: Ensures logger is initialized only once using `sync.Once`.

## Possible Future Enhancements

- **Log Rotation**: Support for automatic log file rotation (e.g., using `lumberjack`).
- **Log Sampling**: Ability to sample logs to reduce volume in high-traffic environments.
- **Sensitive Data Masking**: Automatic masking or filtering of sensitive fields (e.g., passwords, tokens).
- **Default Context Fields**: Add more default fields (e.g., environment, service name) to every log entry.
- **Performance Optimization**: Use pooling for field conversion to reduce allocations.
- **Log Metrics**: Track and expose metrics about log volume and levels.
- **Log Filtering**: Ability to filter out certain log entries based on rules or environment.
- **Integration with Log Aggregators**: Out-of-the-box support for sending logs to external systems (e.g., ELK, Datadog).
- **Custom Timestamp Formatting**: Allow configuration of timestamp format in logs.

---

## Usage Example

```go
import "tixgo/internal/common/logger"

func main() {
    logger.Init(&logger.Config{
        Level:     slog.LevelInfo,
        Output:    os.Stdout,
        AddSource: true,
    })

    ctx := context.Background()
    logger.Info(ctx, "Service started", logger.F("version", "1.0.0"))
}
```

---

## Contributing

Feel free to open issues or pull requests to discuss or contribute new features!
