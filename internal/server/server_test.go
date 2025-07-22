package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shogotsuneto/jwks-mock-api/pkg/config"
)

// TestNew tests server creation
func TestNew(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 3000,
			Host: "localhost",
		},
		JWT: config.JWTConfig{
			Issuer:   "http://localhost:3000",
			Audience: "test-api",
		},
		Keys: config.KeysConfig{
			Count:  2,
			KeyIDs: []string{"test-key-1", "test-key-2"},
		},
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if srv == nil {
		t.Fatal("New() returned nil server")
	}

	if srv.config != cfg {
		t.Error("Server config not set correctly")
	}

	if srv.keyManager == nil {
		t.Error("Key manager not initialized")
	}

	if srv.handler == nil {
		t.Error("Handler not initialized")
	}
}

// TestNewWithInvalidKeys tests server creation with invalid key configuration
func TestNewWithInvalidKeys(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 3000,
			Host: "localhost",
		},
		JWT: config.JWTConfig{
			Issuer:   "http://localhost:3000",
			Audience: "test-api",
		},
		Keys: config.KeysConfig{
			Count:  0,
			KeyIDs: []string{}, // Empty key IDs should work
		},
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed with empty keys: %v", err)
	}

	if srv == nil {
		t.Fatal("New() returned nil server with empty keys")
	}
}

// TestSetupRoutes tests route configuration
func TestSetupRoutes(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 3000,
			Host: "localhost",
		},
		JWT: config.JWTConfig{
			Issuer:   "http://localhost:3000",
			Audience: "test-api",
		},
		Keys: config.KeysConfig{
			Count:  1,
			KeyIDs: []string{"test-key"},
		},
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	router := srv.setupRoutes()
	if router == nil {
		t.Fatal("setupRoutes() returned nil router")
	}

	// Test that all expected routes are accessible
	testCases := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/.well-known/jwks.json", http.StatusOK},
		{"GET", "/health", http.StatusOK},
		{"GET", "/keys", http.StatusOK},
		{"POST", "/generate-token", http.StatusBadRequest}, // No body
		{"POST", "/generate-invalid-token", http.StatusBadRequest}, // No body
		{"POST", "/introspect", http.StatusOK}, // Should return inactive for empty token
	}

	for _, tc := range testCases {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.method == "POST" && tc.path == "/introspect" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else if tc.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.status {
				t.Errorf("Expected status %d for %s %s, got %d", tc.status, tc.method, tc.path, w.Code)
			}
		})
	}
}

// TestServerIntegration tests basic server functionality without actually starting it
func TestServerIntegration(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 0, // Use random port
			Host: "localhost",
		},
		JWT: config.JWTConfig{
			Issuer:   "http://localhost:3000",
			Audience: "test-api",
		},
		Keys: config.KeysConfig{
			Count:  1,
			KeyIDs: []string{"integration-test-key"},
		},
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test that we can set up the HTTP server without starting it
	router := srv.setupRoutes()
	
	// Create an HTTP server manually for testing
	testServer := &http.Server{
		Addr:         "localhost:0",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if testServer.Handler == nil {
		t.Error("Server handler not set correctly")
	}

	if testServer.ReadTimeout != 15*time.Second {
		t.Error("Read timeout not set correctly")
	}

	if testServer.WriteTimeout != 15*time.Second {
		t.Error("Write timeout not set correctly")
	}

	if testServer.IdleTimeout != 60*time.Second {
		t.Error("Idle timeout not set correctly")
	}
}