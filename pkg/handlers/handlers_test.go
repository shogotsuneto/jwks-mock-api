package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/shogotsuneto/jwks-mock-api/internal/keys"
	"github.com/shogotsuneto/jwks-mock-api/pkg/config"
)

// testServer represents a test server setup for testing handlers
type testServer struct {
	handler    *Handler
	keyManager *keys.Manager
	config     *config.Config
	router     *mux.Router
}

// newTestServer creates a new test server with test configuration
func newTestServer(t *testing.T) *testServer {
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

	keyManager := keys.NewManager()
	err := keyManager.GenerateKeys(cfg.Keys.KeyIDs)
	if err != nil {
		t.Fatalf("Failed to generate test keys: %v", err)
	}

	handler := New(cfg, keyManager)

	router := mux.NewRouter()
	router.Use(handler.CORS)
	router.HandleFunc("/.well-known/jwks.json", handler.JWKS).Methods("GET")
	router.HandleFunc("/generate-token", handler.GenerateToken).Methods("POST")
	router.HandleFunc("/generate-invalid-token", handler.GenerateInvalidToken).Methods("POST")
	router.HandleFunc("/introspect", handler.Introspect).Methods("POST")
	router.HandleFunc("/health", handler.Health).Methods("GET")
	router.HandleFunc("/keys", handler.Keys).Methods("GET")
	
	// Add OPTIONS handlers for CORS testing
	router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS middleware will handle this
	})

	return &testServer{
		handler:    handler,
		keyManager: keyManager,
		config:     cfg,
		router:     router,
	}
}

// TestHealthEndpoint tests the /health endpoint
func TestHealthEndpoint(t *testing.T) {
	ts := newTestServer(t)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	if response.Service != "jwt-dev-service" {
		t.Errorf("Expected service 'jwt-dev-service', got '%s'", response.Service)
	}

	if len(response.AvailableKeys) != 2 {
		t.Errorf("Expected 2 available keys, got %d", len(response.AvailableKeys))
	}

	expectedKeys := []string{"test-key-1", "test-key-2"}
	for i, key := range response.AvailableKeys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key '%s', got '%s'", expectedKeys[i], key)
		}
	}
}

// TestKeysEndpoint tests the /keys endpoint
func TestKeysEndpoint(t *testing.T) {
	ts := newTestServer(t)

	req := httptest.NewRequest("GET", "/keys", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response KeysResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.TotalKeys != 2 {
		t.Errorf("Expected 2 total keys, got %d", response.TotalKeys)
	}

	if len(response.AvailableKeys) != 2 {
		t.Errorf("Expected 2 available keys, got %d", len(response.AvailableKeys))
	}

	// Check key structure
	for i, key := range response.AvailableKeys {
		expectedKid := []string{"test-key-1", "test-key-2"}[i]
		
		if key["kid"] != expectedKid {
			t.Errorf("Expected kid '%s', got '%v'", expectedKid, key["kid"])
		}
		
		if key["alg"] != "RS256" {
			t.Errorf("Expected alg 'RS256', got '%v'", key["alg"])
		}
		
		if key["use"] != "sig" {
			t.Errorf("Expected use 'sig', got '%v'", key["use"])
		}
	}
}

// TestJWKSEndpoint tests the /.well-known/jwks.json endpoint
func TestJWKSEndpoint(t *testing.T) {
	ts := newTestServer(t)

	req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, expectedContentType) {
		t.Errorf("Expected content type to contain '%s', got '%s'", expectedContentType, contentType)
	}

	// Check cache control header
	if cacheControl := w.Header().Get("Cache-Control"); cacheControl != "public, max-age=3600" {
		t.Errorf("Expected cache control 'public, max-age=3600', got '%s'", cacheControl)
	}

	var jwks map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &jwks); err != nil {
		t.Fatalf("Failed to unmarshal JWKS response: %v", err)
	}

	// Check JWKS structure
	keys, ok := jwks["keys"].([]interface{})
	if !ok {
		t.Fatal("JWKS response should contain 'keys' array")
	}

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys in JWKS, got %d", len(keys))
	}

	// Validate each key structure
	for i, keyInterface := range keys {
		key, ok := keyInterface.(map[string]interface{})
		if !ok {
			t.Fatalf("Key %d is not a valid object", i)
		}

		// Check required fields
		requiredFields := []string{"kty", "use", "kid", "alg", "n", "e"}
		for _, field := range requiredFields {
			if _, exists := key[field]; !exists {
				t.Errorf("Key %d missing required field '%s'", i, field)
			}
		}

		// Check specific values
		if key["kty"] != "RSA" {
			t.Errorf("Expected kty 'RSA', got '%v'", key["kty"])
		}
		if key["use"] != "sig" {
			t.Errorf("Expected use 'sig', got '%v'", key["use"])
		}
		if key["alg"] != "RS256" {
			t.Errorf("Expected alg 'RS256', got '%v'", key["alg"])
		}
	}
}

