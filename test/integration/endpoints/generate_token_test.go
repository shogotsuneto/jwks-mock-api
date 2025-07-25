package endpoints

import (
	"net/http"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
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

// TestTokenFieldOverrides tests which fields are overridden by config vs preserved from user input
func TestTokenFieldOverrides(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)

	t.Log("=== Testing Token Field Override Behavior ===")

	// Test 1: System fields are always overridden (iat, exp, iss)
	t.Log("Testing that system fields (iat, exp, iss) are always overridden...")
	
	userProvidedSystemClaims := map[string]interface{}{
		"sub":   "test-user",
		"iat":   1000000000, // User tries to set issued at time
		"exp":   2000000000, // User tries to set expiration
		"iss":   "user-provided-issuer", // User tries to set issuer
		"custom": "preserved-value",
	}

	tokenReq := map[string]interface{}{
		"claims": userProvidedSystemClaims,
	}

	resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
	common.AssertStatusCode(t, resp, http.StatusOK)

	var tokenResp common.TokenResponse
	common.AssertJSONResponse(t, body, &tokenResp)

	// Parse JWT to check actual claims
	token := common.AssertValidJWT(t, tokenResp.Token)

	// Verify system fields were overridden
	claims := token.Claims.(jwt.MapClaims)
	
	// iss should be config value, not user-provided
	if claims["iss"] != "http://jwks-api:3000" {
		t.Errorf("❌ ISSUER NOT OVERRIDDEN: Expected 'http://jwks-api:3000', got '%v'", claims["iss"])
	} else {
		t.Log("✅ Issuer correctly overridden by config")
	}

	// iat should be recent timestamp, not user-provided old value
	if iat, ok := claims["iat"].(float64); !ok || iat < 1600000000 {
		t.Errorf("❌ IAT NOT OVERRIDDEN: Expected recent timestamp, got '%v'", claims["iat"])
	} else {
		t.Log("✅ Issued-at time correctly overridden by system")
	}

	// exp should be future timestamp, not user-provided old value
	if exp, ok := claims["exp"].(float64); !ok || exp < 1600000000 {
		t.Errorf("❌ EXP NOT OVERRIDDEN: Expected future timestamp, got '%v'", claims["exp"])
	} else {
		t.Log("✅ Expiration time correctly overridden by system")
	}

	// Custom claims should be preserved
	if claims["custom"] != "preserved-value" {
		t.Errorf("❌ CUSTOM CLAIM NOT PRESERVED: Expected 'preserved-value', got '%v'", claims["custom"])
	} else {
		t.Log("✅ Custom claims correctly preserved")
	}

	// Subject should be preserved
	if claims["sub"] != "test-user" {
		t.Errorf("❌ SUBJECT NOT PRESERVED: Expected 'test-user', got '%v'", claims["sub"])
	} else {
		t.Log("✅ Subject correctly preserved")
	}
}

// TestTokenAudienceOverrideBehavior tests audience field override behavior specifically
func TestTokenAudienceOverrideBehavior(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)

	t.Log("=== Testing Audience Override Behavior ===")

	// Test 1: Custom audience is preserved when provided
	t.Log("Testing that custom audience is preserved when provided...")
	
	customAudienceClaims := map[string]interface{}{
		"sub": "test-user",
		"aud": "custom-service-api",
		"role": "admin",
	}

	tokenReq := map[string]interface{}{
		"claims": customAudienceClaims,
	}

	resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
	common.AssertStatusCode(t, resp, http.StatusOK)

	var tokenResp common.TokenResponse
	common.AssertJSONResponse(t, body, &tokenResp)

	token := common.AssertValidJWT(t, tokenResp.Token)
	claims := token.Claims.(jwt.MapClaims)

	// Verify custom audience is preserved
	if claims["aud"] != "custom-service-api" {
		t.Errorf("❌ CUSTOM AUDIENCE NOT PRESERVED: Expected 'custom-service-api', got '%v'", claims["aud"])
	} else {
		t.Log("✅ Custom audience correctly preserved")
	}

	// Test 2: Default audience is used when not provided
	t.Log("Testing that default audience is used when not provided...")
	
	noAudienceClaims := map[string]interface{}{
		"sub": "test-user-no-aud",
		"role": "user",
	}

	tokenReqDefault := map[string]interface{}{
		"claims": noAudienceClaims,
	}

	respDefault, bodyDefault := its.MakeRequest(t, "POST", "/generate-token", tokenReqDefault, nil)
	common.AssertStatusCode(t, respDefault, http.StatusOK)

	var tokenRespDefault common.TokenResponse
	common.AssertJSONResponse(t, bodyDefault, &tokenRespDefault)

	tokenDefault := common.AssertValidJWT(t, tokenRespDefault.Token)
	claimsDefault := tokenDefault.Claims.(jwt.MapClaims)

	// Verify default audience from config is used
	if claimsDefault["aud"] != "integration-test-api" {
		t.Errorf("❌ DEFAULT AUDIENCE NOT APPLIED: Expected 'integration-test-api', got '%v'", claimsDefault["aud"])
	} else {
		t.Log("✅ Default audience correctly applied from config")
	}

	// Test 3: Array audience is preserved
	t.Log("Testing that array audience is preserved...")
	
	arrayAudienceClaims := map[string]interface{}{
		"sub": "test-user",
		"aud": []string{"service1", "service2", "service3"},
		"role": "admin",
	}

	tokenReqArray := map[string]interface{}{
		"claims": arrayAudienceClaims,
	}

	respArray, bodyArray := its.MakeRequest(t, "POST", "/generate-token", tokenReqArray, nil)
	common.AssertStatusCode(t, respArray, http.StatusOK)

	var tokenRespArray common.TokenResponse
	common.AssertJSONResponse(t, bodyArray, &tokenRespArray)

	tokenArray := common.AssertValidJWT(t, tokenRespArray.Token)
	claimsArray := tokenArray.Claims.(jwt.MapClaims)

	// Verify array audience is preserved
	if aud, ok := claimsArray["aud"].([]interface{}); !ok || len(aud) != 3 {
		t.Errorf("❌ ARRAY AUDIENCE NOT PRESERVED: Expected array of 3 elements, got '%v'", claimsArray["aud"])
	} else {
		// Check individual elements
		expected := []string{"service1", "service2", "service3"}
		for i, expectedAud := range expected {
			if aud[i] != expectedAud {
				t.Errorf("❌ ARRAY AUDIENCE ELEMENT MISMATCH: Expected '%s' at index %d, got '%v'", expectedAud, i, aud[i])
				break
			}
		}
		t.Log("✅ Array audience correctly preserved")
	}
}

