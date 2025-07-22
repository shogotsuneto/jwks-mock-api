package common

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// AssertJSONResponse validates that a response is valid JSON and unmarshals it
func AssertJSONResponse(t *testing.T, body []byte, target interface{}) {
	t.Helper()
	
	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v\nResponse body: %s", err, string(body))
	}
}

// AssertValidJWT validates that a token string is a valid JWT
func AssertValidJWT(t *testing.T, tokenStr string) *jwt.Token {
	t.Helper()
	
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("Failed to parse JWT token: %v", err)
	}
	
	return token
}

// AssertJWTClaims validates specific claims in a JWT token
func AssertJWTClaims(t *testing.T, token *jwt.Token, expectedClaims map[string]interface{}) {
	t.Helper()
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to get JWT claims")
	}
	
	for key, expectedValue := range expectedClaims {
		actualValue, exists := claims[key]
		if !exists {
			t.Errorf("Expected claim '%s' not found in token", key)
			continue
		}
		
		// Handle different types of comparisons
		switch expected := expectedValue.(type) {
		case string:
			if actual, ok := actualValue.(string); !ok || actual != expected {
				t.Errorf("Expected claim '%s' to be '%v', got '%v'", key, expected, actualValue)
			}
		case float64:
			if actual, ok := actualValue.(float64); !ok || actual != expected {
				t.Errorf("Expected claim '%s' to be '%v', got '%v'", key, expected, actualValue)
			}
		case []interface{}:
			actualSlice, ok := actualValue.([]interface{})
			if !ok {
				t.Errorf("Expected claim '%s' to be a slice, got %T", key, actualValue)
				continue
			}
			if len(actualSlice) != len(expected) {
				t.Errorf("Expected claim '%s' to have %d items, got %d", key, len(expected), len(actualSlice))
				continue
			}
			// Note: This is a basic comparison, could be enhanced for complex types
		default:
			// Generic comparison for other types
			if actualValue != expectedValue {
				t.Errorf("Expected claim '%s' to be '%v', got '%v'", key, expectedValue, actualValue)
			}
		}
	}
}

// AssertValidJWKS validates that a JWKS response contains valid keys
func AssertValidJWKS(t *testing.T, jwks *JWKSResponse) {
	t.Helper()
	
	if len(jwks.Keys) == 0 {
		t.Fatal("JWKS response contains no keys")
	}
	
	for i, key := range jwks.Keys {
		if key.Kty == "" {
			t.Errorf("Key %d missing kty (key type)", i)
		}
		if key.Use == "" {
			t.Errorf("Key %d missing use", i)
		}
		if key.KeyID == "" {
			t.Errorf("Key %d missing kid (key ID)", i)
		}
		if key.Alg == "" {
			t.Errorf("Key %d missing alg (algorithm)", i)
		}
		if key.N == "" {
			t.Errorf("Key %d missing n (modulus)", i)
		}
		if key.E == "" {
			t.Errorf("Key %d missing e (exponent)", i)
		}
	}
}

// AssertResponseContains checks if response body contains expected strings
func AssertResponseContains(t *testing.T, body []byte, expectedStrings ...string) {
	t.Helper()
	
	bodyStr := string(body)
	for _, expected := range expectedStrings {
		if !strings.Contains(bodyStr, expected) {
			t.Errorf("Response body does not contain expected string: %s\nActual body: %s", expected, bodyStr)
		}
	}
}

// AssertResponseNotContains checks if response body does not contain specified strings
func AssertResponseNotContains(t *testing.T, body []byte, unexpectedStrings ...string) {
	t.Helper()
	
	bodyStr := string(body)
	for _, unexpected := range unexpectedStrings {
		if strings.Contains(bodyStr, unexpected) {
			t.Errorf("Response body contains unexpected string: %s\nActual body: %s", unexpected, bodyStr)
		}
	}
}

// AssertStatusCode checks if the HTTP response has the expected status code
func AssertStatusCode(t *testing.T, resp *http.Response, expectedCode int) {
	t.Helper()
	
	if resp.StatusCode != expectedCode {
		t.Errorf("Expected status code %d, got %d", expectedCode, resp.StatusCode)
	}
}

// AssertContentType checks if the HTTP response has the expected content type
func AssertContentType(t *testing.T, resp *http.Response, expectedType string) {
	t.Helper()
	
	actualType := resp.Header.Get("Content-Type")
	if !strings.Contains(actualType, expectedType) {
		t.Errorf("Expected content type to contain '%s', got '%s'", expectedType, actualType)
	}
}