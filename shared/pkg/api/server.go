// Package api provides REST API framework for all microservices.
package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
)

// Server represents an HTTP server.
type Server struct {
	server     *http.Server
	router     *Router
	logger     *logger.Logger
	config     Config
	shutdownWG sync.WaitGroup
	shutdownMu sync.Mutex
	isShutdown bool
}

// Config holds server configuration.
type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// Address returns the server address.
func (c *Config) Address() string {
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DefaultConfig returns default server configuration.
func DefaultConfig() Config {
	return Config{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
}

// New creates a new HTTP server.
func New(cfg Config, log *logger.Logger) *Server {
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 10 * time.Second
	}

	router := NewRouter()

	return &Server{
		router: router,
		logger: log,
		config: cfg,
		server: &http.Server{
			Addr:         cfg.Address(),
			Handler:      router,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}
}

// Router returns the server's router.
func (s *Server) Router() *Router {
	return s.router
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.logger.Info("Starting server",
		logger.String("address", s.config.Address()),
	)

	s.shutdownWG.Add(1)
	go func() {
		defer s.shutdownWG.Done()

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server error", logger.Err(err))
		}
	}()

	s.logger.Info("Server started", logger.String("address", s.config.Address()))
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	s.shutdownMu.Lock()
	if s.isShutdown {
		s.shutdownMu.Unlock()
		return nil
	}
	s.isShutdown = true
	s.shutdownMu.Unlock()

	s.logger.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Shutdown error", logger.Err(err))
		return err
	}

	s.shutdownWG.Wait()
	s.logger.Info("Server stopped")
	return nil
}

// WaitForShutdown waits for interrupt signal and shuts down the server.
func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Received shutdown signal")
	if err := s.Shutdown(); err != nil {
		s.logger.Error("Error during shutdown", logger.Err(err))
		os.Exit(1)
	}
}

// RegisterHandler registers a handler function.
func (s *Server) RegisterHandler(method, path string, handler http.HandlerFunc) {
	s.router.HandleFunc(method, path, handler)
}

// RegisterMiddleware registers global middleware.
func (s *Server) RegisterMiddleware(middleware ...MiddlewareFunc) {
	s.router.Use(middleware...)
}

// GET registers a GET handler.
func (s *Server) GET(path string, handler http.HandlerFunc) {
	s.router.GET(path, handler)
}

// POST registers a POST handler.
func (s *Server) POST(path string, handler http.HandlerFunc) {
	s.router.POST(path, handler)
}

// PUT registers a PUT handler.
func (s *Server) PUT(path string, handler http.HandlerFunc) {
	s.router.PUT(path, handler)
}

// DELETE registers a DELETE handler.
func (s *Server) DELETE(path string, handler http.HandlerFunc) {
	s.router.DELETE(path, handler)
}

// PATCH registers a PATCH handler.
func (s *Server) PATCH(path string, handler http.HandlerFunc) {
	s.router.PATCH(path, handler)
}