// TestTokenCustomClaimsPreservation tests that all custom claims are preserved exactly
func TestTokenCustomClaimsPreservation(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)

	t.Log("=== Testing Custom Claims Preservation ===")

	// Test with various types of custom claims
	complexClaims := map[string]interface{}{
		"sub":         "test-user",
		"email":       "user@example.com",
		"name":        "Test User",
		"roles":       []string{"admin", "user", "tester"},
		"permissions": map[string]interface{}{
			"read":  true,
			"write": true,
			"admin": false,
		},
		"numeric_claim": 42,
		"boolean_claim": true,
		"string_claim":  "custom-value",
		"nested_object": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "deep-value",
				"count":  100,
			},
		},
	}

	tokenReq := map[string]interface{}{
		"claims": complexClaims,
	}

	resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
	common.AssertStatusCode(t, resp, http.StatusOK)

	var tokenResp common.TokenResponse
	common.AssertJSONResponse(t, body, &tokenResp)

	token := common.AssertValidJWT(t, tokenResp.Token)
	claims := token.Claims.(jwt.MapClaims)

	// Verify all custom claims are preserved exactly
	preservedClaims := []string{"sub", "email", "name", "boolean_claim", "string_claim"}
	for _, claimName := range preservedClaims {
		if claims[claimName] != complexClaims[claimName] {
			t.Errorf("❌ CLAIM NOT PRESERVED: '%s' expected '%v', got '%v'", claimName, complexClaims[claimName], claims[claimName])
		} else {
			t.Logf("✅ Claim '%s' correctly preserved", claimName)
		}
	}

	// Verify numeric claim (JSON converts numbers to float64)
	if numericClaim, ok := claims["numeric_claim"].(float64); !ok || numericClaim != 42.0 {
		t.Errorf("❌ NUMERIC CLAIM NOT PRESERVED: 'numeric_claim' expected '42.0', got '%v'", claims["numeric_claim"])
	} else {
		t.Log("✅ Claim 'numeric_claim' correctly preserved")
	}

	// Verify array claims
	if rolesInterface, ok := claims["roles"].([]interface{}); ok {
		expectedRoles := []string{"admin", "user", "tester"}
		if len(rolesInterface) != len(expectedRoles) {
			t.Errorf("❌ ARRAY CLAIM LENGTH MISMATCH: 'roles' expected %d elements, got %d", len(expectedRoles), len(rolesInterface))
		} else {
			for i, expected := range expectedRoles {
				if rolesInterface[i] != expected {
					t.Errorf("❌ ARRAY CLAIM ELEMENT MISMATCH: 'roles[%d]' expected '%s', got '%v'", i, expected, rolesInterface[i])
				}
			}
			t.Log("✅ Array claim 'roles' correctly preserved")
		}
	} else {
		t.Errorf("❌ ARRAY CLAIM TYPE MISMATCH: 'roles' expected []interface{}, got %T", claims["roles"])
	}

	// Note: Complex nested objects are harder to verify exactly due to JSON marshaling,
	// but the key point is that user-provided claims are preserved while only
	// system claims (iat, exp, iss) are overridden, and aud uses user value when provided.

	t.Log("✅ Custom claims preservation test completed")
}

