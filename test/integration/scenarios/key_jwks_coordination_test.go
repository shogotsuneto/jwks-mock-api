package scenarios

import (
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestKeyJWKSCoordination tests the coordination between key management endpoints and JWKS endpoint
// Adds and removes keys multiple times and verifies changes are immediately reflected in JWKS
func TestKeyJWKSCoordination(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)

	// Get initial JWKS state
	resp, body := its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	var initialJWKS common.JWKSResponse
	common.AssertJSONResponse(t, body, &initialJWKS)
	initialKeyCount := len(initialJWKS.Keys)

	// Helper function to add a key and verify it appears in JWKS
	addKeyAndVerify := func(kid string) {
		// Add key
		addPayload := map[string]interface{}{"kid": kid}
		resp, body := its.MakeRequest(t, "POST", "/keys", addPayload, map[string]string{
			"Content-Type": "application/json",
		})
		common.AssertStatusCode(t, resp, 201)

		var addResp common.AddKeyResponse
		common.AssertJSONResponse(t, body, &addResp)
		if !addResp.Success {
			t.Fatalf("Failed to add key %s", kid)
		}

		// Verify key appears in JWKS
		resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
		common.AssertStatusCode(t, resp, 200)

		var jwks common.JWKSResponse
		common.AssertJSONResponse(t, body, &jwks)

		found := false
		for _, key := range jwks.Keys {
			if key.KeyID == kid {
				found = true
				// Verify key properties
				if key.Kty != "RSA" || key.Use != "sig" || key.Alg != "RS256" {
					t.Errorf("Invalid key properties for %s", kid)
				}
				break
			}
		}
		if !found {
			t.Fatalf("Key %s not found in JWKS after addition", kid)
		}
	}

	// Helper function to remove a key and verify it disappears from JWKS
	removeKeyAndVerify := func(kid string) {
		// Remove key
		resp, body := its.MakeRequest(t, "DELETE", "/keys/"+kid, nil, nil)
		common.AssertStatusCode(t, resp, 200)

		var removeResp common.RemoveKeyResponse
		common.AssertJSONResponse(t, body, &removeResp)
		if !removeResp.Success {
			t.Fatalf("Failed to remove key %s", kid)
		}

		// Verify key disappears from JWKS
		resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
		common.AssertStatusCode(t, resp, 200)

		var jwks common.JWKSResponse
		common.AssertJSONResponse(t, body, &jwks)

		for _, key := range jwks.Keys {
			if key.KeyID == kid {
				t.Fatalf("Key %s still found in JWKS after removal", kid)
			}
		}
	}

	// Helper function to verify JWKS has expected key count
	verifyKeyCount := func(expected int) {
		resp, body := its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
		common.AssertStatusCode(t, resp, 200)

		var jwks common.JWKSResponse
		common.AssertJSONResponse(t, body, &jwks)

		if len(jwks.Keys) != expected {
			t.Fatalf("Expected %d keys in JWKS, got %d", expected, len(jwks.Keys))
		}
	}

	// Test cycle 1: Add key1, then key2, then remove both
	addKeyAndVerify("test-key-1")
	verifyKeyCount(initialKeyCount + 1)

	addKeyAndVerify("test-key-2")
	verifyKeyCount(initialKeyCount + 2)

	removeKeyAndVerify("test-key-1")
	verifyKeyCount(initialKeyCount + 1)

	removeKeyAndVerify("test-key-2")
	verifyKeyCount(initialKeyCount)

	// Test cycle 2: Repeat with different key to demonstrate multiple operations
	addKeyAndVerify("test-key-3")
	verifyKeyCount(initialKeyCount + 1)

	removeKeyAndVerify("test-key-3")
	verifyKeyCount(initialKeyCount)

	t.Log("✅ Key-JWKS coordination verified across multiple add/remove cycles")
}