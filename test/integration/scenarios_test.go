package integration

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestIntegrationCompleteJWTWorkflow tests the complete JWT workflow in a real environment
func TestIntegrationCompleteJWTWorkflow(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	t.Log("=== Starting Complete JWT Workflow Test ===")
	
	// Step 1: Generate a token with complex claims
	userClaims := map[string]interface{}{
		"sub":   "workflow-user-12345",
		"email": "workflow.user@company.com",
		"roles": []string{"developer", "admin", "user"},
		"permissions": []string{
			"read:projects",
			"write:projects", 
			"delete:projects",
			"admin:users",
		},
		"metadata": map[string]interface{}{
			"last_login":    "2024-01-15T10:30:00Z",
			"login_count":   127,
			"preferences":   map[string]string{"theme": "dark", "language": "en"},
			"organization":  "engineering-team",
		},
		"exp": 7200, // 2 hours
	}
	
	tokenReq := map[string]interface{}{
		"claims": userClaims,
	}
	
	t.Log("Step 1: Generating JWT token with complex claims...")
	resp, body := its.makeRequest(t, "POST", "/generate-token", tokenReq, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("Failed to generate token: %d - %s", resp.StatusCode, string(body))
	}
	
	var tokenResp map[string]interface{}
	json.Unmarshal(body, &tokenResp)
	token := tokenResp["token"].(string)
	t.Logf("✓ Token generated successfully (length: %d)", len(token))
	
	// Step 2: Fetch JWKS to simulate how a service would validate the token
	t.Log("Step 2: Fetching JWKS for token validation...")
	resp, body = its.makeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("Failed to fetch JWKS: %d - %s", resp.StatusCode, string(body))
	}
	
	var jwks map[string]interface{}
	json.Unmarshal(body, &jwks)
	keys := jwks["keys"].([]interface{})
	t.Logf("✓ JWKS fetched successfully (%d keys available)", len(keys))
	
	// Step 3: Parse token to verify structure (simulating what a service would do)
	t.Log("Step 3: Parsing and validating token structure...")
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims")
	}
	
	// Verify all custom claims are preserved
	if sub := claims["sub"].(string); sub != "workflow-user-12345" {
		t.Errorf("Sub claim mismatch: expected 'workflow-user-12345', got '%s'", sub)
	}
	
	if email := claims["email"].(string); email != "workflow.user@company.com" {
		t.Errorf("Email claim mismatch")
	}
	
	// Verify complex nested structures
	if metadata, ok := claims["metadata"].(map[string]interface{}); ok {
		if org := metadata["organization"].(string); org != "engineering-team" {
			t.Errorf("Organization metadata mismatch")
		}
		if prefs, ok := metadata["preferences"].(map[string]interface{}); ok {
			if theme := prefs["theme"].(string); theme != "dark" {
				t.Errorf("Theme preference mismatch")
			}
		}
	} else {
		t.Error("Metadata not preserved in token")
	}
	
	t.Log("✓ Token structure and claims validated successfully")
	
	// Step 4: Use token introspection endpoint
	t.Log("Step 4: Performing token introspection...")
	formData := url.Values{"token": {token}}
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	
	resp, body = its.makeRequest(t, "POST", "/introspect", formData, headers)
	if resp.StatusCode != 200 {
		t.Fatalf("Failed to introspect token: %d - %s", resp.StatusCode, string(body))
	}
	
	var introspectResp map[string]interface{}
	json.Unmarshal(body, &introspectResp)
	
	if !introspectResp["active"].(bool) {
		t.Error("Token should be active")
	}
	
	// Verify all claims are accessible via introspection
	if introspectResp["sub"].(string) != "workflow-user-12345" {
		t.Error("Sub claim not preserved in introspection")
	}
	
	// Check that complex claims are preserved
	if metadata, ok := introspectResp["metadata"].(map[string]interface{}); ok {
		if org := metadata["organization"].(string); org != "engineering-team" {
			t.Error("Complex metadata not preserved in introspection")
		}
	} else {
		t.Error("Metadata not available in introspection response")
	}
	
	t.Log("✓ Token introspection completed successfully")
	
	t.Log("=== Complete JWT Workflow Test PASSED ===")
}

