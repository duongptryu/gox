# Server Package

This package provides HTTP server utilities following the [Wild Workouts Go DDD example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/tree/master/internal/common/server) patterns for consistent server setup, middleware configuration, and graceful shutdown.

## Features

### Core Components

- **HTTP Server**: Configurable HTTP server with timeouts and graceful shutdown
- **Router Setup**: Standardized Gin router configuration with middleware pipeline
- **Graceful Shutdown**: Signal handling for clean server termination
- **Health Endpoints**: Standard health, readiness, and liveness checks
- **Configuration**: Type-safe configuration utilities

### Wild Workouts Compliance

This implementation follows the Wild Workouts common server patterns:

- ✅ **Centralized server utilities**
- ✅ **Graceful shutdown handling**
- ✅ **Standardized middleware pipeline**
- ✅ **Health check endpoints**
- ✅ **Configuration abstraction**

## Usage

### Basic Server Setup

```go
package main

import (
    "context"
    "tixgo/config"
    "tixgo/shared/server"
    "tixgo/shared/logger"
)

func main() {
    ctx := context.Background()
    
    // Load application configuration
    appConfig, err := config.LoadConfig()
    if err != nil {
        logger.Fatal(ctx, "Failed to load config", logger.F("error", err))
    }
    
    // Setup router with middleware
    routerConfig := server.RouterConfigFromAppConfig(appConfig)
    router := server.SetupRouter(routerConfig)
    
    // Add your API routes
    v1 := server.AddAPIGroup(router, "v1")
    v1.GET("/users", getUsersHandler)
    
    // Create server
    serverConfig := server.ConfigFromAppConfig(appConfig)
    srv := server.New(serverConfig, router)
    
    // Start server (blocks until shutdown)
    if err := srv.Start(ctx); err != nil {
        logger.Fatal(ctx, "Server failed", logger.F("error", err))
    }
}
```

### Router Configuration

```go
// Manual router configuration
routerConfig := server.RouterConfig{
    Environment: "prod",
    EnableCORS:  true,
    EnableAuth:  true,
}

router := server.SetupRouter(routerConfig)
```

### Adding Protected Routes

```go
// Create API group
v1 := server.AddAPIGroup(router, "v1")

// Add public routes
v1.POST("/login", loginHandler)
v1.POST("/register", registerHandler)

// Add protected routes (requires JWT service)
protected := server.AddProtectedGroup(v1, "/users")
protected.Use(middleware.RequireAuth(jwtService)) // Add auth middleware
protected.GET("/profile", getProfileHandler)
```

### Custom Server Configuration

```go
serverConfig := server.Config{
    Host:         "0.0.0.0",
    Port:         8080,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}

srv := server.New(serverConfig, router)
```

## Configuration

### Server Config

```go
type Config struct {
    Host         string        // Server host (e.g., "localhost", "0.0.0.0")
    Port         int           // Server port (e.g., 8080)
    ReadTimeout  time.Duration // HTTP read timeout
    WriteTimeout time.Duration // HTTP write timeout
    IdleTimeout  time.Duration // HTTP idle timeout
}
```

### Router Config

```go
type RouterConfig struct {
    Environment string // Environment ("dev", "stg", "prod")
    EnableCORS  bool   // Enable CORS middleware
    EnableAuth  bool   // Enable auth-related features
}
```

## Health Endpoints

The server automatically adds health check endpoints:

### Available Endpoints

- **`GET /health`** - Basic health check
  ```json
  {
    "status": "ok",
    "timestamp": 1641234567,
    "service": "tixgo-api"
  }
  ```

- **`GET /ready`** - Readiness check (service ready to handle requests)
  ```json
  {
    "status": "ready"
  }
  ```

- **`GET /live`** - Liveness check (service is alive)
  ```json
  {
    "status": "alive"
  }
  ```

## Middleware Pipeline

The standard middleware pipeline includes:

1. **Request Context** - Adds request/operation IDs
2. **Request Logger** - Structured HTTP request logging
3. **Recovery** - Panic recovery with error logging
4. **CORS** - Cross-origin request support (if enabled)
5. **Error Handler** - Centralized error handling

## Graceful Shutdown

The server handles graceful shutdown automatically:

- Listens for `SIGINT` and `SIGTERM` signals
- Gives active requests 30 seconds to complete
- Logs shutdown progress
- Returns error if forced shutdown occurs

### Shutdown Behavior

```bash
# Send interrupt signal
Ctrl+C

# Server logs:
# "Received shutdown signal, shutting down gracefully..."
# "Shutting down HTTP server..."
# "HTTP server shut down gracefully"
```

## Integration with Main Application

Update your `cmd/api_server/main.go` to use the server package:

```go
// Old approach (manual setup)
func setupHTTPServer(ctx context.Context, cfg *config.AppConfig, modules *Modules) *http.Server {
    // ... manual router and server setup
}

// New approach (using server package)
func setupHTTPServer(ctx context.Context, cfg *config.AppConfig, modules *Modules) *server.Server {
    // Setup router
    routerConfig := server.RouterConfigFromAppConfig(cfg)
    router := server.SetupRouter(routerConfig)
    
    // Register module routes
    registerRoutes(router, modules)
    
    // Create server
    serverConfig := server.ConfigFromAppConfig(cfg)
    return server.New(serverConfig, router)
}

func startServer(ctx context.Context, srv *server.Server) {
    if err := srv.Start(ctx); err != nil {
        logger.Fatal(ctx, "Server failed", logger.F("error", err))
    }
}
```

## Error Handling

The server integrates with the existing error handling system:

- Uses `shared/syserr` for structured errors
- Centralized error handling via middleware
- Structured error responses
- Automatic error logging

## Production Considerations

- **Timeouts**: Configure appropriate read/write/idle timeouts
- **Host Binding**: Use `"0.0.0.0"` for container deployments
- **Health Checks**: Monitor `/health`, `/ready`, `/live` endpoints
- **Graceful Shutdown**: Ensure proper signal handling in orchestrators
- **Logging**: Monitor server startup and shutdown logs

## References

- [Wild Workouts Server Package](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/tree/master/internal/common/server)
- [Gin Framework](https://gin-gonic.com/)
- [Go HTTP Server](https://pkg.go.dev/net/http#Server)
- [Graceful Shutdown Patterns](https://threedots.tech/) 