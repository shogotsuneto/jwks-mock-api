package endpoints

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestIntrospection tests token introspection endpoint
func TestIntrospection(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	// First generate a token
	claims := map[string]interface{}{
		"sub":    "introspection-test-user",
		"scope":  "read write",
		"client": "test-client",
	}
	
	tokenReq := map[string]interface{}{
		"claims": claims,
	}
	
	resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
	common.AssertStatusCode(t, resp, http.StatusOK)
	
	var tokenResp common.TokenResponse
	common.AssertJSONResponse(t, body, &tokenResp)
	
	// Test introspection with form data
	formData := url.Values{
		"token": {tokenResp.AccessToken},
	}
	
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	
	resp, body = its.MakeRequest(t, "POST", "/introspect", formData, headers)
	
	common.AssertStatusCode(t, resp, http.StatusOK)
	common.AssertContentType(t, resp, "application/json")
	
	var introspectResp common.IntrospectionResponse
	common.AssertJSONResponse(t, body, &introspectResp)
	
	// Verify introspection response
	if !introspectResp.Active {
		t.Errorf("Expected active=true, got %v", introspectResp.Active)
	}
	
	if introspectResp.Sub != "introspection-test-user" {
		t.Errorf("Expected sub 'introspection-test-user', got %s", introspectResp.Sub)
	}
	
	t.Log("Token introspection test passed")
}