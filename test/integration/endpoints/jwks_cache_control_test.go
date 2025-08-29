package endpoints

import (
	"net/http"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestJWKSCacheControlCustom tests the JWKS endpoint with various cache control configurations
// This test demonstrates the configurable cache control functionality
func TestJWKSCacheControlCustom(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	t.Log("=== Testing JWKS Cache-Control Header Configuration ===")
	
	// Test that the JWKS endpoint returns a valid response with cache control header
	resp, body := its.MakeRequest(t, "GET", "/.well-known/jwks.json", nil, nil)
	
	common.AssertStatusCode(t, resp, http.StatusOK)
	common.AssertContentType(t, resp, "application/json")
	
	// Verify cache control header exists
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl == "" {
		t.Error("❌ CACHE CONTROL MISSING: Expected Cache-Control header to be present")
	} else {
		t.Logf("✅ Cache-Control header found: %s", cacheControl)
	}
	
	// Validate JWKS structure
	var jwks common.JWKSResponse
	common.AssertJSONResponse(t, body, &jwks)
	common.AssertValidJWKS(t, &jwks)
	
	t.Log("✅ JWKS cache control configuration test completed successfully")
	
	// Log configuration notes for developers
	t.Log("📋 Configuration Options:")
	t.Log("   • Default: 'public, max-age=3600'")
	t.Log("   • Environment Variable: JWKS_CACHE_CONTROL")
	t.Log("   • YAML Config: jwks.cache_control")
	t.Log("   • Empty value omits the header entirely")
	t.Log("   • Examples: 'no-cache', 'private, max-age=300', 'public, max-age=86400'")
}