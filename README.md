# gox

A modular Go library providing reusable, production-ready packages for building robust Go services. This project is designed as a public library to help you reuse custom packages (such as logging, error handling, context management, event bus, server setup, and more) across multiple Go projects.

## Project Goals

- **Reusability:** Provide a set of well-designed, framework-agnostic packages for common backend needs.
- **Consistency:** Standardize logging, error handling, context propagation, and more across your Go services.
- **Extensibility:** Make it easy to extend or customize each package for your own needs.

## Packages

### 1. `syserr` — Structured Error Handling

- Rich error information, stack traces, error codes, and metadata fields.
- Type-safe error codes and error wrapping.
- Helper functions for extracting codes, fields, and stack traces.

**Usage Example:**
```go
import "github.com/duongptryu/gox/syserr"

err := syserr.New(syserr.InternalCode, "something went wrong", syserr.F("user_id", 123))
wrapped := syserr.Wrap(err, syserr.InternalCode, "failed to process request")
code := syserr.GetCodeFromGenericError(wrapped)
```

---

### 2. `logger` — Structured, Context-Aware Logging

- Built on Go's `log/slog` with JSON output.
- Supports log levels, context propagation, operation/request IDs, and custom fields.
- Integrates with `syserr` for error logging.

**Usage Example:**
```go
import "github.com/duongptryu/gox/logger"

logger.Init(&logger.Config{
    Level:     slog.LevelInfo,
    Output:    os.Stdout,
    AddSource: true,
})

ctx := context.Background()
logger.Info(ctx, "Service started", logger.F("version", "1.0.0"))
```

---

### 3. `context` — Context Utilities

- Manage operation IDs, request IDs, and user context in a type-safe, framework-agnostic way.
- Designed for traceability and correlation across service boundaries.

**Usage Example:**
```go
import pkgContext "github.com/duongptryu/gox/context"

ctx := context.Background()
ctx = pkgContext.WithOperationID(ctx, "operation-123")
operationID := pkgContext.GetOperationID(ctx)
```

---

### 4. `eventbus` — CQRS Event Bus

- Built on [Watermill](https://watermill.io/), supports command and event handling for distributed systems.
- Register handlers, publish/subscribe to commands and events.

**Usage Example:**
```go
import "github.com/duongptryu/gox/eventbus"

cfg := eventbus.Config{Publisher: publisher, Subscriber: subscriber}
bus, _ := eventbus.NewBus(cfg)
bus.RegisterCommandHandler("MyCommand", &MyCommandHandler{})
bus.PublishCommand(ctx, &MyCommand{Data: "hello"})
```

---

### 5. `server/httpserver` — HTTP Server Utilities

- Standardized Gin router setup, middleware pipeline, and graceful shutdown.
- Health, readiness, and liveness endpoints out of the box.

**Usage Example:**
```go
import "github.com/duongptryu/gox/server/httpserver"

router := httpserver.SetupRouter(httpserver.RouterConfig{Environment: "prod", EnableCORS: true})
srv := httpserver.New(httpserver.Config{Host: "0.0.0.0", Port: 8080}, router)
srv.Start(context.Background())
```

---

### 6. `database` — Database Connection & Migration

- Utilities for SQL database connection pooling and migrations (using `sqlx` and `golang-migrate`).

---

### 7. `middleware` — HTTP Middleware

- Common middleware for logging, CORS, error handling, recovery, authentication, and request context.

---

### 8. `response` & `pagination`

- Helpers for standardized API responses and pagination handling.

---

## Installation

```bash
go get github.com/duongptryu/gox
```

## Requirements

- Go 1.18 or higher

## Contributing

Contributions, issues, and feature requests are welcome! Please open an issue or pull request to discuss improvements or new features.

## License

MIT

---

## Acknowledgements

- [Watermill](https://watermill.io/) for event bus
- [Gin](https://gin-gonic.com/) for HTTP server
- [Go's slog](https://pkg.go.dev/log/slog) for logging

---

## About

This library is maintained by Duong. Its main aim is to provide a public, reusable set of Go packages for use in your own projects, with a focus on custom, production-grade solutions. 