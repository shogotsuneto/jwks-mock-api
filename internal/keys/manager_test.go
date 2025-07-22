package keys

import (
	"strings"
	"testing"
)

// TestNewManager tests the creation of a new key manager
func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.keys == nil {
		t.Fatal("NewManager() did not initialize keys slice")
	}

	if len(manager.keys) != 0 {
		t.Errorf("Expected empty keys slice, got %d keys", len(manager.keys))
	}
}

// TestGenerateKeys tests key generation functionality
func TestGenerateKeys(t *testing.T) {
	tests := []struct {
		name     string
		keyIDs   []string
		expected int
	}{
		{
			name:     "Generate single key",
			keyIDs:   []string{"test-key-1"},
			expected: 1,
		},
		{
			name:     "Generate multiple keys",
			keyIDs:   []string{"test-key-1", "test-key-2", "test-key-3"},
			expected: 3,
		},
		{
			name:     "Generate with empty slice",
			keyIDs:   []string{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh manager for each test
			manager := NewManager()

			err := manager.GenerateKeys(tt.keyIDs)
			if err != nil {
				t.Fatalf("GenerateKeys() failed: %v", err)
			}

			if len(manager.keys) != tt.expected {
				t.Errorf("Expected %d keys, got %d", tt.expected, len(manager.keys))
			}

			// Validate each generated key
			for i, keyPair := range manager.keys {
				expectedKid := tt.keyIDs[i]

				// Check key ID
				if keyPair.Kid != expectedKid {
					t.Errorf("Expected key ID '%s', got '%s'", expectedKid, keyPair.Kid)
				}

				// Check private key
				if keyPair.PrivateKey == nil {
					t.Error("Private key is nil")
				} else {
					// Validate it's a valid RSA key
					if keyPair.PrivateKey.N == nil || keyPair.PrivateKey.D == nil {
						t.Error("Invalid RSA private key")
					}
				}

				// Check public key
				if keyPair.PublicKey == nil {
					t.Error("Public key is nil")
				} else {
					// Validate it's a valid RSA public key
					if keyPair.PublicKey.N == nil || keyPair.PublicKey.E == 0 {
						t.Error("Invalid RSA public key")
					}
				}

				// Check JWK
				if keyPair.JWK == nil {
					t.Error("JWK is nil")
				} else {
					// Validate JWK properties
					if keyPair.JWK.KeyID() != expectedKid {
						t.Errorf("Expected JWK KeyID '%s', got '%s'", expectedKid, keyPair.JWK.KeyID())
					}

					if keyPair.JWK.Algorithm().String() != "RS256" {
						t.Errorf("Expected JWK algorithm 'RS256', got '%s'", keyPair.JWK.Algorithm())
					}

					if keyPair.JWK.KeyUsage() != "sig" {
						t.Errorf("Expected JWK usage 'sig', got '%s'", keyPair.JWK.KeyUsage())
					}
				}
			}
		})
	}
}

