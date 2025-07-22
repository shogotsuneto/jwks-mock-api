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
	
	t.Log("Step 1: Generating JWT token with complex claims...")
	resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
	common.AssertStatusCode(t, resp, 200)
	
	var tokenResp common.TokenResponse
	common.AssertJSONResponse(t, body, &tokenResp)
	t.Logf("✓ Token generated successfully (length: %d)", len(tokenResp.Token))
	
	// Step 2: Fetch JWKS to simulate how a service would validate the token
	t.Log("Step 2: Fetching JWKS for token validation...")
	resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)
	
	var jwks common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwks)
	t.Logf("✓ JWKS fetched successfully (%d keys available)", len(jwks.Keys))
	
	// Step 3: Parse token to verify structure (simulating what a service would do)
	t.Log("Step 3: Parsing and validating token structure...")
	parsedToken := common.AssertValidJWT(t, tokenResp.Token)
	
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
	formData := url.Values{"token": {tokenResp.Token}}
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	
	resp, body = its.MakeRequest(t, "POST", "/introspect", formData, headers)
	common.AssertStatusCode(t, resp, 200)
	
	var introspectResp common.IntrospectionResponse
	common.AssertJSONResponse(t, body, &introspectResp)
	
	if !introspectResp.Active {
		t.Error("Token should be active")
	}
	
	// Verify all claims are accessible via introspection
	if introspectResp.Sub != "workflow-user-12345" {
		t.Error("Sub claim not preserved in introspection")
	}
	
	// Check that complex claims are preserved
	if metadata, ok := introspectResp.Claims["metadata"].(map[string]interface{}); ok {
		if org := metadata["organization"].(string); org != "engineering-team" {
			t.Error("Complex metadata not preserved in introspection")
		}
	} else {
		t.Error("Metadata not available in introspection response")
	}
	
	t.Log("✓ Token introspection completed successfully")
	
	t.Log("=== Complete JWT Workflow Test PASSED ===")
}