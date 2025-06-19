# Context Package

This package provides utilities for managing context values in Go applications, enabling traceability and correlation across service boundaries without dependencies on specific web frameworks or middleware.

## Features

- **Operation ID Management**: Set and retrieve operation IDs in context for tracking operations across service calls
- **Request ID Management**: Handle request IDs for tracing individual requests through the system
- **User Context Support**: Store and retrieve user information in context for service-to-service calls
- **Type-Safe Context Keys**: Uses custom types for context keys to avoid collisions and ensure type safety
- **Framework Agnostic**: Pure Go context utilities that work with any framework or service

## Usage

### Operation ID Management

```go
import (
    "context"
    pkgContext "tixgo/internal/common/context"
)

// Set operation ID in context
ctx := context.Background()
ctx = pkgContext.WithOperationID(ctx, "operation-123")

// Retrieve operation ID from context
operationID := pkgContext.GetOperationID(ctx)
if operationID != "" {
    // Use operation ID for logging, tracing, etc.
    log.Printf("Processing operation: %s", operationID)
}
```

### Request ID Management

```go
// Set request ID in context
ctx = pkgContext.WithRequestID(ctx, "request-456")

// Retrieve request ID from context
requestID := pkgContext.GetRequestID(ctx)
```

### User Context Management

```go
// Set user information in context
ctx = pkgContext.WithUserID(ctx, "user-789")
ctx = pkgContext.WithUserType(ctx, "admin")

// Retrieve user information from context
userID := pkgContext.GetUserIDFromContext(ctx)
userType := pkgContext.GetUserTypeFromContext(ctx)
```

### Service-to-Service Calls

```go
func callDownstreamService(ctx context.Context, data interface{}) error {
    // Context automatically carries operation ID, request ID, etc.
    req, err := http.NewRequestWithContext(ctx, "POST", "/api/service", nil)
    if err != nil {
        return err
    }
    
    // Optionally add operation ID to headers for external services
    if operationID := pkgContext.GetOperationID(ctx); operationID != "" {
        req.Header.Set("X-Operation-ID", operationID)
    }
    
    // Make the call...
    return nil
}
```

## Integration with Middleware

This package works seamlessly with the middleware package (`internal/common/middleware`):

```go
// In middleware (Gin-specific)
func someMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Set operation ID from header
        operationID := c.GetHeader("X-Operation-ID")
        if operationID != "" {
            ctx := pkgContext.WithOperationID(c.Request.Context(), operationID)
            c.Request = c.Request.WithContext(ctx)
        }
        c.Next()
    }
}

// In handlers
func handler(c *gin.Context) {
    // Get operation ID from context
    operationID := pkgContext.GetOperationID(c.Request.Context())
    
    // Use in service calls
    err := someService.DoWork(c.Request.Context(), data)
    // ...
}
```

## Best Practices

### Context Propagation

- Always pass context through your application layers
- Use the context utilities to maintain traceability across service boundaries
- Don't store context values in structs; pass context as the first parameter

### Operation IDs

- Use operation IDs to track business operations across multiple services
- Generate meaningful operation IDs that help with debugging
- Propagate operation IDs in HTTP headers for external service calls

### Request IDs

- Use request IDs to track individual HTTP requests
- Generate unique request IDs for each incoming request
- Include request IDs in logs for easier debugging

### User Context

- Set user context early in the request lifecycle
- Use user context for authorization and audit logging
- Don't rely on user context for security decisions in external services

## Context Keys

The package uses typed context keys to avoid collisions:

```go
type contextKey string

const (
    OperationIDKey contextKey = "operationID"
    RequestIDKey   contextKey = "requestID"
    UserIDKey      contextKey = "userID"
    UserTypeKey    contextKey = "userType"
    AuthClaimsKey  contextKey = "authClaims"
)
```

## Logging Integration

This package integrates with the logger package to automatically include operation IDs in log entries:

```go
import (
    "tixgo/internal/common/logger"
    pkgContext "tixgo/internal/common/context"
)

func businessLogic(ctx context.Context) {
    // Operation ID will automatically be included in logs
    logger.Info(ctx, "Starting business operation")
    
    // Add operation ID manually if needed
    ctx = pkgContext.WithOperationID(ctx, "custom-operation-id")
    logger.Info(ctx, "Custom operation started")
}
```

## Architecture

This package is designed to be:

1. **Framework Agnostic**: Works with any Go application, not just web applications
2. **Dependency Free**: Only depends on standard library context package
3. **Type Safe**: Uses typed context keys to prevent collisions
4. **Performance Oriented**: Minimal overhead for context operations
5. **Integration Friendly**: Works seamlessly with logging, middleware, and service layers

## Migration from Old `ctx` Package

If you're migrating from the old `ctx` package:

```go
// Old import
// pkgCtx "tixgo/internal/common/ctx"

// New import
pkgContext "tixgo/internal/common/context"

// Function names remain the same
operationID := pkgContext.GetOperationID(ctx)
ctx = pkgContext.WithOperationID(ctx, operationID)
```

The API is backward compatible, so existing code should work with minimal changes. 