// TestGetRandomKey tests random key retrieval
func TestGetRandomKey(t *testing.T) {
	tests := []struct {
		name        string
		keyIDs      []string
		shouldError bool
	}{
		{
			name:        "Get random key from multiple keys",
			keyIDs:      []string{"key-1", "key-2", "key-3"},
			shouldError: false,
		},
		{
			name:        "Get random key from single key",
			keyIDs:      []string{"single-key"},
			shouldError: false,
		},
		{
			name:        "Get random key from empty manager",
			keyIDs:      []string{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()
			
			if len(tt.keyIDs) > 0 {
				err := manager.GenerateKeys(tt.keyIDs)
				if err != nil {
					t.Fatalf("Failed to generate keys: %v", err)
				}
			}

			keyPair, err := manager.GetRandomKey()

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if keyPair != nil {
					t.Error("Expected nil key pair but got one")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if keyPair == nil {
					t.Error("Expected key pair but got nil")
				} else {
					// Validate the returned key pair
					if keyPair.Kid == "" {
						t.Error("Key pair has empty Kid")
					}
					if keyPair.PrivateKey == nil {
						t.Error("Key pair has nil PrivateKey")
					}
					if keyPair.PublicKey == nil {
						t.Error("Key pair has nil PublicKey")
					}
					if keyPair.JWK == nil {
						t.Error("Key pair has nil JWK")
					}

					// Verify the returned key is one of the expected keys
					found := false
					for _, expectedKid := range tt.keyIDs {
						if keyPair.Kid == expectedKid {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Returned key ID '%s' not in expected keys %v", keyPair.Kid, tt.keyIDs)
					}
				}
			}
		})
	}
}

// TestGetKeyByID tests key retrieval by ID
func TestGetKeyByID(t *testing.T) {
	manager := NewManager()
	keyIDs := []string{"test-key-1", "test-key-2", "test-key-3"}

	err := manager.GenerateKeys(keyIDs)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	tests := []struct {
		name        string
		requestedID string
		shouldError bool
	}{
		{
			name:        "Get existing key by ID",
			requestedID: "test-key-2",
			shouldError: false,
		},
		{
			name:        "Get first key by ID",
			requestedID: "test-key-1",
			shouldError: false,
		},
		{
			name:        "Get last key by ID",
			requestedID: "test-key-3",
			shouldError: false,
		},
		{
			name:        "Get non-existent key",
			requestedID: "non-existent-key",
			shouldError: true,
		},
		{
			name:        "Get key with empty ID",
			requestedID: "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPair, err := manager.GetKeyByID(tt.requestedID)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if keyPair != nil {
					t.Error("Expected nil key pair but got one")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if keyPair == nil {
					t.Error("Expected key pair but got nil")
				} else {
					if keyPair.Kid != tt.requestedID {
						t.Errorf("Expected key ID '%s', got '%s'", tt.requestedID, keyPair.Kid)
					}
					
					// Validate the key pair
					if keyPair.PrivateKey == nil {
						t.Error("Key pair has nil PrivateKey")
					}
					if keyPair.PublicKey == nil {
						t.Error("Key pair has nil PublicKey")
					}
					if keyPair.JWK == nil {
						t.Error("Key pair has nil JWK")
					}
				}
			}
		})
	}
}

// TestGetJWKS tests JWKS generation
func TestGetJWKS(t *testing.T) {
	tests := []struct {
		name           string
		keyIDs         []string
		expectedCount  int
	}{
		{
			name:          "Generate JWKS with multiple keys",
			keyIDs:        []string{"key-1", "key-2", "key-3"},
			expectedCount: 3,
		},
		{
			name:          "Generate JWKS with single key",
			keyIDs:        []string{"single-key"},
			expectedCount: 1,
		},
		{
			name:          "Generate JWKS with no keys",
			keyIDs:        []string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()

			if len(tt.keyIDs) > 0 {
				err := manager.GenerateKeys(tt.keyIDs)
				if err != nil {
					t.Fatalf("Failed to generate keys: %v", err)
				}
			}

			jwks, err := manager.GetJWKS()
			if err != nil {
				t.Fatalf("GetJWKS() failed: %v", err)
			}

			if jwks == nil {
				t.Fatal("GetJWKS() returned nil")
			}

			// Check the number of keys in the set
			keyCount := jwks.Len()

			if keyCount != tt.expectedCount {
				t.Errorf("Expected %d keys in JWKS, got %d", tt.expectedCount, keyCount)
			}

			// For now, just verify we got the right number of keys
			// More detailed validation would require deeper introspection
		})
	}
}

// TestGetAllKeyIDs tests getting all key IDs
func TestGetAllKeyIDs(t *testing.T) {
	tests := []struct {
		name     string
		keyIDs   []string
	}{
		{
			name:   "Get multiple key IDs",
			keyIDs: []string{"key-1", "key-2", "key-3"},
		},
		{
			name:   "Get single key ID",
			keyIDs: []string{"single-key"},
		},
		{
			name:   "Get no key IDs",
			keyIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()

			if len(tt.keyIDs) > 0 {
				err := manager.GenerateKeys(tt.keyIDs)
				if err != nil {
					t.Fatalf("Failed to generate keys: %v", err)
				}
			}

			allKeyIDs := manager.GetAllKeyIDs()

			if len(allKeyIDs) != len(tt.keyIDs) {
				t.Errorf("Expected %d key IDs, got %d", len(tt.keyIDs), len(allKeyIDs))
			}

			// Check that all expected key IDs are present
			for _, expectedID := range tt.keyIDs {
				found := false
				for _, actualID := range allKeyIDs {
					if actualID == expectedID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected key ID '%s' not found in result", expectedID)
				}
			}

			// Check that no unexpected key IDs are present
			for _, actualID := range allKeyIDs {
				found := false
				for _, expectedID := range tt.keyIDs {
					if actualID == expectedID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unexpected key ID '%s' found in result", actualID)
				}
			}
		})
	}
}

