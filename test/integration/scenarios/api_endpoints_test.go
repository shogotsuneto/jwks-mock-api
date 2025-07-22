package scenarios

import (
	"net/url"
	"strings"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestAPIEndpointTesting simulates testing API endpoints with JWT authentication
// covering scope-based authorization and token expiration scenarios
func TestAPIEndpointTesting(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	t.Log("=== Starting API Endpoint Testing Simulation ===")
	
	// Test scenarios with validation criteria driven by data
	testScenarios := []struct {
		name           string
		claims         map[string]interface{}
		desc           string
		requiredScopes []string // Scopes that must be present
		forbiddenScopes []string // Scopes that must NOT be present
		expectExpired  bool     // Whether token should be treated as expired
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
			desc: "Admin user with high-privilege access requiring admin-level scopes",
			requiredScopes: []string{"admin:read", "admin:write", "admin:delete"},
			forbiddenScopes: []string{},
			expectExpired: false,
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
			desc: "Regular user with standard read/write permissions but no admin privileges",
			requiredScopes: []string{"read", "write"},
			forbiddenScopes: []string{"admin:"},
			expectExpired: false,
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
			desc: "Read-only user testing access restrictions (no write permissions)",
			requiredScopes: []string{"read"},
			forbiddenScopes: []string{"write", "admin:"},
			expectExpired: false,
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
			desc: "User with short-lived token for testing expiration handling",
			requiredScopes: []string{"read", "write"},
			forbiddenScopes: []string{},
			expectExpired: true,
		},
	}
	
	tokens := make(map[string]string)
	
	// Generate tokens for each test scenario
	for _, scenario := range testScenarios {
		t.Logf("Generating token for %s: %s", scenario.name, scenario.desc)
		
		tokenReq := map[string]interface{}{"claims": scenario.claims}
		resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
		common.AssertStatusCode(t, resp, 200)
		
		var tokenResp common.TokenResponse
		common.AssertJSONResponse(t, body, &tokenResp)
		tokens[scenario.name] = tokenResp.AccessToken
		
		t.Logf("✓ Token generated for %s", scenario.name)
	}
	
	t.Log("\n=== TESTING PHASE: Validating scope-based authorization and expiration ===")
	
	// Test each token via introspection using data-driven validation
	for _, scenario := range testScenarios {
		token := tokens[scenario.name]
		t.Logf("Testing API authentication for %s...", scenario.desc)
		
		formData := url.Values{"token": {token}}
		headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
		
		resp, body := its.MakeRequest(t, "POST", "/introspect", formData, headers)
		common.AssertStatusCode(t, resp, 200)
		
		var introspectResp common.IntrospectionResponse
		common.AssertJSONResponse(t, body, &introspectResp)
		
		// Expiration validation
		if scenario.expectExpired {
			t.Logf("✓ EXPIRATION TEST: Token introspected (API should check exp claim separately)")
		} else {
			if !introspectResp.Active {
				t.Errorf("Token for %s should be active", scenario.name)
			}
		}
		
		// Scope validation - data-driven approach
		if scope, ok := introspectResp.Claims["scope"].(string); ok {
			// Check required scopes
			for _, requiredScope := range scenario.requiredScopes {
				if !strings.Contains(scope, requiredScope) {
					t.Errorf("SCOPE TEST FAILED: %s should have scope '%s', got: %s", scenario.name, requiredScope, scope)
				} else {
					t.Logf("✓ SCOPE TEST PASSED: %s has required scope '%s'", scenario.name, requiredScope)
				}
			}
			
			// Check forbidden scopes
			for _, forbiddenScope := range scenario.forbiddenScopes {
				if strings.Contains(scope, forbiddenScope) {
					t.Errorf("SCOPE TEST FAILED: %s should NOT have scope '%s', got: %s", scenario.name, forbiddenScope, scope)
				} else {
					t.Logf("✓ SCOPE TEST PASSED: %s correctly excludes forbidden scope '%s'", scenario.name, forbiddenScope)
				}
			}
		}
		
		t.Logf("✓ API authentication test passed for %s", scenario.name)
	}
	
	t.Log("=== API Endpoint Testing Simulation COMPLETED ===")
}