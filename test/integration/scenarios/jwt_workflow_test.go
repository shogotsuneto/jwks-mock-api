package scenarios

import (
	"net/url"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
	"github.com/golang-jwt/jwt/v5"
)

// TestCompleteJWTWorkflow tests the complete JWT workflow in a real environment
func TestCompleteJWTWorkflow(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
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
	
	resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
	common.AssertStatusCode(t, resp, 200)
	
	var tokenResp common.TokenResponse
	common.AssertJSONResponse(t, body, &tokenResp)
	
	// Step 2: Fetch JWKS to simulate how a service would validate the token
	resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)
	
	var jwks common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwks)
	
	// Step 3: Parse token to verify structure (simulating what a service would do)
	parsedToken := common.AssertValidJWT(t, tokenResp.Token)
	
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("❌ JWT WORKFLOW FAILED: Failed to extract claims")
	}
	
	// Verify all custom claims are preserved
	if sub := claims["sub"].(string); sub != "workflow-user-12345" {
		t.Errorf("❌ JWT WORKFLOW FAILED: Sub claim mismatch: expected 'workflow-user-12345', got '%s'", sub)
	}
	
	if email := claims["email"].(string); email != "workflow.user@company.com" {
		t.Error("❌ JWT WORKFLOW FAILED: Email claim mismatch")
	}
	
	// Verify complex nested structures
	if metadata, ok := claims["metadata"].(map[string]interface{}); ok {
		if org := metadata["organization"].(string); org != "engineering-team" {
			t.Error("❌ JWT WORKFLOW FAILED: Organization metadata mismatch")
		}
		if prefs, ok := metadata["preferences"].(map[string]interface{}); ok {
			if theme := prefs["theme"].(string); theme != "dark" {
				t.Error("❌ JWT WORKFLOW FAILED: Theme preference mismatch")
			}
		}
	} else {
		t.Error("❌ JWT WORKFLOW FAILED: Metadata not preserved in token")
	}
	
	// Step 4: Use token introspection endpoint
	formData := url.Values{"token": {tokenResp.Token}}
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	
	resp, body = its.MakeRequest(t, "POST", "/introspect", formData, headers)
	common.AssertStatusCode(t, resp, 200)
	
	var introspectResp common.IntrospectionResponse
	common.AssertJSONResponse(t, body, &introspectResp)
	
	if !introspectResp.Active {
		t.Error("❌ JWT WORKFLOW FAILED: Token should be active")
	}
	
	// Per RFC 7662, introspection endpoint guarantees basic token validation
	// Sub field is a standard claim that can be verified
	if introspectResp.Sub != "workflow-user-12345" {
		t.Error("❌ JWT WORKFLOW FAILED: Sub claim not preserved in introspection")
	}
	
	t.Log("✅ Complete JWT Workflow Test PASSED")
}