// TestGetKeyCount tests getting the key count
func TestGetKeyCount(t *testing.T) {
	tests := []struct {
		name          string
		keyIDs        []string
		expectedCount int
	}{
		{
			name:          "Count multiple keys",
			keyIDs:        []string{"key-1", "key-2", "key-3"},
			expectedCount: 3,
		},
		{
			name:          "Count single key",
			keyIDs:        []string{"single-key"},
			expectedCount: 1,
		},
		{
			name:          "Count no keys",
			keyIDs:        []string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()

			if len(tt.keyIDs) > 0 {
				err := manager.GenerateKeys(tt.keyIDs)
				if err != nil {
					t.Fatalf("Failed to generate keys: %v", err)
				}
			}

			count := manager.GetKeyCount()

			if count != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
			}
		})
	}
}

// TestKeyPairPEMConversion tests PEM conversion methods
func TestKeyPairPEMConversion(t *testing.T) {
	manager := NewManager()
	err := manager.GenerateKeys([]string{"test-key"})
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	keyPair, err := manager.GetKeyByID("test-key")
	if err != nil {
		t.Fatalf("Failed to get test key: %v", err)
	}

	// Test private key PEM conversion
	privateKeyPEM, err := keyPair.PrivateKeyToPEM()
	if err != nil {
		t.Errorf("PrivateKeyToPEM() failed: %v", err)
	}

	if privateKeyPEM == "" {
		t.Error("PrivateKeyToPEM() returned empty string")
	}

	// Check PEM format
	if !strings.Contains(privateKeyPEM, "-----BEGIN PRIVATE KEY-----") {
		t.Error("Private key PEM missing BEGIN header")
	}

	if !strings.Contains(privateKeyPEM, "-----END PRIVATE KEY-----") {
		t.Error("Private key PEM missing END header")
	}

	// Test public key PEM conversion
	publicKeyPEM, err := keyPair.PublicKeyToPEM()
	if err != nil {
		t.Errorf("PublicKeyToPEM() failed: %v", err)
	}

	if publicKeyPEM == "" {
		t.Error("PublicKeyToPEM() returned empty string")
	}

	// Check PEM format
	if !strings.Contains(publicKeyPEM, "-----BEGIN PUBLIC KEY-----") {
		t.Error("Public key PEM missing BEGIN header")
	}

	if !strings.Contains(publicKeyPEM, "-----END PUBLIC KEY-----") {
		t.Error("Public key PEM missing END header")
	}
}

// TestKeyPairConsistency tests that private and public keys are consistent
func TestKeyPairConsistency(t *testing.T) {
	manager := NewManager()
	err := manager.GenerateKeys([]string{"consistency-test"})
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	keyPair, err := manager.GetKeyByID("consistency-test")
	if err != nil {
		t.Fatalf("Failed to get test key: %v", err)
	}

	// Check that the public key in the keyPair matches the one derived from the private key
	derivedPublicKey := &keyPair.PrivateKey.PublicKey

	if keyPair.PublicKey.N.Cmp(derivedPublicKey.N) != 0 {
		t.Error("Public key N does not match private key's public key N")
	}

	if keyPair.PublicKey.E != derivedPublicKey.E {
		t.Error("Public key E does not match private key's public key E")
	}
}

// Benchmark tests

// BenchmarkGenerateKeys benchmarks key generation
func BenchmarkGenerateKeys(b *testing.B) {
	keyIDs := []string{"bench-key-1", "bench-key-2"}
	
	for i := 0; i < b.N; i++ {
		manager := NewManager()
		err := manager.GenerateKeys(keyIDs)
		if err != nil {
			b.Fatalf("GenerateKeys failed: %v", err)
		}
	}
}

// BenchmarkGetRandomKey benchmarks random key retrieval
func BenchmarkGetRandomKey(b *testing.B) {
	manager := NewManager()
	keyIDs := []string{"bench-key-1", "bench-key-2", "bench-key-3"}
	err := manager.GenerateKeys(keyIDs)
	if err != nil {
		b.Fatalf("Failed to generate keys: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetRandomKey()
		if err != nil {
			b.Fatalf("GetRandomKey failed: %v", err)
		}
	}
}

// BenchmarkGetJWKS benchmarks JWKS generation
func BenchmarkGetJWKS(b *testing.B) {
	manager := NewManager()
	keyIDs := []string{"bench-key-1", "bench-key-2", "bench-key-3"}
	err := manager.GenerateKeys(keyIDs)
	if err != nil {
		b.Fatalf("Failed to generate keys: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetJWKS()
		if err != nil {
			b.Fatalf("GetJWKS failed: %v", err)
		}
	}
}