// TestGenerateTokenEndpoint tests the /generate-token endpoint
func TestGenerateTokenEndpoint(t *testing.T) {
	ts := newTestServer(t)

	tests := []struct {
		name           string
		requestBody    TokenRequest
		expectedStatus int
		validateToken  bool
	}{
		{
			name: "Valid token request with custom claims",
			requestBody: TokenRequest{
				Claims: map[string]interface{}{
					"sub":   "user123",
					"email": "user@example.com",
					"roles": []string{"admin", "user"},
					"metadata": map[string]interface{}{
						"loginCount": 42,
						"department": "Engineering",
					},
				},
				ExpiresIn: intPtr(7200),
			},
			expectedStatus: http.StatusOK,
			validateToken:  true,
		},
		{
			name: "Valid token request with default claims",
			requestBody: TokenRequest{
				Claims: map[string]interface{}{},
			},
			expectedStatus: http.StatusOK,
			validateToken:  true,
		},
		{
			name: "Valid token request with default expiration",
			requestBody: TokenRequest{
				Claims: map[string]interface{}{
					"sub": "testuser",
				},
			},
			expectedStatus: http.StatusOK,
			validateToken:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/generate-token", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ts.router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateToken {
				var response TokenResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Validate response structure
				if response.AccessToken == "" {
					t.Error("Expected non-empty access_token")
				}

				if response.KeyID == "" {
					t.Error("Expected non-empty key_id")
				}

				expectedExpiresIn := 3600
				if tt.requestBody.ExpiresIn != nil {
					expectedExpiresIn = *tt.requestBody.ExpiresIn
				}
				if response.ExpiresIn != expectedExpiresIn {
					t.Errorf("Expected expires_in %d, got %d", expectedExpiresIn, response.ExpiresIn)
				}

				// Validate the actual JWT token
				validateGeneratedToken(t, ts, response.AccessToken, tt.requestBody.Claims)
			}
		})
	}
}

// TestGenerateTokenInvalidJSON tests the /generate-token endpoint with invalid JSON
func TestGenerateTokenInvalidJSON(t *testing.T) {
	ts := newTestServer(t)

	req := httptest.NewRequest("POST", "/generate-token", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGenerateInvalidTokenEndpoint tests the /generate-invalid-token endpoint
func TestGenerateInvalidTokenEndpoint(t *testing.T) {
	ts := newTestServer(t)

	requestBody := TokenRequest{
		Claims: map[string]interface{}{
			"sub":   "invalid-user",
			"email": "invalid@example.com",
		},
		ExpiresIn: intPtr(1800),
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/generate-invalid-token", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate response structure
	if response.AccessToken == "" {
		t.Error("Expected non-empty access_token")
	}

	if response.KeyID == "" {
		t.Error("Expected non-empty key_id")
	}

	if response.ExpiresIn != 1800 {
		t.Errorf("Expected expires_in 1800, got %d", response.ExpiresIn)
	}

	// The token should be invalid when validated against our JWKS
	// This is because it's signed with a different private key
	validateInvalidToken(t, ts, response.AccessToken)
}

// TestIntrospectEndpoint tests the /introspect endpoint
func TestIntrospectEndpoint(t *testing.T) {
	ts := newTestServer(t)

	// First generate a valid token
	validTokenClaims := map[string]interface{}{
		"sub":   "user123",
		"email": "user@example.com",
		"roles": []string{"admin"},
	}
	validToken := generateTestToken(t, ts, validTokenClaims, 3600)

	// Generate an invalid token
	invalidToken := generateTestInvalidToken(t, ts, validTokenClaims, 3600)

	tests := []struct {
		name           string
		token          string
		expectedActive bool
		expectedStatus int
	}{
		{
			name:           "Valid token introspection",
			token:          validToken,
			expectedActive: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid token introspection",
			token:          invalidToken,
			expectedActive: false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty token",
			token:          "",
			expectedActive: false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Malformed token",
			token:          "invalid.jwt.token",
			expectedActive: false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formData := url.Values{}
			formData.Set("token", tt.token)

			req := httptest.NewRequest("POST", "/introspect", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			ts.router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response IntrospectionResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Active != tt.expectedActive {
				t.Errorf("Expected active %v, got %v", tt.expectedActive, response.Active)
			}

			if tt.expectedActive {
				// For active tokens, validate additional fields
				if response.TokenType != "Bearer" {
					t.Errorf("Expected token_type 'Bearer', got '%s'", response.TokenType)
				}

				if response.Sub != "user123" {
					t.Errorf("Expected sub 'user123', got '%s'", response.Sub)
				}

				if response.Username != "user123" {
					t.Errorf("Expected username 'user123', got '%s'", response.Username)
				}

				if response.Iss != ts.config.JWT.Issuer {
					t.Errorf("Expected iss '%s', got '%s'", ts.config.JWT.Issuer, response.Iss)
				}

				if response.Aud != ts.config.JWT.Audience {
					t.Errorf("Expected aud '%s', got '%s'", ts.config.JWT.Audience, response.Aud)
				}
			}
		})
	}
}

// TestIntrospectInvalidForm tests the /introspect endpoint with invalid form data
func TestIntrospectInvalidForm(t *testing.T) {
	ts := newTestServer(t)

	req := httptest.NewRequest("POST", "/introspect", strings.NewReader("%invalid form data"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response IntrospectionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Active != false {
		t.Error("Expected active false for invalid form data")
	}
}

// TestCORSHeaders tests that CORS headers are properly set
func TestCORSHeaders(t *testing.T) {
	ts := newTestServer(t)

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/health", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for OPTIONS, got %d", http.StatusOK, w.Code)
	}

	// Check CORS headers
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	for header, expectedValue := range expectedHeaders {
		if actualValue := w.Header().Get(header); actualValue != expectedValue {
			t.Errorf("Expected %s header '%s', got '%s'", header, expectedValue, actualValue)
		}
	}

	// Test that CORS headers are set on regular requests too
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
	}
}

// Helper functions

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}

// validateGeneratedToken validates that a generated token is properly formed and contains expected claims
func validateGeneratedToken(t *testing.T, ts *testServer, tokenString string, expectedClaims map[string]interface{}) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the kid from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing key ID in token header")
		}

		// Find the corresponding key
		keyPair, err := ts.keyManager.GetKeyByID(kid)
		if err != nil {
			return nil, fmt.Errorf("key not found for kid: %s", kid)
		}

		return keyPair.PublicKey, nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if !token.Valid {
		t.Fatal("Token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims from token")
	}

	// Validate standard claims
	if claims["iss"] != ts.config.JWT.Issuer {
		t.Errorf("Expected iss '%s', got '%v'", ts.config.JWT.Issuer, claims["iss"])
	}

	if claims["aud"] != ts.config.JWT.Audience {
		t.Errorf("Expected aud '%s', got '%v'", ts.config.JWT.Audience, claims["aud"])
	}

	// Check that iat and exp are present and reasonable
	if _, ok := claims["iat"]; !ok {
		t.Error("Token missing 'iat' claim")
	}

	if _, ok := claims["exp"]; !ok {
		t.Error("Token missing 'exp' claim")
	}

	// Validate custom claims
	for key, expectedValue := range expectedClaims {
		if actualValue, exists := claims[key]; !exists {
			t.Errorf("Expected claim '%s' not found in token", key)
		} else {
			// For complex types, we'll do a basic comparison
			if !compareClaimValues(expectedValue, actualValue) {
				t.Errorf("Expected claim '%s' to be '%v', got '%v'", key, expectedValue, actualValue)
			}
		}
	}
}

