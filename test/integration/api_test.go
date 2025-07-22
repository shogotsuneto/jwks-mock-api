package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultAPIURL = "http://localhost:3001"
	testTimeout   = 30 * time.Second
)

// integrationTestSuite manages the integration test environment
type integrationTestSuite struct {
	apiURL     string
	httpClient *http.Client
}

// newIntegrationTestSuite creates a new integration test suite
func newIntegrationTestSuite() *integrationTestSuite {
	apiURL := os.Getenv("JWKS_API_URL")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	timeout := testTimeout
	if timeoutStr := os.Getenv("TEST_TIMEOUT"); timeoutStr != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsedTimeout
		}
	}

	return &integrationTestSuite{
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// waitForAPI waits for the API to be ready
func (its *integrationTestSuite) waitForAPI(t *testing.T) {
	t.Helper()
	
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := its.httpClient.Get(its.apiURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Logf("API is ready after %d attempts", i+1)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		
		t.Logf("Waiting for API... attempt %d/%d", i+1, maxAttempts)
		time.Sleep(2 * time.Second)
	}
	
	t.Fatalf("API did not become ready after %d attempts", maxAttempts)
}

// makeRequest is a helper to make HTTP requests
func (its *integrationTestSuite) makeRequest(t *testing.T, method, endpoint string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	
	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
		case []byte:
			reqBody = bytes.NewReader(v)
		case url.Values:
			reqBody = strings.NewReader(v.Encode())
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}
			reqBody = bytes.NewReader(jsonBody)
		}
	}
	
	req, err := http.NewRequest(method, its.apiURL+endpoint, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Set default content type for POST requests
	if method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	resp, err := its.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request to %s %s: %v", method, endpoint, err)
	}
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	resp.Body.Close()
	
	return resp, respBody
}

// TestIntegrationHealth tests the health endpoint with real HTTP requests
func TestIntegrationHealth(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	resp, body := its.makeRequest(t, "GET", "/health", nil, nil)
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
	
	var healthResp map[string]interface{}
	if err := json.Unmarshal(body, &healthResp); err != nil {
		t.Fatalf("Failed to parse health response: %v", err)
	}
	
	// Verify required fields
	if status, ok := healthResp["status"].(string); !ok || status != "ok" {
		t.Errorf("Expected status 'ok', got %v", healthResp["status"])
	}
	
	if _, ok := healthResp["available_keys"]; !ok {
		t.Error("Expected 'available_keys' field in health response")
	}
	
	t.Logf("Health check passed: %s", string(body))
}

// TestIntegrationJWKS tests the JWKS endpoint
func TestIntegrationJWKS(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	resp, body := its.makeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
	
	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}
	
	var jwks map[string]interface{}
	if err := json.Unmarshal(body, &jwks); err != nil {
		t.Fatalf("Failed to parse JWKS response: %v", err)
	}
	
	// Verify JWKS structure
	keys, ok := jwks["keys"].([]interface{})
	if !ok {
		t.Fatal("Expected 'keys' array in JWKS response")
	}
	
	if len(keys) == 0 {
		t.Fatal("Expected at least one key in JWKS")
	}
	
	// Verify key structure
	for i, keyInterface := range keys {
		key, ok := keyInterface.(map[string]interface{})
		if !ok {
			t.Errorf("Key %d is not a valid object", i)
			continue
		}
		
		requiredFields := []string{"kty", "use", "kid", "alg", "n", "e"}
		for _, field := range requiredFields {
			if _, exists := key[field]; !exists {
				t.Errorf("Key %d missing required field: %s", i, field)
			}
		}
	}
	
	t.Logf("JWKS validation passed with %d keys", len(keys))
}

