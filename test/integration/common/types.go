package common

import (
	"net/http"
	"time"
)

const (
	DefaultAPIURL = "http://localhost:3001"
	TestTimeout   = 30 * time.Second
)

// IntegrationTestSuite manages the integration test environment
type IntegrationTestSuite struct {
	APIURL     string
	HTTPClient *http.Client
}

// TokenResponse represents the response from token generation endpoint
type TokenResponse struct {
	Token      string                 `json:"token"`
	ExpiresIn  int                    `json:"expires_in"`
	KeyID      string                 `json:"key_id"`
	RawRequest map[string]interface{} `json:"raw_request"`
}

// IntrospectionResponse represents the response from token introspection endpoint
type IntrospectionResponse struct {
	Active   bool                   `json:"active"`
	Sub      string                 `json:"sub"`
	Aud      interface{}            `json:"aud"`
	Iss      string                 `json:"iss"`
	Exp      int64                  `json:"exp"`
	Iat      int64                  `json:"iat"`
	TokenUse string                 `json:"token_use"`
	Claims   map[string]interface{} `json:"claims"`
}

// HealthResponse represents the response from health endpoint
type HealthResponse struct {
	Status        string   `json:"status"`
	Service       string   `json:"service"`
	AvailableKeys []string `json:"available_keys"`
}

// JWKSResponse represents the response from JWKS endpoint
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	KeyID string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// KeysResponse represents the response from keys endpoint  
type KeysResponse struct {
	TotalKeys     int                      `json:"total_keys"`
	AvailableKeys []map[string]interface{} `json:"available_keys"`
}

// AddKeyResponse represents the response from adding a key
type AddKeyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Kid     string `json:"kid"`
}

// RemoveKeyResponse represents the response from removing a key
type RemoveKeyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Kid     string `json:"kid"`
}

// KeyInfo represents information about a key
type KeyInfo struct {
	ID        string `json:"id"`
	Algorithm string `json:"algorithm"`
	Use       string `json:"use"`
}