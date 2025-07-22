package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// KeyPair represents an RSA key pair with metadata
type KeyPair struct {
	Kid        string `json:"kid"`
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	JWK        jwk.Key
}

// Manager manages multiple key pairs for JWT signing
type Manager struct {
	keys []KeyPair
}

// NewManager creates a new key manager
func NewManager() *Manager {
	return &Manager{
		keys: make([]KeyPair, 0),
	}
}

// GenerateKeys generates the specified number of RSA key pairs
func (m *Manager) GenerateKeys(keyIDs []string) error {
	m.keys = make([]KeyPair, 0, len(keyIDs))

	for _, kid := range keyIDs {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate RSA key for %s: %w", kid, err)
		}

		// Create JWK from the private key
		jwkKey, err := jwk.FromRaw(privateKey)
		if err != nil {
			return fmt.Errorf("failed to create JWK for %s: %w", kid, err)
		}

		// Set the key ID and algorithm
		if err := jwkKey.Set(jwk.KeyIDKey, kid); err != nil {
			return fmt.Errorf("failed to set key ID for %s: %w", kid, err)
		}

		if err := jwkKey.Set(jwk.AlgorithmKey, "RS256"); err != nil {
			return fmt.Errorf("failed to set algorithm for %s: %w", kid, err)
		}

		if err := jwkKey.Set(jwk.KeyUsageKey, "sig"); err != nil {
			return fmt.Errorf("failed to set key usage for %s: %w", kid, err)
		}

		keyPair := KeyPair{
			Kid:        kid,
			PrivateKey: privateKey,
			PublicKey:  &privateKey.PublicKey,
			JWK:        jwkKey,
		}

		m.keys = append(m.keys, keyPair)
	}

	return nil
}

// GetRandomKey returns a random key pair for token signing
func (m *Manager) GetRandomKey() (*KeyPair, error) {
	if len(m.keys) == 0 {
		return nil, fmt.Errorf("no keys available")
	}

	// Generate a random index
	randomNum, err := rand.Int(rand.Reader, big.NewInt(int64(len(m.keys))))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %w", err)
	}

	index := randomNum.Int64()
	return &m.keys[index], nil
}

// GetKeyByID returns a key pair by its ID
func (m *Manager) GetKeyByID(kid string) (*KeyPair, error) {
	for i := range m.keys {
		if m.keys[i].Kid == kid {
			return &m.keys[i], nil
		}
	}
	return nil, fmt.Errorf("key not found: %s", kid)
}

// GetJWKS returns the JSON Web Key Set for all public keys
func (m *Manager) GetJWKS() (jwk.Set, error) {
	set := jwk.NewSet()

	for _, keyPair := range m.keys {
		// Create a public key JWK from the private key JWK
		pubKey, err := jwk.PublicKeyOf(keyPair.JWK)
		if err != nil {
			return nil, fmt.Errorf("failed to extract public key for %s: %w", keyPair.Kid, err)
		}

		if err := set.AddKey(pubKey); err != nil {
			return nil, fmt.Errorf("failed to add public key to set for %s: %w", keyPair.Kid, err)
		}
	}

	return set, nil
}

// GetAllKeyIDs returns all available key IDs
func (m *Manager) GetAllKeyIDs() []string {
	keyIDs := make([]string, len(m.keys))
	for i, key := range m.keys {
		keyIDs[i] = key.Kid
	}
	return keyIDs
}

// GetKeyCount returns the number of available keys
func (m *Manager) GetKeyCount() int {
	return len(m.keys)
}

// PrivateKeyToPEM converts a private key to PEM format
func (kp *KeyPair) PrivateKeyToPEM() (string, error) {
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(kp.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	})

	return string(privateKeyPEM), nil
}

// PublicKeyToPEM converts a public key to PEM format
func (kp *KeyPair) PublicKeyToPEM() (string, error) {
	publicKeyDER, err := x509.MarshalPKIXPublicKey(kp.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	return string(publicKeyPEM), nil
}
