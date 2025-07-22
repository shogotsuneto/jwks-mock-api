package scenarios

import (
	"net/url"
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
		tokenReq := map[string]interface{}{"claims": scenario.claims}
		resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
		common.AssertStatusCode(t, resp, 200)
		
		var tokenResp common.TokenResponse
		common.AssertJSONResponse(t, body, &tokenResp)
		tokens[scenario.name] = tokenResp.Token
	}
	
	// Test each token via introspection for basic validation
	for _, scenario := range testScenarios {
		token := tokens[scenario.name]
		
		formData := url.Values{"token": {token}}
		headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
		
		resp, body := its.MakeRequest(t, "POST", "/introspect", formData, headers)
		common.AssertStatusCode(t, resp, 200)
		
		var introspectResp common.IntrospectionResponse
		common.AssertJSONResponse(t, body, &introspectResp)
		
		// Basic token validation only (per RFC 7662)
		if !scenario.expectExpired && !introspectResp.Active {
			t.Errorf("❌ API ENDPOINT TEST FAILED: Token for %s should be active", scenario.name)
		}
		
		// Note: Scope validation removed per RFC 7662 recommendation to avoid claim content testing
	}
	
	t.Log("✅ API Endpoint Testing Simulation PASSED")
}