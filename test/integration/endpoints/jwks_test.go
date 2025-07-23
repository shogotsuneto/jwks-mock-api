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