// TestIntegrationTokenGeneration tests token generation endpoint
func TestIntegrationTokenGeneration(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	// Test with custom claims
	claims := map[string]interface{}{
		"sub":   "integration-test-user",
		"email": "test@integration.com",
		"roles": []string{"user", "tester"},
		"exp":   3600, // 1 hour
	}
	
	tokenReq := map[string]interface{}{
		"claims": claims,
	}
	
	resp, body := its.makeRequest(t, "POST", "/generate-token", tokenReq, nil)
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
	
	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		t.Fatalf("Failed to parse token response: %v", err)
	}
	
	token, ok := tokenResp["token"].(string)
	if !ok || token == "" {
		t.Fatal("Expected 'token' field in response")
	}
	
	// Verify token structure (should have 3 parts separated by dots)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected JWT with 3 parts, got %d", len(parts))
	}
	
	// Parse token without verification (we'll verify separately)
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}
	
	tokenClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims from token")
	}
	
	// Verify custom claims are present
	if sub, ok := tokenClaims["sub"].(string); !ok || sub != "integration-test-user" {
		t.Errorf("Expected sub 'integration-test-user', got %v", tokenClaims["sub"])
	}
	
	t.Logf("Token generation test passed, token length: %d", len(token))
}

// TestIntegrationTokenIntrospection tests token introspection endpoint
func TestIntegrationTokenIntrospection(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	// First generate a token
	claims := map[string]interface{}{
		"sub":    "introspection-test-user",
		"scope":  "read write",
		"client": "test-client",
	}
	
	tokenReq := map[string]interface{}{
		"claims": claims,
	}
	
	resp, body := its.makeRequest(t, "POST", "/generate-token", tokenReq, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to generate token: %d - %s", resp.StatusCode, string(body))
	}
	
	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		t.Fatalf("Failed to parse token response: %v", err)
	}
	
	token := tokenResp["token"].(string)
	
	// Test introspection with form data
	formData := url.Values{
		"token": {token},
	}
	
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	
	resp, body = its.makeRequest(t, "POST", "/introspect", formData, headers)
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
	
	var introspectResp map[string]interface{}
	if err := json.Unmarshal(body, &introspectResp); err != nil {
		t.Fatalf("Failed to parse introspection response: %v", err)
	}
	
	// Verify introspection response
	if active, ok := introspectResp["active"].(bool); !ok || !active {
		t.Errorf("Expected active=true, got %v", introspectResp["active"])
	}
	
	if sub, ok := introspectResp["sub"].(string); !ok || sub != "introspection-test-user" {
		t.Errorf("Expected sub 'introspection-test-user', got %v", introspectResp["sub"])
	}
	
	t.Log("Token introspection test passed")
}

// TestIntegrationInvalidToken tests introspection with invalid token
func TestIntegrationInvalidToken(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	// Test with completely invalid token
	formData := url.Values{
		"token": {"invalid.token.here"},
	}
	
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	
	resp, body := its.makeRequest(t, "POST", "/introspect", formData, headers)
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
	
	var introspectResp map[string]interface{}
	if err := json.Unmarshal(body, &introspectResp); err != nil {
		t.Fatalf("Failed to parse introspection response: %v", err)
	}
	
	// Verify token is marked as inactive
	if active, ok := introspectResp["active"].(bool); !ok || active {
		t.Errorf("Expected active=false for invalid token, got %v", introspectResp["active"])
	}
	
	t.Log("Invalid token test passed")
}

// TestIntegrationCORS tests CORS support
func TestIntegrationCORS(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	// Test preflight OPTIONS request
	headers := map[string]string{
		"Origin":                         "http://localhost:3000",
		"Access-Control-Request-Method":  "POST",
		"Access-Control-Request-Headers": "Content-Type",
	}
	
	resp, _ := its.makeRequest(t, "OPTIONS", "/generate-token", nil, headers)
	
	// Verify CORS headers
	if resp.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header")
	}
	
	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods header")
	}
	
	if resp.Header.Get("Access-Control-Allow-Headers") == "" {
		t.Error("Expected Access-Control-Allow-Headers header")
	}
	
	t.Log("CORS test passed")
}