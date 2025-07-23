package scenarios

import (
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestKeyJWKSCoordination tests the coordination between key management endpoints and JWKS endpoint
// This test adds and removes keys multiple times and verifies that changes are immediately reflected in the JWKS endpoint
func TestKeyJWKSCoordination(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)

	t.Log("=== Starting Key-JWKS Coordination Test ===")

	// Step 1: Get initial JWKS state
	t.Log("Getting initial JWKS state...")
	resp, body := its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)
	common.AssertContentType(t, resp, "application/json")

	var initialJWKS common.JWKSResponse
	common.AssertJSONResponse(t, body, &initialJWKS)

	initialKeyCount := len(initialJWKS.Keys)
	t.Logf("✅ Initial JWKS has %d keys", initialKeyCount)

	// Collect initial key IDs for verification
	initialKeyIDs := make(map[string]bool)
	for _, key := range initialJWKS.Keys {
		initialKeyIDs[key.KeyID] = true
	}

	// Step 2: Add first test key
	t.Log("Adding first test key 'test-coordination-key-1'...")
	addKeyPayload := map[string]interface{}{
		"kid": "test-coordination-key-1",
	}

	resp, body = its.MakeRequest(t, "POST", "/keys", addKeyPayload, map[string]string{
		"Content-Type": "application/json",
	})
	common.AssertStatusCode(t, resp, 201)

	var addKeyResp common.AddKeyResponse
	common.AssertJSONResponse(t, body, &addKeyResp)

	if !addKeyResp.Success {
		t.Fatalf("❌ COORDINATION FAILED: Expected success=true for first key addition, got %v", addKeyResp.Success)
	}

	// Step 3: Verify JWKS immediately reflects the new key
	t.Log("Verifying JWKS reflects first key addition...")
	resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	var jwksAfterAdd1 common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwksAfterAdd1)

	if len(jwksAfterAdd1.Keys) != initialKeyCount+1 {
		t.Fatalf("❌ COORDINATION FAILED: Expected %d keys in JWKS after first addition, got %d",
			initialKeyCount+1, len(jwksAfterAdd1.Keys))
	}

	// Verify the new key is present
	foundKey1 := false
	for _, key := range jwksAfterAdd1.Keys {
		if key.KeyID == "test-coordination-key-1" {
			foundKey1 = true
			// Verify key properties
			if key.Kty != "RSA" {
				t.Errorf("❌ COORDINATION FAILED: Expected kty='RSA', got '%s'", key.Kty)
			}
			if key.Use != "sig" {
				t.Errorf("❌ COORDINATION FAILED: Expected use='sig', got '%s'", key.Use)
			}
			if key.Alg != "RS256" {
				t.Errorf("❌ COORDINATION FAILED: Expected alg='RS256', got '%s'", key.Alg)
			}
			break
		}
	}

	if !foundKey1 {
		t.Fatal("❌ COORDINATION FAILED: Key 'test-coordination-key-1' not found in JWKS after addition")
	}

	t.Log("✅ JWKS correctly reflects first key addition")

	// Step 4: Add second test key
	t.Log("Adding second test key 'test-coordination-key-2'...")
	addKeyPayload2 := map[string]interface{}{
		"kid": "test-coordination-key-2",
	}

	resp, body = its.MakeRequest(t, "POST", "/keys", addKeyPayload2, map[string]string{
		"Content-Type": "application/json",
	})
	common.AssertStatusCode(t, resp, 201)

	common.AssertJSONResponse(t, body, &addKeyResp)

	if !addKeyResp.Success {
		t.Fatalf("❌ COORDINATION FAILED: Expected success=true for second key addition, got %v", addKeyResp.Success)
	}

	// Step 5: Verify JWKS reflects both test keys
	t.Log("Verifying JWKS reflects both test keys...")
	resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	var jwksAfterAdd2 common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwksAfterAdd2)

	if len(jwksAfterAdd2.Keys) != initialKeyCount+2 {
		t.Fatalf("❌ COORDINATION FAILED: Expected %d keys in JWKS after second addition, got %d",
			initialKeyCount+2, len(jwksAfterAdd2.Keys))
	}

	// Verify both keys are present
	foundKey1After2 := false
	foundKey2 := false
	for _, key := range jwksAfterAdd2.Keys {
		if key.KeyID == "test-coordination-key-1" {
			foundKey1After2 = true
		}
		if key.KeyID == "test-coordination-key-2" {
			foundKey2 = true
		}
	}

	if !foundKey1After2 || !foundKey2 {
		t.Fatal("❌ COORDINATION FAILED: Both test keys should be present in JWKS after second addition")
	}

	t.Log("✅ JWKS correctly reflects both key additions")

	// Step 6: Remove first test key
	t.Log("Removing first test key 'test-coordination-key-1'...")
	resp, body = its.MakeRequest(t, "DELETE", "/keys/test-coordination-key-1", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	var removeKeyResp common.RemoveKeyResponse
	common.AssertJSONResponse(t, body, &removeKeyResp)

	if !removeKeyResp.Success {
		t.Fatalf("❌ COORDINATION FAILED: Expected success=true for first key removal, got %v", removeKeyResp.Success)
	}

	// Step 7: Verify JWKS immediately reflects first key removal (should only have second key)
	t.Log("Verifying JWKS reflects first key removal...")
	resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	var jwksAfterRemove1 common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwksAfterRemove1)

	if len(jwksAfterRemove1.Keys) != initialKeyCount+1 {
		t.Fatalf("❌ COORDINATION FAILED: Expected %d keys in JWKS after first removal, got %d",
			initialKeyCount+1, len(jwksAfterRemove1.Keys))
	}

	// Verify first key is gone and second key remains
	stillFoundKey1 := false
	stillFoundKey2 := false
	for _, key := range jwksAfterRemove1.Keys {
		if key.KeyID == "test-coordination-key-1" {
			stillFoundKey1 = true
		}
		if key.KeyID == "test-coordination-key-2" {
			stillFoundKey2 = true
		}
	}

	if stillFoundKey1 {
		t.Fatal("❌ COORDINATION FAILED: Key 'test-coordination-key-1' should be removed from JWKS")
	}

	if !stillFoundKey2 {
		t.Fatal("❌ COORDINATION FAILED: Key 'test-coordination-key-2' should still be present in JWKS")
	}

	t.Log("✅ JWKS correctly reflects first key removal")

	// Step 8: Remove second test key
	t.Log("Removing second test key 'test-coordination-key-2'...")
	resp, body = its.MakeRequest(t, "DELETE", "/keys/test-coordination-key-2", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	common.AssertJSONResponse(t, body, &removeKeyResp)

	if !removeKeyResp.Success {
		t.Fatalf("❌ COORDINATION FAILED: Expected success=true for second key removal, got %v", removeKeyResp.Success)
	}

	// Step 9: Verify JWKS is back to initial state
	t.Log("Verifying JWKS is back to initial state...")
	resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	var jwksAfterRemove2 common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwksAfterRemove2)

	if len(jwksAfterRemove2.Keys) != initialKeyCount {
		t.Fatalf("❌ COORDINATION FAILED: Expected %d keys in JWKS after all removals (back to initial), got %d",
			initialKeyCount, len(jwksAfterRemove2.Keys))
	}

	// Verify neither test key remains
	finalFoundKey1 := false
	finalFoundKey2 := false
	for _, key := range jwksAfterRemove2.Keys {
		if key.KeyID == "test-coordination-key-1" {
			finalFoundKey1 = true
		}
		if key.KeyID == "test-coordination-key-2" {
			finalFoundKey2 = true
		}
	}

	if finalFoundKey1 || finalFoundKey2 {
		t.Fatal("❌ COORDINATION FAILED: Both test keys should be completely removed from JWKS")
	}

	// Verify original keys are still present
	for originalKid := range initialKeyIDs {
		found := false
		for _, key := range jwksAfterRemove2.Keys {
			if key.KeyID == originalKid {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("❌ COORDINATION FAILED: Original key '%s' should still be present after test", originalKid)
		}
	}

	t.Log("✅ JWKS correctly back to initial state")

	// Step 10: Repeat the cycle one more time to demonstrate "multiple times"
	t.Log("Repeating add/remove cycle to demonstrate multiple operations...")

	// Add third test key
	t.Log("Adding third test key 'test-coordination-key-3'...")
	addKeyPayload3 := map[string]interface{}{
		"kid": "test-coordination-key-3",
	}

	resp, body = its.MakeRequest(t, "POST", "/keys", addKeyPayload3, map[string]string{
		"Content-Type": "application/json",
	})
	common.AssertStatusCode(t, resp, 201)

	// Verify JWKS reflects third key
	resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	var jwksAfterAdd3 common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwksAfterAdd3)

	if len(jwksAfterAdd3.Keys) != initialKeyCount+1 {
		t.Fatalf("❌ COORDINATION FAILED: Expected %d keys after third addition, got %d",
			initialKeyCount+1, len(jwksAfterAdd3.Keys))
	}

	foundKey3 := false
	for _, key := range jwksAfterAdd3.Keys {
		if key.KeyID == "test-coordination-key-3" {
			foundKey3 = true
			break
		}
	}

	if !foundKey3 {
		t.Fatal("❌ COORDINATION FAILED: Key 'test-coordination-key-3' not found in JWKS")
	}

	t.Log("✅ Third key addition reflected in JWKS")

	// Remove third test key
	t.Log("Removing third test key 'test-coordination-key-3'...")
	resp, body = its.MakeRequest(t, "DELETE", "/keys/test-coordination-key-3", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	// Verify JWKS reflects third key removal
	resp, body = its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	common.AssertStatusCode(t, resp, 200)

	var finalJWKS common.JWKSResponse
	common.AssertJSONResponse(t, body, &finalJWKS)

	if len(finalJWKS.Keys) != initialKeyCount {
		t.Fatalf("❌ COORDINATION FAILED: Expected %d keys after final removal, got %d",
			initialKeyCount, len(finalJWKS.Keys))
	}

	finalFoundKey3 := false
	for _, key := range finalJWKS.Keys {
		if key.KeyID == "test-coordination-key-3" {
			finalFoundKey3 = true
			break
		}
	}

	if finalFoundKey3 {
		t.Fatal("❌ COORDINATION FAILED: Key 'test-coordination-key-3' should be removed from JWKS")
	}

	t.Log("✅ Third key removal reflected in JWKS")

	t.Log("✅ Key-JWKS Coordination Test PASSED")
	t.Log("✅ Successfully demonstrated multiple add/remove operations with immediate JWKS reflection")
}