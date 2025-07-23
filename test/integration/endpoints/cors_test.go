package endpoints

import (
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestCORS tests CORS support
func TestCORS(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	// Test preflight OPTIONS request
	headers := map[string]string{
		"Origin":                         "http://localhost:3000",
		"Access-Control-Request-Method":  "POST",
		"Access-Control-Request-Headers": "Content-Type",
	}
	
	resp, _ := its.MakeRequest(t, "OPTIONS", "/generate-token", nil, headers)
	
	// Verify CORS headers
	if resp.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Fatal("❌ CORS TEST FAILED: Expected Access-Control-Allow-Origin header")
	}
	
	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Fatal("❌ CORS TEST FAILED: Expected Access-Control-Allow-Methods header")
	}
	
	if resp.Header.Get("Access-Control-Allow-Headers") == "" {
		t.Fatal("❌ CORS TEST FAILED: Expected Access-Control-Allow-Headers header")
	}
	
	t.Log("✅ CORS test passed")
}