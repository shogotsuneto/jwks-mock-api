package endpoints

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestKeyManagement tests the key management endpoints with real HTTP requests
func TestKeyManagement(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)

	t.Log("=== Starting Key Management Test ===")

	// Test 1: Get initial keys
	t.Log("Testing GET /keys to get initial keys...")
	resp, body := its.MakeRequest(t, "GET", "/keys", nil, nil)
	common.AssertStatusCode(t, resp, http.StatusOK)
	common.AssertContentType(t, resp, "application/json")

	var keysResp common.KeysResponse
	common.AssertJSONResponse(t, body, &keysResp)

	if keysResp.TotalKeys < 1 {
		t.Fatal("❌ KEY MANAGEMENT FAILED: Expected at least 1 initial key")
	}

	initialKeyCount := keysResp.TotalKeys
	t.Logf("✅ Initial key count: %d", initialKeyCount)

	// Test 2: Add a new key
	t.Log("Testing POST /keys to add a new key...")
	addKeyPayload := map[string]interface{}{
		"kid": "test-key-dynamic",
	}

	resp, body = its.MakeRequest(t, "POST", "/keys", addKeyPayload, map[string]string{
		"Content-Type": "application/json",
	})
	common.AssertStatusCode(t, resp, http.StatusCreated)
	common.AssertContentType(t, resp, "application/json")

	var addKeyResp common.AddKeyResponse
	common.AssertJSONResponse(t, body, &addKeyResp)

	if !addKeyResp.Success {
		t.Fatalf("❌ KEY MANAGEMENT FAILED: Expected success=true, got %v", addKeyResp.Success)
	}

	if addKeyResp.Kid != "test-key-dynamic" {
		t.Fatalf("❌ KEY MANAGEMENT FAILED: Expected kid='test-key-dynamic', got %s", addKeyResp.Kid)
	}

	t.Log("✅ Successfully added new key")

	// Test 3: Verify key was added by checking keys list
	t.Log("Testing GET /keys to verify key was added...")
	resp, body = its.MakeRequest(t, "GET", "/keys", nil, nil)
	common.AssertStatusCode(t, resp, http.StatusOK)

	common.AssertJSONResponse(t, body, &keysResp)

	if keysResp.TotalKeys != initialKeyCount+1 {
		t.Fatalf("❌ KEY MANAGEMENT FAILED: Expected %d keys after adding, got %d",
			initialKeyCount+1, keysResp.TotalKeys)
	}

	// Verify the new key is in the list
	foundKey := false
	for _, key := range keysResp.AvailableKeys {
		if kid, ok := key["kid"].(string); ok && kid == "test-key-dynamic" {
			foundKey = true
			break
		}
	}

	if !foundKey {
		t.Fatal("❌ KEY MANAGEMENT FAILED: New key 'test-key-dynamic' not found in keys list")
	}

	t.Log("✅ Successfully verified key was added")

	// Test 4: Try to add duplicate key (should fail)
	t.Log("Testing POST /keys with duplicate key ID (should fail)...")
	resp, body = its.MakeRequest(t, "POST", "/keys", addKeyPayload, map[string]string{
		"Content-Type": "application/json",
	})
	common.AssertStatusCode(t, resp, http.StatusConflict)

	common.AssertJSONResponse(t, body, &addKeyResp)

	if addKeyResp.Success {
		t.Fatal("❌ KEY MANAGEMENT FAILED: Expected success=false for duplicate key")
	}

	t.Log("✅ Successfully rejected duplicate key")

	// Test 5: Remove the key we added
	t.Log("Testing DELETE /keys/{kid} to remove the test key...")
	resp, body = its.MakeRequest(t, "DELETE", "/keys/test-key-dynamic", nil, nil)
	common.AssertStatusCode(t, resp, http.StatusOK)

	var removeKeyResp common.RemoveKeyResponse
	common.AssertJSONResponse(t, body, &removeKeyResp)

	if !removeKeyResp.Success {
		t.Fatalf("❌ KEY MANAGEMENT FAILED: Expected success=true for key removal, got %v", removeKeyResp.Success)
	}

	if removeKeyResp.Kid != "test-key-dynamic" {
		t.Fatalf("❌ KEY MANAGEMENT FAILED: Expected kid='test-key-dynamic', got %s", removeKeyResp.Kid)
	}

	t.Log("✅ Successfully removed test key")

	// Test 6: Verify key was removed
	t.Log("Testing GET /keys to verify key was removed...")
	resp, body = its.MakeRequest(t, "GET", "/keys", nil, nil)
	common.AssertStatusCode(t, resp, http.StatusOK)

	common.AssertJSONResponse(t, body, &keysResp)

	if keysResp.TotalKeys != initialKeyCount {
		t.Fatalf("❌ KEY MANAGEMENT FAILED: Expected %d keys after removing, got %d",
			initialKeyCount, keysResp.TotalKeys)
	}

	t.Log("✅ Successfully verified key was removed")

	// Test 7: Try to remove non-existent key (should fail)
	t.Log("Testing DELETE /keys/{kid} with non-existent key (should fail)...")
	resp, body = its.MakeRequest(t, "DELETE", "/keys/non-existent-key", nil, nil)
	common.AssertStatusCode(t, resp, http.StatusNotFound)

	common.AssertJSONResponse(t, body, &removeKeyResp)

	if removeKeyResp.Success {
		t.Fatal("❌ KEY MANAGEMENT FAILED: Expected success=false for non-existent key")
	}

	t.Log("✅ Successfully rejected removal of non-existent key")

	// Test 8: Try to remove last key (should fail if only one key remains)
	if initialKeyCount == 1 {
		t.Log("Testing DELETE /keys/{kid} for last remaining key (should fail)...")
		// Get the current key ID
		resp, body = its.MakeRequest(t, "GET", "/keys", nil, nil)
		common.AssertStatusCode(t, resp, http.StatusOK)
		common.AssertJSONResponse(t, body, &keysResp)

		if len(keysResp.AvailableKeys) > 0 {
			if kid, ok := keysResp.AvailableKeys[0]["kid"].(string); ok {
				resp, body = its.MakeRequest(t, "DELETE", fmt.Sprintf("/keys/%s", kid), nil, nil)
				common.AssertStatusCode(t, resp, http.StatusBadRequest)

				common.AssertJSONResponse(t, body, &removeKeyResp)

				if removeKeyResp.Success {
					t.Fatal("❌ KEY MANAGEMENT FAILED: Expected success=false when removing last key")
				}

				t.Log("✅ Successfully prevented removal of last key")
			}
		}
	}

	t.Log("✅ Key Management Test PASSED")
}

