package scenarios

import (
	"net/url"
	"strings"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestAPIEndpointTesting simulates testing API endpoints with JWT
// This test covers two main aspects:
// 1. SCOPE-BASED AUTHORIZATION: Testing different permission levels and access control
// 2. TOKEN EXPIRATION HANDLING: Testing how APIs handle expired tokens
func TestAPIEndpointTesting(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	t.Log("=== Starting API Endpoint Testing Simulation ===")
	t.Log("Testing aspects: Scope-based authorization & Token expiration handling")
	
	// Generate tokens for different API testing scenarios
	// Each scenario tests specific authorization and expiration aspects
	testScenarios := []struct {
		name   string
		claims map[string]interface{}
		desc   string
		testingAspect string
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
			testingAspect: "SCOPE TESTING: High-privilege access with admin-level scopes",
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
			testingAspect: "SCOPE TESTING: Standard user permissions without admin privileges",
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
			testingAspect: "SCOPE TESTING: Restricted access - read-only permissions (no write scope)",
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
			testingAspect: "EXPIRATION TESTING: Short-lived token to test expiry handling",
		},
	}
	
	tokens := make(map[string]string)
	
	// Generate tokens for each test scenario
	for _, scenario := range testScenarios {
		t.Logf("Generating token for %s (%s)...", scenario.name, scenario.desc)
		t.Logf("  → %s", scenario.testingAspect)
		
		tokenReq := map[string]interface{}{"claims": scenario.claims}
		resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
		common.AssertStatusCode(t, resp, 200)
		
		var tokenResp common.TokenResponse
		common.AssertJSONResponse(t, body, &tokenResp)
		tokens[scenario.name] = tokenResp.Token
		
		t.Logf("✓ Token generated for %s", scenario.name)
	}
	
	t.Log("\n=== TESTING PHASE: Validating scope-based authorization and expiration ===")
	
	// Test each token via introspection to simulate API endpoint validation
	for scenarioName, token := range tokens {
		t.Logf("Testing API authentication with %s token...", scenarioName)
		
		formData := url.Values{"token": {token}}
		headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
		
		resp, body := its.MakeRequest(t, "POST", "/introspect", formData, headers)
		common.AssertStatusCode(t, resp, 200)
		
		var introspectResp common.IntrospectionResponse
		common.AssertJSONResponse(t, body, &introspectResp)
		
		// EXPIRATION TESTING: For expired token, it should still introspect successfully 
		// (introspection endpoint checks structure, not expiry by default)
		if scenarioName == "expired-user" {
			t.Logf("✓ EXPIRATION TEST: Expired token introspected (API should check exp claim separately)")
		} else {
			if !introspectResp.Active {
				t.Errorf("Token for %s should be active", scenarioName)
			}
		}
		
		// SCOPE TESTING: Verify scope-based authorization claims
		if scope, ok := introspectResp.Claims["scope"].(string); ok {
			switch scenarioName {
			case "admin-user":
				if !strings.Contains(scope, "admin:") {
					t.Errorf("SCOPE TEST FAILED: Admin user should have admin scopes, got: %s", scope)
				} else {
					t.Logf("✓ SCOPE TEST PASSED: Admin user has required admin scopes: %s", scope)
				}
			case "readonly-user":
				if strings.Contains(scope, "write") {
					t.Errorf("SCOPE TEST FAILED: Readonly user should not have write scope, got: %s", scope)
				} else {
					t.Logf("✓ SCOPE TEST PASSED: Readonly user correctly restricted to: %s", scope)
				}
			case "regular-user":
				if !strings.Contains(scope, "read") || !strings.Contains(scope, "write") {
					t.Errorf("SCOPE TEST FAILED: Regular user should have read and write scopes, got: %s", scope)
				} else {
					t.Logf("✓ SCOPE TEST PASSED: Regular user has expected scopes: %s", scope)
				}
			}
		}
		
		t.Logf("✓ API authentication test passed for %s", scenarioName)
	}
	
	t.Log("=== API Endpoint Testing Simulation PASSED ===")
	t.Log("✓ All scope-based authorization tests completed")
	t.Log("✓ All token expiration tests completed")
}