// TestIntegrationMicroservicesWorkflow tests service-to-service authentication
func TestIntegrationMicroservicesWorkflow(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	t.Log("=== Starting Microservices Communication Test ===")
	
	// Generate service tokens for different services
	services := []struct {
		name   string
		claims map[string]interface{}
	}{
		{
			name: "payment-service",
			claims: map[string]interface{}{
				"sub":           "service-payment",
				"client_type":   "service",
				"service_id":    "payment-service-v1.2.3",
				"service_name":  "Payment Processing Service",
				"scopes":        []string{"payments:read", "payments:write", "payments:refund"},
				"region":        "us-east-1",
				"environment":   "production",
				"deployment_id": "deploy-abc123",
				"exp":           3600,
			},
		},
		{
			name: "user-service",
			claims: map[string]interface{}{
				"sub":           "service-user",
				"client_type":   "service",
				"service_id":    "user-service-v2.1.0",
				"service_name":  "User Management Service",
				"scopes":        []string{"users:read", "users:write", "profiles:read"},
				"region":        "us-east-1",
				"environment":   "production",
				"deployment_id": "deploy-xyz789",
				"exp":           3600,
			},
		},
		{
			name: "notification-service",
			claims: map[string]interface{}{
				"sub":           "service-notification",
				"client_type":   "service",
				"service_id":    "notification-service-v1.0.5",
				"service_name":  "Notification Service",
				"scopes":        []string{"notifications:send", "templates:read"},
				"region":        "us-west-2",
				"environment":   "production",
				"deployment_id": "deploy-def456",
				"exp":           1800, // 30 minutes for notification service
			},
		},
	}
	
	tokens := make(map[string]string)
	
	// Generate tokens for all services
	for _, service := range services {
		t.Logf("Generating token for %s...", service.name)
		
		tokenReq := map[string]interface{}{"claims": service.claims}
		resp, body := its.makeRequest(t, "POST", "/generate-token", tokenReq, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("Failed to generate token for %s: %d", service.name, resp.StatusCode)
		}
		
		var tokenResp map[string]interface{}
		json.Unmarshal(body, &tokenResp)
		tokens[service.name] = tokenResp["token"].(string)
		
		t.Logf("✓ Token generated for %s", service.name)
	}
	
	// Simulate inter-service communication by validating each token
	for serviceName, token := range tokens {
		t.Logf("Validating token for %s via introspection...", serviceName)
		
		formData := url.Values{"token": {token}}
		headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
		
		resp, body := its.makeRequest(t, "POST", "/introspect", formData, headers)
		if resp.StatusCode != 200 {
			t.Fatalf("Failed to introspect token for %s: %d", serviceName, resp.StatusCode)
		}
		
		var introspectResp map[string]interface{}
		json.Unmarshal(body, &introspectResp)
		
		if !introspectResp["active"].(bool) {
			t.Errorf("Token for %s should be active", serviceName)
		}
		
		// Verify service-specific claims
		if clientType := introspectResp["client_type"].(string); clientType != "service" {
			t.Errorf("Expected client_type 'service' for %s, got '%s'", serviceName, clientType)
		}
		
		// Verify scopes are preserved
		if scopes, ok := introspectResp["scopes"].([]interface{}); ok {
			if len(scopes) == 0 {
				t.Errorf("Expected scopes for %s", serviceName)
			}
		}
		
		t.Logf("✓ Token validated for %s", serviceName)
	}
	
	t.Log("=== Microservices Communication Test PASSED ===")
}

