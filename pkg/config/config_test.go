package config

import (
	"os"
	"testing"
)

// TestLoad tests the config loading functionality
func TestLoad(t *testing.T) {
	// Clear environment variables before tests
	originalEnvVars := map[string]string{
		"PORT":        os.Getenv("PORT"),
		"HOST":        os.Getenv("HOST"),
		"JWT_ISSUER":  os.Getenv("JWT_ISSUER"),
		"JWT_AUDIENCE": os.Getenv("JWT_AUDIENCE"),
		"KEY_COUNT":   os.Getenv("KEY_COUNT"),
		"KEY_IDS":     os.Getenv("KEY_IDS"),
	}

	// Clean environment
	for key := range originalEnvVars {
		os.Unsetenv(key)
	}

	// Restore environment after tests
	defer func() {
		for key, value := range originalEnvVars {
			if value != "" {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("Load default config", func(t *testing.T) {
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Check default values
		if cfg.Server.Port != 3000 {
			t.Errorf("Expected default port 3000, got %d", cfg.Server.Port)
		}

		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("Expected default host '0.0.0.0', got '%s'", cfg.Server.Host)
		}

		if cfg.JWT.Issuer != "http://localhost:3000" {
			t.Errorf("Expected default issuer 'http://localhost:3000', got '%s'", cfg.JWT.Issuer)
		}

		if cfg.JWT.Audience != "dev-api" {
			t.Errorf("Expected default audience 'dev-api', got '%s'", cfg.JWT.Audience)
		}

		if cfg.Keys.Count != 2 {
			t.Errorf("Expected default key count 2, got %d", cfg.Keys.Count)
		}

		expectedKeyIDs := []string{"key-1", "key-2"}
		if len(cfg.Keys.KeyIDs) != len(expectedKeyIDs) {
			t.Errorf("Expected %d key IDs, got %d", len(expectedKeyIDs), len(cfg.Keys.KeyIDs))
		}

		for i, expectedID := range expectedKeyIDs {
			if i >= len(cfg.Keys.KeyIDs) || cfg.Keys.KeyIDs[i] != expectedID {
				t.Errorf("Expected key ID '%s' at index %d, got '%s'", expectedID, i, cfg.Keys.KeyIDs[i])
			}
		}
	})

	t.Run("Load config with environment variables", func(t *testing.T) {
		// Set environment variables
		os.Setenv("PORT", "8080")
		os.Setenv("HOST", "127.0.0.1")
		os.Setenv("JWT_ISSUER", "https://test.example.com")
		os.Setenv("JWT_AUDIENCE", "test-audience")
		os.Setenv("KEY_COUNT", "3")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Check environment overrides
		if cfg.Server.Port != 8080 {
			t.Errorf("Expected port 8080 from env, got %d", cfg.Server.Port)
		}

		if cfg.Server.Host != "127.0.0.1" {
			t.Errorf("Expected host '127.0.0.1' from env, got '%s'", cfg.Server.Host)
		}

		if cfg.JWT.Issuer != "https://test.example.com" {
			t.Errorf("Expected issuer 'https://test.example.com' from env, got '%s'", cfg.JWT.Issuer)
		}

		if cfg.JWT.Audience != "test-audience" {
			t.Errorf("Expected audience 'test-audience' from env, got '%s'", cfg.JWT.Audience)
		}

		if cfg.Keys.Count != 3 {
			t.Errorf("Expected key count 3 from env, got %d", cfg.Keys.Count)
		}

		// Key IDs should be generated based on count
		expectedKeyIDs := []string{"key-1", "key-2", "key-3"}
		if len(cfg.Keys.KeyIDs) != len(expectedKeyIDs) {
			t.Errorf("Expected %d key IDs, got %d", len(expectedKeyIDs), len(cfg.Keys.KeyIDs))
		}

		for i, expectedID := range expectedKeyIDs {
			if i >= len(cfg.Keys.KeyIDs) || cfg.Keys.KeyIDs[i] != expectedID {
				t.Errorf("Expected key ID '%s' at index %d, got '%s'", expectedID, i, cfg.Keys.KeyIDs[i])
			}
		}

		// Clean up for next test
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
		os.Unsetenv("JWT_ISSUER")
		os.Unsetenv("JWT_AUDIENCE")
		os.Unsetenv("KEY_COUNT")
	})

	t.Run("Load config with KEY_IDS environment variable", func(t *testing.T) {
		// Set KEY_IDS environment variable
		os.Setenv("KEY_IDS", "custom-key-1,custom-key-2,custom-key-3")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Check that KEY_IDS overrides KEY_COUNT
		expectedKeyIDs := []string{"custom-key-1", "custom-key-2", "custom-key-3"}
		if len(cfg.Keys.KeyIDs) != len(expectedKeyIDs) {
			t.Errorf("Expected %d key IDs, got %d", len(expectedKeyIDs), len(cfg.Keys.KeyIDs))
		}

		for i, expectedID := range expectedKeyIDs {
			if i >= len(cfg.Keys.KeyIDs) || cfg.Keys.KeyIDs[i] != expectedID {
				t.Errorf("Expected key ID '%s' at index %d, got '%s'", expectedID, i, cfg.Keys.KeyIDs[i])
			}
		}

		if cfg.Keys.Count != 3 {
			t.Errorf("Expected count to be updated to 3, got %d", cfg.Keys.Count)
		}

		// Clean up
		os.Unsetenv("KEY_IDS")
	})

	t.Run("Load config with KEY_IDS containing spaces", func(t *testing.T) {
		// Set KEY_IDS with spaces
		os.Setenv("KEY_IDS", " spaced-key-1 , spaced-key-2 , spaced-key-3 ")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Check that spaces are trimmed
		expectedKeyIDs := []string{"spaced-key-1", "spaced-key-2", "spaced-key-3"}
		if len(cfg.Keys.KeyIDs) != len(expectedKeyIDs) {
			t.Errorf("Expected %d key IDs, got %d", len(expectedKeyIDs), len(cfg.Keys.KeyIDs))
		}

		for i, expectedID := range expectedKeyIDs {
			if i >= len(cfg.Keys.KeyIDs) || cfg.Keys.KeyIDs[i] != expectedID {
				t.Errorf("Expected key ID '%s' at index %d, got '%s'", expectedID, i, cfg.Keys.KeyIDs[i])
			}
		}

		// Clean up
		os.Unsetenv("KEY_IDS")
	})

	t.Run("Load config with invalid PORT", func(t *testing.T) {
		// Set invalid PORT
		os.Setenv("PORT", "invalid-port")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Should fallback to default port
		if cfg.Server.Port != 3000 {
			t.Errorf("Expected default port 3000 for invalid PORT, got %d", cfg.Server.Port)
		}

		// Clean up
		os.Unsetenv("PORT")
	})

	t.Run("Load config with invalid KEY_COUNT", func(t *testing.T) {
		// Set invalid KEY_COUNT
		os.Setenv("KEY_COUNT", "invalid-count")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Should fallback to default count and key IDs
		if cfg.Keys.Count != 2 {
			t.Errorf("Expected default count 2 for invalid KEY_COUNT, got %d", cfg.Keys.Count)
		}

		expectedKeyIDs := []string{"key-1", "key-2"}
		if len(cfg.Keys.KeyIDs) != len(expectedKeyIDs) {
			t.Errorf("Expected %d key IDs for invalid KEY_COUNT, got %d", len(expectedKeyIDs), len(cfg.Keys.KeyIDs))
		}

		// Clean up
		os.Unsetenv("KEY_COUNT")
	})

	t.Run("Load config with zero KEY_COUNT", func(t *testing.T) {
		// Set KEY_COUNT to 0
		os.Setenv("KEY_COUNT", "0")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Should fallback to default count and key IDs
		if cfg.Keys.Count != 2 {
			t.Errorf("Expected default count 2 for zero KEY_COUNT, got %d", cfg.Keys.Count)
		}

		// Clean up
		os.Unsetenv("KEY_COUNT")
	})
}

// TestLoadFromFile tests loading configuration from a file
func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	configContent := `
server:
  port: 9090
  host: "custom-host"
jwt:
  issuer: "https://file.example.com"
  audience: "file-audience"
keys:
  count: 4
  key_ids: ["file-key-1", "file-key-2", "file-key-3", "file-key-4"]
`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	// Clear environment variables to ensure file takes precedence over defaults
	originalEnvVars := map[string]string{
		"PORT":         os.Getenv("PORT"),
		"HOST":         os.Getenv("HOST"),
		"JWT_ISSUER":   os.Getenv("JWT_ISSUER"),
		"JWT_AUDIENCE": os.Getenv("JWT_AUDIENCE"),
		"KEY_COUNT":    os.Getenv("KEY_COUNT"),
		"KEY_IDS":      os.Getenv("KEY_IDS"),
	}

	for key := range originalEnvVars {
		os.Unsetenv(key)
	}

	defer func() {
		for key, value := range originalEnvVars {
			if value != "" {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("Load config from file", func(t *testing.T) {
		cfg, err := Load(tmpFile.Name())
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Check file values
		if cfg.Server.Port != 9090 {
			t.Errorf("Expected port 9090 from file, got %d", cfg.Server.Port)
		}

		if cfg.Server.Host != "custom-host" {
			t.Errorf("Expected host 'custom-host' from file, got '%s'", cfg.Server.Host)
		}

		if cfg.JWT.Issuer != "https://file.example.com" {
			t.Errorf("Expected issuer 'https://file.example.com' from file, got '%s'", cfg.JWT.Issuer)
		}

		if cfg.JWT.Audience != "file-audience" {
			t.Errorf("Expected audience 'file-audience' from file, got '%s'", cfg.JWT.Audience)
		}

		if cfg.Keys.Count != 4 {
			t.Errorf("Expected key count 4 from file, got %d", cfg.Keys.Count)
		}

		expectedKeyIDs := []string{"file-key-1", "file-key-2", "file-key-3", "file-key-4"}
		if len(cfg.Keys.KeyIDs) != len(expectedKeyIDs) {
			t.Errorf("Expected %d key IDs from file, got %d", len(expectedKeyIDs), len(cfg.Keys.KeyIDs))
		}

		for i, expectedID := range expectedKeyIDs {
			if i >= len(cfg.Keys.KeyIDs) || cfg.Keys.KeyIDs[i] != expectedID {
				t.Errorf("Expected key ID '%s' at index %d from file, got '%s'", expectedID, i, cfg.Keys.KeyIDs[i])
			}
		}
	})

	t.Run("Environment variables override file config", func(t *testing.T) {
		// Set environment variables
		os.Setenv("PORT", "7777")
		os.Setenv("JWT_ISSUER", "https://env.example.com")

		cfg, err := Load(tmpFile.Name())
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Environment should override file
		if cfg.Server.Port != 7777 {
			t.Errorf("Expected port 7777 from env override, got %d", cfg.Server.Port)
		}

		if cfg.JWT.Issuer != "https://env.example.com" {
			t.Errorf("Expected issuer 'https://env.example.com' from env override, got '%s'", cfg.JWT.Issuer)
		}

		// File values should still be used where no env override
		if cfg.Server.Host != "custom-host" {
			t.Errorf("Expected host 'custom-host' from file, got '%s'", cfg.Server.Host)
		}

		if cfg.JWT.Audience != "file-audience" {
			t.Errorf("Expected audience 'file-audience' from file, got '%s'", cfg.JWT.Audience)
		}

		// Clean up
		os.Unsetenv("PORT")
		os.Unsetenv("JWT_ISSUER")
	})
}

// TestLoadFromNonExistentFile tests loading from a non-existent file
func TestLoadFromNonExistentFile(t *testing.T) {
	_, err := Load("/non/existent/config.yaml")
	if err == nil {
		t.Error("Expected error for non-existent config file, got nil")
	}
}

// TestLoadFromInvalidFile tests loading from an invalid YAML file
func TestLoadFromInvalidFile(t *testing.T) {
	// Create a temporary invalid YAML file
	invalidContent := `
invalid: yaml: content
  - missing
    proper: structure
`

	tmpFile, err := os.CreateTemp("", "test-invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	_, err = Load(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML file, got nil")
	}
}

// TestLoadFromEmptyFile tests loading from an empty file
func TestLoadFromEmptyFile(t *testing.T) {
	// Create an empty temporary file
	tmpFile, err := os.CreateTemp("", "test-empty-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Clear environment variables
	originalEnvVars := map[string]string{
		"PORT":         os.Getenv("PORT"),
		"HOST":         os.Getenv("HOST"),
		"JWT_ISSUER":   os.Getenv("JWT_ISSUER"),
		"JWT_AUDIENCE": os.Getenv("JWT_AUDIENCE"),
		"KEY_COUNT":    os.Getenv("KEY_COUNT"),
		"KEY_IDS":      os.Getenv("KEY_IDS"),
	}

	for key := range originalEnvVars {
		os.Unsetenv(key)
	}

	defer func() {
		for key, value := range originalEnvVars {
			if value != "" {
				os.Setenv(key, value)
			}
		}
	}()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load() failed for empty file: %v", err)
	}

	// Should still get default values
	if cfg.Server.Port != 3000 {
		t.Errorf("Expected default port 3000 for empty file, got %d", cfg.Server.Port)
	}

	if cfg.JWT.Issuer != "http://localhost:3000" {
		t.Errorf("Expected default issuer for empty file, got '%s'", cfg.JWT.Issuer)
	}
}