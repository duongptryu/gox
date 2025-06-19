package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/duongptryu/gox/logger"
)

// Config holds the server configuration
type Config struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Server wraps the HTTP server with additional functionality
type Server struct {
	httpServer *http.Server
	config     Config
}

// New creates a new server instance with the given configuration and handler
func New(config Config, handler http.Handler) *Server {
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      handler,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		config:     config,
	}
}

// Start starts the HTTP server and handles graceful shutdown
func (s *Server) Start(ctx context.Context) error {
	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Info(ctx, "Starting HTTP server",
			logger.F("address", s.httpServer.Addr),
			logger.F("read_timeout", s.config.ReadTimeout),
			logger.F("write_timeout", s.config.WriteTimeout))

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to start server: %w", err)
		}
	}()

	logger.Info(ctx, "HTTP server started successfully", logger.F("address", s.httpServer.Addr))

	// Wait for interrupt signal or error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info(ctx, "Received shutdown signal, shutting down gracefully...")
		return s.shutdown(ctx)
	case err := <-errChan:
		return err
	}
}

// shutdown performs graceful shutdown of the HTTP server
func (s *Server) shutdown(ctx context.Context) error {
	// Create a deadline for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info(ctx, "Shutting down HTTP server...")

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "Server forced to shutdown", logger.F("error", err))
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	logger.Info(ctx, "HTTP server shut down gracefully")
	return nil
}

// Addr returns the server address
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// Handler returns the server handler
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}