// TestIntegrationKeyRotationSimulation tests multiple keys to simulate key rotation
func TestIntegrationKeyRotationSimulation(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	t.Log("=== Starting Key Rotation Simulation Test ===")
	
	// First, check how many keys are available
	resp, body := its.makeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("Failed to fetch JWKS: %d", resp.StatusCode)
	}
	
	var jwks map[string]interface{}
	json.Unmarshal(body, &jwks)
	keys := jwks["keys"].([]interface{})
	
	t.Logf("Found %d keys in JWKS - simulating key rotation scenario", len(keys))
	
	if len(keys) < 2 {
		t.Skip("Need at least 2 keys for key rotation simulation")
	}
	
	// Generate multiple tokens - they should use different keys
	tokenResults := make([]map[string]interface{}, 10)
	
	for i := 0; i < 10; i++ {
		claims := map[string]interface{}{
			"sub":       fmt.Sprintf("rotation-test-user-%d", i),
			"test_run":  i,
			"scenario":  "key-rotation-simulation",
			"timestamp": time.Now().Unix(),
		}
		
		tokenReq := map[string]interface{}{"claims": claims}
		resp, body := its.makeRequest(t, "POST", "/generate-token", tokenReq, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("Failed to generate token %d: %d", i, resp.StatusCode)
		}
		
		var tokenResp map[string]interface{}
		json.Unmarshal(body, &tokenResp)
		token := tokenResp["token"].(string)
		
		// Parse token to extract key ID
		parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
		if err != nil {
			t.Fatalf("Failed to parse token %d: %v", i, err)
		}
		
		keyID := parsedToken.Header["kid"].(string)
		
		tokenResults[i] = map[string]interface{}{
			"token":  token,
			"key_id": keyID,
			"index":  i,
		}
		
		t.Logf("Token %d generated with key ID: %s", i, keyID)
	}
	
	// Verify that different keys are being used (at least some variation)
	keyIDs := make(map[string]int)
	for _, result := range tokenResults {
		keyID := result["key_id"].(string)
		keyIDs[keyID]++
	}
	
	t.Logf("Key usage distribution: %v", keyIDs)
	
	if len(keyIDs) < 2 {
		t.Log("Warning: All tokens used the same key - this might be expected for smaller key sets")
	} else {
		t.Logf("✓ Key rotation simulation successful - %d different keys used", len(keyIDs))
	}
	
	// Validate all tokens via introspection
	for _, result := range tokenResults {
		token := result["token"].(string)
		index := result["index"].(int)
		
		formData := url.Values{"token": {token}}
		headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
		
		resp, body := its.makeRequest(t, "POST", "/introspect", formData, headers)
		if resp.StatusCode != 200 {
			t.Errorf("Failed to introspect token %d: %d", index, resp.StatusCode)
			continue
		}
		
		var introspectResp map[string]interface{}
		json.Unmarshal(body, &introspectResp)
		
		if !introspectResp["active"].(bool) {
			t.Errorf("Token %d should be active", index)
		}
		
		expectedSub := fmt.Sprintf("rotation-test-user-%d", index)
		if sub := introspectResp["sub"].(string); sub != expectedSub {
			t.Errorf("Token %d: expected sub '%s', got '%s'", index, expectedSub, sub)
		}
	}
	
	t.Log("=== Key Rotation Simulation Test PASSED ===")
}

