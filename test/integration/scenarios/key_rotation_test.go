package scenarios

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestKeyRotationSimulation tests multiple keys to simulate key rotation
func TestKeyRotationSimulation(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	t.Log("=== Starting Key Rotation Simulation Test ===")
	
	// First, check how many keys are available
	resp, body := its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)
	
	var jwks common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwks)
	
	t.Logf("Found %d keys in JWKS - simulating key rotation scenario", len(jwks.Keys))
	
	if len(jwks.Keys) < 2 {
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
		resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
		common.AssertStatusCode(t, resp, 200)
		
		var tokenResp common.TokenResponse
		common.AssertJSONResponse(t, body, &tokenResp)
		
		// Parse token to extract key ID
		parsedToken := common.AssertValidJWT(t, tokenResp.AccessToken)
		
		keyID := parsedToken.Header["kid"].(string)
		
		tokenResults[i] = map[string]interface{}{
			"token":  tokenResp.AccessToken,
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
		
		resp, body := its.MakeRequest(t, "POST", "/introspect", formData, headers)
		common.AssertStatusCode(t, resp, 200)
		
		var introspectResp common.IntrospectionResponse
		common.AssertJSONResponse(t, body, &introspectResp)
		
		if !introspectResp.Active {
			t.Errorf("Token %d should be active", index)
		}
		
		expectedSub := fmt.Sprintf("rotation-test-user-%d", index)
		if introspectResp.Sub != expectedSub {
			t.Errorf("Token %d: expected sub '%s', got '%s'", index, expectedSub, introspectResp.Sub)
		}
	}
	
	t.Log("=== Key Rotation Simulation Test PASSED ===")
}