// TestKeyManagementInvalidRequests tests invalid requests to key management endpoints
func TestKeyManagementInvalidRequests(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)

	t.Log("=== Starting Key Management Invalid Requests Test ===")

	// Test 1: Add key without kid
	t.Log("Testing POST /keys without kid (should fail)...")
	invalidPayload := map[string]interface{}{}

	resp, body := its.MakeRequest(t, "POST", "/keys", invalidPayload, map[string]string{
		"Content-Type": "application/json",
	})
	common.AssertStatusCode(t, resp, http.StatusBadRequest)

	var addKeyResp common.AddKeyResponse
	common.AssertJSONResponse(t, body, &addKeyResp)

	if addKeyResp.Success {
		t.Fatal("❌ KEY MANAGEMENT FAILED: Expected success=false for missing kid")
	}

	t.Log("✅ Successfully rejected request without kid")

	// Test 2: Add key with empty kid
	t.Log("Testing POST /keys with empty kid (should fail)...")
	emptyKidPayload := map[string]interface{}{
		"kid": "",
	}

	resp, body = its.MakeRequest(t, "POST", "/keys", emptyKidPayload, map[string]string{
		"Content-Type": "application/json",
	})
	common.AssertStatusCode(t, resp, http.StatusBadRequest)

	common.AssertJSONResponse(t, body, &addKeyResp)

	if addKeyResp.Success {
		t.Fatal("❌ KEY MANAGEMENT FAILED: Expected success=false for empty kid")
	}

	t.Log("✅ Successfully rejected request with empty kid")

	// Test 3: Invalid JSON for add key
	t.Log("Testing POST /keys with invalid JSON (should fail)...")
	resp, _ = its.MakeRequestRaw(t, "POST", "/keys", []byte("invalid json"), map[string]string{
		"Content-Type": "application/json",
	})
	common.AssertStatusCode(t, resp, http.StatusBadRequest)

	t.Log("✅ Successfully rejected invalid JSON")

	t.Log("✅ Key Management Invalid Requests Test PASSED")
}