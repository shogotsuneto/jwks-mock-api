package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/shogotsuneto/jwks-mock-api/internal/keys"
	"github.com/shogotsuneto/jwks-mock-api/pkg/config"
	"github.com/shogotsuneto/jwks-mock-api/pkg/handlers"
	"github.com/shogotsuneto/jwks-mock-api/pkg/logger"
)

// Server represents the JWKS mock server
type Server struct {
	config     *config.Config
	keyManager *keys.Manager
	handler    *handlers.Handler
	server     *http.Server
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	// Initialize key manager
	keyManager := keys.NewManager()

	// Generate keys based on configuration
	keyIDs := cfg.InitialKeys.KeyIDs

	if err := keyManager.GenerateKeys(keyIDs); err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	// Initialize handlers
	handler := handlers.New(cfg, keyManager)

	server := &Server{
		config:     cfg,
		keyManager: keyManager,
		handler:    handler,
	}

	return server, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	router := s.setupRoutes()

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Infof("Environment variables:")
	logger.Infof("JWT_AUDIENCE: %s", s.config.JWT.Audience)
	logger.Infof("JWT_ISSUER: %s", s.config.JWT.Issuer)
	logger.Infof("PORT: %d", s.config.Server.Port)
	logger.Infof("HOST: %s", s.config.Server.Host)

	logger.Infof("Keys initialized successfully: %v", s.keyManager.GetAllKeyIDs())
	logger.Infof("JWT Dev Service starting on %s", s.server.Addr)
	logger.Infof("Available keys: %v", s.keyManager.GetAllKeyIDs())
	logger.Infof("JWKS endpoint: http://%s:%d/.well-known/jwks.json", s.config.Server.Host, s.config.Server.Port)
	logger.Infof("Generate token: POST http://%s:%d/generate-token", s.config.Server.Host, s.config.Server.Port)
	logger.Infof("Generate invalid token: POST http://%s:%d/generate-invalid-token", s.config.Server.Host, s.config.Server.Port)
	logger.Infof("Keys info: GET http://%s:%d/keys", s.config.Server.Host, s.config.Server.Port)
	logger.Infof("Add key: POST http://%s:%d/keys", s.config.Server.Host, s.config.Server.Port)
	logger.Infof("Remove key: DELETE http://%s:%d/keys/{kid}", s.config.Server.Host, s.config.Server.Port)

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	s.waitForShutdown()

	return nil
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply access logging middleware first
	router.Use(s.handler.AccessLog)
	
	// Apply CORS middleware
	router.Use(s.handler.CORS)

	// JWKS endpoint
	router.HandleFunc("/.well-known/jwks.json", s.handler.JWKS).Methods("GET", "OPTIONS")

	// Token generation endpoints
	router.HandleFunc("/generate-token", s.handler.GenerateToken).Methods("POST", "OPTIONS")
	router.HandleFunc("/generate-invalid-token", s.handler.GenerateInvalidToken).Methods("POST", "OPTIONS")

	// Token introspection endpoint (OAuth 2.0 RFC 7662)
	router.HandleFunc("/introspect", s.handler.Introspect).Methods("POST", "OPTIONS")

	// Health and info endpoints
	router.HandleFunc("/health", s.handler.Health).Methods("GET", "OPTIONS")
	router.HandleFunc("/keys", s.handler.Keys).Methods("GET", "OPTIONS")

	// Key management endpoints
	router.HandleFunc("/keys", s.handler.AddKey).Methods("POST", "OPTIONS")
	router.HandleFunc("/keys/{kid}", s.handler.RemoveKey).Methods("DELETE", "OPTIONS")

	return router
}

// waitForShutdown waits for interrupt signal and gracefully shuts down the server
func (s *Server) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Received shutdown signal. Gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("HTTP server shutdown completed")
}
