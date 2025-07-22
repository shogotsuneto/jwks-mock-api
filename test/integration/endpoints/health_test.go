package endpoints

import (
	"net/http"
	"testing"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestHealth tests the health endpoint with real HTTP requests
func TestHealth(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	resp, body := its.MakeRequest(t, "GET", "/health", nil, nil)
	
	common.AssertStatusCode(t, resp, http.StatusOK)
	common.AssertContentType(t, resp, "application/json")
	
	var healthResp common.HealthResponse
	common.AssertJSONResponse(t, body, &healthResp)
	
	// Verify required fields
	if healthResp.Status != "ok" {
		t.Errorf("Expected status 'ok', got %s", healthResp.Status)
	}

	if healthResp.Service == "" {
		t.Error("Expected 'service' field in health response")
	}

	if len(healthResp.AvailableKeys) == 0 {
		t.Error("Expected 'available_keys' field with keys in health response")
	}
	
	t.Logf("Health check passed: %s", string(body))
}