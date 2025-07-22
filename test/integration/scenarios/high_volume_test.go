package scenarios

import (
	"fmt"
	"testing"
	"time"

	"github.com/shogotsuneto/jwks-mock-api/test/integration/common"
)

// TestHighVolumeTokenGeneration tests performance under load
func TestHighVolumeTokenGeneration(t *testing.T) {
	its := common.NewIntegrationTestSuite()
	its.WaitForAPI(t)
	
	t.Log("=== Starting High Volume Token Generation Test ===")
	
	tokenCount := 50
	successCount := 0
	start := time.Now()
	
	for i := 0; i < tokenCount; i++ {
		claims := map[string]interface{}{
			"sub":        fmt.Sprintf("load-test-user-%d", i),
			"batch":      "high-volume-test",
			"sequence":   i,
			"timestamp":  time.Now().Unix(),
			"load_test":  true,
		}
		
		tokenReq := map[string]interface{}{"claims": claims}
		resp, body := its.MakeRequest(t, "POST", "/generate-token", tokenReq, nil)
		
		if resp.StatusCode == 200 {
			successCount++
		} else {
			t.Logf("❌ Token %d failed: %d - %s", i, resp.StatusCode, string(body))
		}
		
		// Small delay to avoid overwhelming the server
		if i%10 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	duration := time.Since(start)
	
	if successCount < tokenCount {
		t.Errorf("❌ HIGH VOLUME TEST FAILED: Expected %d successful tokens, got %d", tokenCount, successCount)
	}
	
	successRate := float64(successCount) / float64(tokenCount) * 100
	if successRate < 95.0 {
		t.Errorf("❌ HIGH VOLUME TEST FAILED: Success rate too low: %.1f%% (expected > 95%%)", successRate)
	}
	
	t.Logf("✅ Generated %d/%d tokens successfully in %v (%.2f tokens/second)", 
		successCount, tokenCount, duration, float64(successCount)/duration.Seconds())
	t.Log("✅ High Volume Token Generation Test PASSED")
}