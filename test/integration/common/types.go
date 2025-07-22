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
	Token   string `json:"token"`
	KeyID   string `json:"key_id"`
	TokenID string `json:"token_id"`
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
	Status string `json:"status"`
	Time   string `json:"time"`
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
	Keys []KeyInfo `json:"keys"`
}

// KeyInfo represents information about a key
type KeyInfo struct {
	ID        string `json:"id"`
	Algorithm string `json:"algorithm"`
	Use       string `json:"use"`
}