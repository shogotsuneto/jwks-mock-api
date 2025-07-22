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
		parsedToken := common.AssertValidJWT(t, tokenResp.Token)
		
		keyID := parsedToken.Header["kid"].(string)
		
		tokenResults[i] = map[string]interface{}{
			"token":  tokenResp.Token,
			"key_id": keyID,
			"index":  i,
		}
	}
	
	// Verify that different keys are being used (at least some variation)
	keyIDs := make(map[string]int)
	for _, result := range tokenResults {
		keyID := result["key_id"].(string)
		keyIDs[keyID]++
	}
	
	if len(keyIDs) < 2 {
		t.Log("⚠️  All tokens used the same key - this might be expected for smaller key sets")
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
			t.Errorf("❌ KEY ROTATION TEST FAILED: Token %d should be active", index)
		}
		
		expectedSub := fmt.Sprintf("rotation-test-user-%d", index)
		if introspectResp.Sub != expectedSub {
			t.Errorf("❌ KEY ROTATION TEST FAILED: Token %d: expected sub '%s', got '%s'", index, expectedSub, introspectResp.Sub)
		}
	}
	
	t.Logf("✅ Key rotation test passed - %d keys tested, %d different key IDs used", len(jwks.Keys), len(keyIDs))
	t.Log("✅ Key Rotation Simulation Test PASSED")
}