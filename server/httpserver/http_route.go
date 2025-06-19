package httpserver

import (
	"net/http"
	"time"

	"github.com/duongptryu/gox/server/middleware"

	"github.com/gin-gonic/gin"
)

// RouterConfig holds router configuration options
type RouterConfig struct {
	Environment string
	EnableCORS  bool
	EnableAuth  bool
}

// SetupRouter creates and configures a Gin router with standard middleware
func SetupRouter(config RouterConfig) *gin.Engine {
	// Set Gin mode based on environment
	if config.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router
	router := gin.New()

	// Add core middleware
	setupCoreMiddleware(router, config)

	// Add health endpoints
	setupHealthEndpoints(router)

	return router
}

// setupCoreMiddleware adds the standard middleware pipeline
func setupCoreMiddleware(router *gin.Engine, config RouterConfig) {
	// Recovery middleware
	router.Use(middleware.Recovery())

	// Request context middleware
	router.Use(middleware.RequestContext())

	// Request logging
	router.Use(middleware.RequestLogger())

	// CORS middleware (if enabled)
	if config.EnableCORS {
		router.Use(middleware.CORS())
	}

	// Error handling middleware (should be last)
	router.Use(middleware.ErrorHandler())
}

// setupHealthEndpoints adds standard health check endpoints
func setupHealthEndpoints(router *gin.Engine) {
	// Basic health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "tixgo-api",
		})
	})

	// Readiness check
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	// Liveness check
	router.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	})
}

// AddAPIGroup creates a versioned API route group
func AddAPIGroup(router *gin.Engine, version string) *gin.RouterGroup {
	return router.Group("/api/" + version)
}

// AddProtectedGroup creates a protected route group
// Note: Authentication middleware should be added when you have the JWT service available
func AddProtectedGroup(group *gin.RouterGroup, path string) *gin.RouterGroup {
	return group.Group(path)
}
