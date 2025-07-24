package scenarios

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestAudienceWorkflow tests the complete audience workflow including token generation, 
// validation, and introspection with custom and default audiences
func TestAudienceWorkflow(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)

	t.Log("=== Starting Audience Workflow Test ===")

	// Test 1: Token generation with custom audience
	t.Log("Testing token generation with custom audience...")
	customAudience := "custom-api-service"
	customClaims := map[string]interface{}{
		"sub": "test-user",
		"aud": customAudience,
		"role": "admin",
	}

	tokenReq := map[string]interface{}{
		"claims": customClaims,
	}

	resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)

	common.AssertStatusCode(t, resp, http.StatusOK)
	common.AssertContentType(t, resp, "application/json")

	var tokenResp common.TokenResponse
	common.AssertJSONResponse(t, body, &tokenResp)

	if tokenResp.Token == "" {
		t.Fatal("❌ TOKEN GENERATION FAILED: Expected 'token' field in response")
	}

	// Parse and validate JWT to check custom audience
	token := common.AssertValidJWT(t, tokenResp.Token)
	expectedClaims := map[string]interface{}{
		"sub": "test-user",
		"aud": customAudience,
		"role": "admin",
	}
	common.AssertJWTClaims(t, token, expectedClaims)

	t.Log("✅ Token with custom audience generated successfully")

	// Test 2: Token generation without audience (should use default)
	t.Log("Testing token generation with default audience...")
	defaultClaims := map[string]interface{}{
		"sub": "test-user-default",
		"role": "user",
	}

	tokenReqDefault := map[string]interface{}{
		"claims": defaultClaims,
	}

	respDefault, bodyDefault := its.MakeRequest(t, "POST", "/generate-token", tokenReqDefault, nil)

	common.AssertStatusCode(t, respDefault, http.StatusOK)
	common.AssertContentType(t, respDefault, "application/json")

	var tokenRespDefault common.TokenResponse
	common.AssertJSONResponse(t, bodyDefault, &tokenRespDefault)

	if tokenRespDefault.Token == "" {
		t.Fatal("❌ TOKEN GENERATION FAILED: Expected 'token' field in response")
	}

	// Parse and validate JWT to check default audience
	tokenDefault := common.AssertValidJWT(t, tokenRespDefault.Token)
	expectedDefaultClaims := map[string]interface{}{
		"sub": "test-user-default",
		"aud": "integration-test-api", // Default audience from config in test environment
		"role": "user",
	}
	common.AssertJWTClaims(t, tokenDefault, expectedDefaultClaims)

	t.Log("✅ Token with default audience generated successfully")

	// Test 3: Token validation with custom audience should work
	t.Log("Testing token introspection with custom audience...")
	introspectReq := fmt.Sprintf("token=%s", tokenResp.Token)
	introspectHeaders := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	respIntrospect, bodyIntrospect := its.MakeRequest(t, "POST", "/introspect", introspectReq, introspectHeaders)

	common.AssertStatusCode(t, respIntrospect, http.StatusOK)
	common.AssertContentType(t, respIntrospect, "application/json")

	var introspectResp common.IntrospectionResponse
	common.AssertJSONResponse(t, bodyIntrospect, &introspectResp)

	if !introspectResp.Active {
		t.Error("❌ INTROSPECTION FAILED: Token with custom audience should be active")
	} else {
		t.Log("✅ Token with custom audience introspected successfully")
	}

	// Test 4: Invalid token generation with custom audience
	t.Log("Testing invalid token generation with custom audience...")
	invalidTokenReq := map[string]interface{}{
		"claims": map[string]interface{}{
			"sub": "test-user",
			"aud": "another-custom-service",
			"role": "tester",
		},
	}

	respInvalid, bodyInvalid := its.MakeRequest(t, "POST", "/generate-invalid-token", invalidTokenReq, nil)

	common.AssertStatusCode(t, respInvalid, http.StatusOK)
	common.AssertContentType(t, respInvalid, "application/json")

	var invalidTokenResp common.TokenResponse
	common.AssertJSONResponse(t, bodyInvalid, &invalidTokenResp)

	if invalidTokenResp.Token == "" {
		t.Fatal("❌ INVALID TOKEN GENERATION FAILED: Expected 'token' field in response")
	}

	// Parse JWT without validation to check custom audience in invalid token
	tokenInvalid := common.ParseJWTWithoutValidation(t, invalidTokenResp.Token)
	expectedInvalidClaims := map[string]interface{}{
		"sub": "test-user",
		"aud": "another-custom-service",
		"role": "tester",
	}
	common.AssertJWTClaims(t, tokenInvalid, expectedInvalidClaims)

	t.Log("✅ Invalid token with custom audience generated successfully")

	t.Log("✅ Audience Workflow Test completed successfully")
}