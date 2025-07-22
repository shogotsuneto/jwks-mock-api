package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite() *IntegrationTestSuite {
	apiURL := os.Getenv("JWKS_API_URL")
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}

	timeout := TestTimeout
	if timeoutStr := os.Getenv("TEST_TIMEOUT"); timeoutStr != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsedTimeout
		}
	}

	return &IntegrationTestSuite{
		APIURL: apiURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// WaitForAPI waits for the API to be ready
func (its *IntegrationTestSuite) WaitForAPI(t *testing.T) {
	t.Helper()
	
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := its.HTTPClient.Get(its.APIURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Logf("API is ready after %d attempts", i+1)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		
		t.Logf("Waiting for API... attempt %d/%d", i+1, maxAttempts)
		time.Sleep(2 * time.Second)
	}
	
	t.Fatalf("API did not become ready after %d attempts", maxAttempts)
}

// MakeRequest is a helper to make HTTP requests
func (its *IntegrationTestSuite) MakeRequest(t *testing.T, method, endpoint string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	
	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
		case []byte:
			reqBody = bytes.NewReader(v)
		case url.Values:
			reqBody = strings.NewReader(v.Encode())
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}
			reqBody = bytes.NewReader(jsonBody)
		}
	}
	
	req, err := http.NewRequest(method, its.APIURL+endpoint, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Set default content type for POST requests
	if method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	resp, err := its.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request to %s %s: %v", method, endpoint, err)
	}
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	resp.Body.Close()
	
	return resp, respBody
}