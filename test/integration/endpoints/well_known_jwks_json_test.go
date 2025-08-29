package endpoints

import (
	"net/http"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestJWKS tests the JWKS endpoint
func TestJWKS(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	resp, body := its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	
	common.AssertStatusCode(t, resp, http.StatusOK)
	common.AssertContentType(t, resp, "application/json")
	
	var jwks common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwks)
	
	// Validate JWKS structure using common assertion
	common.AssertValidJWKS(t, &jwks)
	
	t.Logf("✅ JWKS validation passed with %d keys", len(jwks.Keys))
}

// TestJWKSCacheControl tests that the JWKS endpoint respects cache control configuration
func TestJWKSCacheControl(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	t.Log("=== Testing JWKS Cache-Control Header ===" )
	
	// Test 1: Default cache control header
	t.Log("Testing default cache control header...")
	resp, body := its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	
	common.AssertStatusCode(t, resp, http.StatusOK)
	common.AssertContentType(t, resp, "application/json")
	
	// Verify default cache control header is present
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl == "" {
		t.Error("❌ CACHE CONTROL MISSING: Expected Cache-Control header to be present")
	} else {
		t.Logf("✅ Cache-Control header found: %s", cacheControl)
	}
	
	// Verify default value (should be "public, max-age=3600" unless overridden)
	expectedDefault := "public, max-age=3600"
	if cacheControl != expectedDefault {
		t.Logf("ℹ️  Cache-Control differs from default (may be customized): Expected '%s', got '%s'", expectedDefault, cacheControl)
	} else {
		t.Log("✅ Default cache control value confirmed")
	}
	
	// Validate JWKS structure
	var jwks common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwks)
	common.AssertValidJWKS(t, &jwks)
	
	t.Log("✅ JWKS cache control test completed successfully")
}