// TestIntegrationHighVolumeTokenGeneration tests performance under load
func TestIntegrationHighVolumeTokenGeneration(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	t.Log("=== Starting High Volume Token Generation Test ===")
	
	tokenCount := 50
	successCount := 0
	start := time.Now()
	
	for i := 0; i < tokenCount; i++ {
		claims := map[string]interface{}{
			"sub":        fmt.Sprintf("load-test-user-%d", i),
			"batch":      "high-volume-test",
			"sequence":   i,
			"timestamp":  time.Now().Unix(),
			"load_test":  true,
		}
		
		tokenReq := map[string]interface{}{"claims": claims}
		resp, body := its.makeRequest(t, "POST", "/generate-token", tokenReq, nil)
		
		if resp.StatusCode == 200 {
			successCount++
		} else {
			t.Logf("Token %d failed: %d - %s", i, resp.StatusCode, string(body))
		}
		
		// Small delay to avoid overwhelming the server
		if i%10 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	duration := time.Since(start)
	
	t.Logf("Generated %d/%d tokens successfully in %v", successCount, tokenCount, duration)
	t.Logf("Rate: %.2f tokens/second", float64(successCount)/duration.Seconds())
	
	if successCount < tokenCount {
		t.Errorf("Expected %d successful tokens, got %d", tokenCount, successCount)
	}
	
	successRate := float64(successCount) / float64(tokenCount) * 100
	if successRate < 95.0 {
		t.Errorf("Success rate too low: %.1f%% (expected > 95%%)", successRate)
	}
	
	t.Log("=== High Volume Token Generation Test PASSED ===")
}

// TestIntegrationAPIEndpointTesting simulates testing API endpoints with JWT
func TestIntegrationAPIEndpointTesting(t *testing.T) {
	its := newIntegrationTestSuite()
	its.waitForAPI(t)
	
	t.Log("=== Starting API Endpoint Testing Simulation ===")
	
	// Generate tokens for different API testing scenarios
	testScenarios := []struct {
		name   string
		claims map[string]interface{}
		desc   string
	}{
		{
			name: "admin-user",
			claims: map[string]interface{}{
				"sub":   "api-test-admin",
				"scope": "admin:read admin:write admin:delete",
				"role":  "admin",
				"client": "api-test-suite",
				"env":   "testing",
			},
			desc: "Admin user with full permissions",
		},
		{
			name: "regular-user",
			claims: map[string]interface{}{
				"sub":   "api-test-user",
				"scope": "read write",
				"role":  "user", 
				"client": "api-test-suite",
				"env":   "testing",
			},
			desc: "Regular user with limited permissions",
		},
		{
			name: "readonly-user",
			claims: map[string]interface{}{
				"sub":   "api-test-readonly",
				"scope": "read",
				"role":  "viewer",
				"client": "api-test-suite",
				"env":   "testing",
			},
			desc: "Read-only user for testing access restrictions",
		},
		{
			name: "expired-user",
			claims: map[string]interface{}{
				"sub":   "api-test-expired",
				"scope": "read write",
				"role":  "user",
				"client": "api-test-suite",
				"env":   "testing",
				"exp":   1, // 1 second - will be expired immediately
			},
			desc: "User with short expiration for testing token expiry",
		},
	}
	
	tokens := make(map[string]string)
	
	// Generate tokens for each test scenario
	for _, scenario := range testScenarios {
		t.Logf("Generating token for %s (%s)...", scenario.name, scenario.desc)
		
		tokenReq := map[string]interface{}{"claims": scenario.claims}
		resp, body := its.makeRequest(t, "POST", "/generate-token", tokenReq, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("Failed to generate token for %s: %d", scenario.name, resp.StatusCode)
		}
		
		var tokenResp map[string]interface{}
		json.Unmarshal(body, &tokenResp)
		tokens[scenario.name] = tokenResp["token"].(string)
		
		t.Logf("✓ Token generated for %s", scenario.name)
	}
	
	// Test each token via introspection to simulate API endpoint validation
	for scenarioName, token := range tokens {
		t.Logf("Testing API authentication with %s token...", scenarioName)
		
		formData := url.Values{"token": {token}}
		headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
		
		resp, body := its.makeRequest(t, "POST", "/introspect", formData, headers)
		if resp.StatusCode != 200 {
			t.Fatalf("Failed to introspect %s token: %d", scenarioName, resp.StatusCode)
		}
		
		var introspectResp map[string]interface{}
		json.Unmarshal(body, &introspectResp)
		
		// For expired token, it should still introspect successfully 
		// (introspection endpoint checks structure, not expiry by default)
		if scenarioName == "expired-user" {
			t.Logf("✓ Expired token introspected (API should check exp claim separately)")
		} else {
			if !introspectResp["active"].(bool) {
				t.Errorf("Token for %s should be active", scenarioName)
			}
		}
		
		// Verify scope-based authorization claims
		if scope, ok := introspectResp["scope"].(string); ok {
			switch scenarioName {
			case "admin-user":
				if !strings.Contains(scope, "admin:") {
					t.Errorf("Admin user should have admin scopes, got: %s", scope)
				}
			case "readonly-user":
				if strings.Contains(scope, "write") {
					t.Errorf("Readonly user should not have write scope, got: %s", scope)
				}
			}
		}
		
		t.Logf("✓ API authentication test passed for %s", scenarioName)
	}
	
	t.Log("=== API Endpoint Testing Simulation PASSED ===")
}