// validateInvalidToken validates that a token should fail validation
func validateInvalidToken(t *testing.T, ts *testServer, tokenString string) {
	// Parse the token - this should fail because it's signed with a different key
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Get the kid from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing key ID in token header")
		}

		// Find the corresponding key
		keyPair, err := ts.keyManager.GetKeyByID(kid)
		if err != nil {
			return nil, fmt.Errorf("key not found for kid: %s", kid)
		}

		return keyPair.PublicKey, nil
	})

	// The token should fail validation
	if err == nil {
		t.Error("Expected invalid token to fail validation, but it passed")
	}
}

// generateTestToken generates a valid token for testing
func generateTestToken(t *testing.T, ts *testServer, claims map[string]interface{}, expiresInSeconds int) string {
	requestBody := TokenRequest{
		Claims:    claims,
		ExpiresIn: &expiresInSeconds,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/generate-token", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to generate test token: status %d", w.Code)
	}

	var response TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal token response: %v", err)
	}

	return response.AccessToken
}

// generateTestInvalidToken generates an invalid token for testing
func generateTestInvalidToken(t *testing.T, ts *testServer, claims map[string]interface{}, expiresInSeconds int) string {
	requestBody := TokenRequest{
		Claims:    claims,
		ExpiresIn: &expiresInSeconds,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/generate-invalid-token", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to generate test invalid token: status %d", w.Code)
	}

	var response TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal token response: %v", err)
	}

	return response.AccessToken
}

// compareClaimValues compares two claim values, handling different types appropriately
func compareClaimValues(expected, actual interface{}) bool {
	// For simple types, direct comparison
	switch exp := expected.(type) {
	case string, int, float64, bool:
		return expected == actual
	case []string:
		// Handle string arrays
		if actualSlice, ok := actual.([]interface{}); ok {
			if len(exp) != len(actualSlice) {
				return false
			}
			for i, v := range exp {
				if actualStr, ok := actualSlice[i].(string); !ok || actualStr != v {
					return false
				}
			}
			return true
		}
		return false
	default:
		// For complex types, convert to JSON and compare
		expectedJSON, _ := json.Marshal(expected)
		actualJSON, _ := json.Marshal(actual)
		return string(expectedJSON) == string(actualJSON)
	}
}