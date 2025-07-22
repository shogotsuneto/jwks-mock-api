package endpoints

import (
	"net/http"
	"strings"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestTokenGeneration tests token generation endpoint
func TestTokenGeneration(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
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
	
	resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
	
	common.AssertStatusCode(t, resp, http.StatusOK)
	common.AssertContentType(t, resp, "application/json")
	
	var tokenResp common.TokenResponse
	common.AssertJSONResponse(t, body, &tokenResp)

	if tokenResp.Token == "" {
		t.Fatal("❌ TOKEN GENERATION FAILED: Expected 'token' field in response")
	}

	// Verify token structure (should have 3 parts separated by dots)
	parts := strings.Split(tokenResp.Token, ".")
	if len(parts) != 3 {
		t.Errorf("❌ TOKEN STRUCTURE INVALID: Expected JWT with 3 parts, got %d", len(parts))
	}

	// Parse and validate JWT
	token := common.AssertValidJWT(t, tokenResp.Token)
	
	// Verify custom claims
	expectedClaims := map[string]interface{}{
		"sub": "integration-test-user",
	}
	common.AssertJWTClaims(t, token, expectedClaims)
	
	t.Log("✅ Token generation test passed")
}