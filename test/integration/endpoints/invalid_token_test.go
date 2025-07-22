package endpoints

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestInvalidToken tests introspection with invalid token
func TestInvalidToken(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	// Test with completely invalid token
	formData := url.Values{
		"token": {"invalid.token.here"},
	}
	
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	
	resp, body := its.MakeRequest(t, "POST", "/introspect", formData, headers)
	
	common.AssertStatusCode(t, resp, http.StatusOK)
	common.AssertContentType(t, resp, "application/json")
	
	var introspectResp common.IntrospectionResponse
	common.AssertJSONResponse(t, body, &introspectResp)
	
	// Verify token is marked as inactive
	if introspectResp.Active {
		t.Errorf("Expected active=false for invalid token, got %v", introspectResp.Active)
	}
	
	t.Log("Invalid token test passed")
}