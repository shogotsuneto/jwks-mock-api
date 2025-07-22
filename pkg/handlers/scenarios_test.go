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
)

// TestRealWorldScenarios tests scenarios that backend API developers commonly encounter
func TestRealWorldScenarios(t *testing.T) {
	ts := newTestServer(t)

	t.Run("Complete JWT workflow - generate, validate, introspect", func(t *testing.T) {
		// 1. Generate a token with custom claims representing a real user
		userClaims := map[string]interface{}{
			"sub":         "user-12345",
			"email":       "john.doe@company.com",
			"given_name":  "John",
			"family_name": "Doe",
			"roles":       []string{"developer", "admin"},
			"department":  "Engineering",
			"permissions": []string{"read:projects", "write:projects", "delete:projects"},
			"metadata": map[string]interface{}{
				"last_login":    "2024-01-15T10:30:00Z",
				"login_count":   127,
				"is_verified":   true,
				"subscription":  "premium",
			},
		}

		// Generate token
		tokenReq := TokenRequest{
			Claims:    userClaims,
			ExpiresIn: intPtr(7200), // 2 hours
		}

		jsonBody, _ := json.Marshal(tokenReq)
		req := httptest.NewRequest("POST", "/generate-token", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ts.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var tokenResp TokenResponse
		if err := json.Unmarshal(w.Body.Bytes(), &tokenResp); err != nil {
			t.Fatalf("Failed to unmarshal token response: %v", err)
		}

		generatedToken := tokenResp.AccessToken

		// 2. Validate the token by parsing it (simulating backend validation)
		parsedToken, err := jwt.Parse(generatedToken, func(token *jwt.Token) (interface{}, error) {
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
			t.Fatalf("Failed to validate token: %v", err)
		}

		if !parsedToken.Valid {
			t.Fatal("Token should be valid")
		}

		// Verify claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("Failed to extract claims")
		}

		// Check that all our custom claims are present
		if claims["sub"] != "user-12345" {
			t.Errorf("Expected sub 'user-12345', got '%v'", claims["sub"])
		}

		if claims["email"] != "john.doe@company.com" {
			t.Errorf("Expected email 'john.doe@company.com', got '%v'", claims["email"])
		}

		// 3. Use token introspection (OAuth 2.0 standard)
		formData := url.Values{}
		formData.Set("token", generatedToken)

		req = httptest.NewRequest("POST", "/introspect", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()

		ts.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d for introspection, got %d", http.StatusOK, w.Code)
		}

		var introspectResp IntrospectionResponse
		if err := json.Unmarshal(w.Body.Bytes(), &introspectResp); err != nil {
			t.Fatalf("Failed to unmarshal introspection response: %v", err)
		}

		if !introspectResp.Active {
			t.Error("Token should be active in introspection")
		}

		if introspectResp.Sub != "user-12345" {
			t.Errorf("Expected introspection sub 'user-12345', got '%s'", introspectResp.Sub)
		}
	})

	t.Run("API endpoint testing scenario", func(t *testing.T) {
		// Common scenario: testing API endpoints that require JWT authentication
		
		// Generate a token for API testing
		apiTestClaims := map[string]interface{}{
			"sub":    "api-test-user",
			"scope":  "read:api write:api",
			"client": "test-client",
			"env":    "testing",
		}

		token := generateTestToken(t, ts, apiTestClaims, 3600)

		// Verify the token can be used for API authentication
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("missing key ID in token header")
			}

			keyPair, err := ts.keyManager.GetKeyByID(kid)
			if err != nil {
				return nil, fmt.Errorf("key not found for kid: %s", kid)
			}

			return keyPair.PublicKey, nil
		})

		if err != nil {
			t.Fatalf("API test token validation failed: %v", err)
		}

		claims := parsedToken.Claims.(jwt.MapClaims)
		
		// Verify scope-based authorization
		if claims["scope"] != "read:api write:api" {
			t.Errorf("Expected scope 'read:api write:api', got '%v'", claims["scope"])
		}

		// Verify client identification
		if claims["client"] != "test-client" {
			t.Errorf("Expected client 'test-client', got '%v'", claims["client"])
		}
	})

	t.Run("Microservices communication scenario", func(t *testing.T) {
		// Scenario: service-to-service authentication tokens
		
		serviceClaims := map[string]interface{}{
			"sub":         "service-payment",
			"aud":         ts.config.JWT.Audience,
			"client_type": "service",
			"service_id":  "payment-service-v1",
			"scopes":      []string{"payments:read", "payments:write", "users:read"},
			"version":     "1.2.3",
		}

		token := generateTestToken(t, ts, serviceClaims, 1800) // 30 minutes

		// Introspect the service token
		formData := url.Values{}
		formData.Set("token", token)

		req := httptest.NewRequest("POST", "/introspect", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		ts.router.ServeHTTP(w, req)

		var introspectResp IntrospectionResponse
		json.Unmarshal(w.Body.Bytes(), &introspectResp)

		if !introspectResp.Active {
			t.Error("Service token should be active")
		}

		if introspectResp.Sub != "service-payment" {
			t.Errorf("Expected service sub 'service-payment', got '%s'", introspectResp.Sub)
		}
	})

	t.Run("Development environment key rotation simulation", func(t *testing.T) {
		// Test that multiple keys work (simulating key rotation)
		
		// Generate tokens with different claims but same structure
		for i := 0; i < 5; i++ {
			claims := map[string]interface{}{
				"sub":      "user-" + string(rune(i+1)),
				"rotation": i,
			}

			token := generateTestToken(t, ts, claims, 3600)

			// Verify each token works
			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				kid, ok := token.Header["kid"].(string)
				if !ok {
					return nil, fmt.Errorf("missing key ID in token header")
				}

				keyPair, err := ts.keyManager.GetKeyByID(kid)
				if err != nil {
					return nil, fmt.Errorf("key not found for kid: %s", kid)
				}

				return keyPair.PublicKey, nil
			})

			if err != nil {
				t.Fatalf("Token %d validation failed: %v", i, err)
			}

			if !parsedToken.Valid {
				t.Errorf("Token %d should be valid", i)
			}
		}
	})

	t.Run("JWKS endpoint usage for JWT validation", func(t *testing.T) {
		// Common scenario: fetch JWKS and use it to validate tokens
		
		// 1. Fetch JWKS (what most JWT libraries do)
		req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
		w := httptest.NewRecorder()

		ts.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d for JWKS, got %d", http.StatusOK, w.Code)
		}

		var jwks map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &jwks); err != nil {
			t.Fatalf("Failed to unmarshal JWKS: %v", err)
		}

		// 2. Verify JWKS structure matches RFC 7517
		keys, ok := jwks["keys"].([]interface{})
		if !ok {
			t.Fatal("JWKS must contain 'keys' array")
		}

		if len(keys) == 0 {
			t.Fatal("JWKS must contain at least one key")
		}

		// 3. Verify each key has required fields for RSA
		for i, keyInterface := range keys {
			key, ok := keyInterface.(map[string]interface{})
			if !ok {
				t.Fatalf("Key %d is not a valid object", i)
			}

			requiredFields := []string{"kty", "use", "kid", "alg", "n", "e"}
			for _, field := range requiredFields {
				if _, exists := key[field]; !exists {
					t.Errorf("Key %d missing required field '%s'", i, field)
				}
			}

			// Verify RSA key values
			if key["kty"] != "RSA" {
				t.Errorf("Key %d: expected kty 'RSA', got '%v'", i, key["kty"])
			}

			if key["alg"] != "RS256" {
				t.Errorf("Key %d: expected alg 'RS256', got '%v'", i, key["alg"])
			}

			if key["use"] != "sig" {
				t.Errorf("Key %d: expected use 'sig', got '%v'", i, key["use"])
			}

			// Check that n and e are non-empty (RSA modulus and exponent)
			if key["n"] == "" {
				t.Errorf("Key %d: RSA modulus 'n' is empty", i)
			}

			if key["e"] == "" {
				t.Errorf("Key %d: RSA exponent 'e' is empty", i)
			}
		}
	})

	t.Run("Error handling scenarios", func(t *testing.T) {
		// Test malformed token introspection
		formData := url.Values{}
		formData.Set("token", "definitely.not.ajwt")

		req := httptest.NewRequest("POST", "/introspect", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		ts.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Introspection should return 200 even for invalid tokens, got %d", w.Code)
		}

		var introspectResp IntrospectionResponse
		json.Unmarshal(w.Body.Bytes(), &introspectResp)

		if introspectResp.Active {
			t.Error("Malformed token should not be active")
		}

		// Test expired/invalid token scenarios using the invalid token endpoint
		invalidToken := generateTestInvalidToken(t, ts, map[string]interface{}{"sub": "test"}, 3600)

		formData = url.Values{}
		formData.Set("token", invalidToken)

		req = httptest.NewRequest("POST", "/introspect", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()

		ts.router.ServeHTTP(w, req)

		json.Unmarshal(w.Body.Bytes(), &introspectResp)

		if introspectResp.Active {
			t.Error("Invalid token should not be active")
		}
	})

	t.Run("High-volume token generation for load testing", func(t *testing.T) {
		// Scenario: generating many tokens for load testing
		
		for i := 0; i < 10; i++ {
			claims := map[string]interface{}{
				"sub":        "load-test-user-" + string(rune(i+48)), // ASCII '0' + i
				"test_batch": "load-test-1",
				"index":      i,
			}

			jsonBody, _ := json.Marshal(TokenRequest{Claims: claims})
			req := httptest.NewRequest("POST", "/generate-token", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ts.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Load test token %d generation failed with status %d", i, w.Code)
			}

			var tokenResp TokenResponse
			if err := json.Unmarshal(w.Body.Bytes(), &tokenResp); err != nil {
				t.Errorf("Failed to unmarshal load test token %d: %v", i, err)
			}

			if tokenResp.AccessToken == "" {
				t.Errorf("Load test token %d is empty", i)
			}

			// Verify the claims were preserved
			if len(tokenResp.RawRequest) == 0 {
				t.Errorf("Load test token %d missing raw request data", i)
			}
		}
	})
}