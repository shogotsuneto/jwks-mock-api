package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shogotsuneto/jwks-mock-api/internal/keys"
	"github.com/shogotsuneto/jwks-mock-api/pkg/config"
)

// Handler contains the HTTP handlers for the JWKS service
type Handler struct {
	config     *config.Config
	keyManager *keys.Manager
}

// New creates a new handler instance
func New(cfg *config.Config, keyManager *keys.Manager) *Handler {
	return &Handler{
		config:     cfg,
		keyManager: keyManager,
	}
}



// TokenResponse represents a token generation response
type TokenResponse struct {
	AccessToken string                 `json:"access_token"`
	TokenType   string                 `json:"token_type"`
	ExpiresIn   string                 `json:"expires_in"`
	KeyID       string                 `json:"key_id"`
	RawRequest  map[string]interface{} `json:"raw_request"`
}

// ValidationRequest represents a token validation request
type ValidationRequest struct {
	Token string `json:"token"`
}

// ValidationResponse represents a token validation response
type ValidationResponse struct {
	Valid   bool                   `json:"valid"`
	Decoded map[string]interface{} `json:"decoded,omitempty"`
	KeyID   string                 `json:"key_id,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status        string   `json:"status"`
	Service       string   `json:"service"`
	AvailableKeys []string `json:"available_keys"`
}

// KeysResponse represents a keys information response
type KeysResponse struct {
	TotalKeys     int                      `json:"total_keys"`
	AvailableKeys []map[string]interface{} `json:"available_keys"`
}

// QuickTokenResponse represents a quick token response
type QuickTokenResponse struct {
	Token       string `json:"token"`
	KeyID       string `json:"key_id"`
	CurlExample string `json:"curl_example"`
}

// JWKS returns the JSON Web Key Set
func (h *Handler) JWKS(w http.ResponseWriter, r *http.Request) {
	jwks, err := h.keyManager.GetJWKS()
	if err != nil {
		log.Printf("Error generating JWKS: %v", err)
		http.Error(w, `{"error": "Failed to generate JWKS"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		log.Printf("Error encoding JWKS response: %v", err)
		http.Error(w, `{"error": "Failed to encode JWKS"}`, http.StatusInternalServerError)
		return
	}
}

// GenerateToken generates a new JWT token with dynamic claims
func (h *Handler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	// Parse the request body as a generic map to capture all fields
	var rawRequest map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawRequest); err != nil {
		http.Error(w, `{"error": "Invalid JSON request"}`, http.StatusBadRequest)
		return
	}

	// Extract expiresIn if present, default to "1h"
	expiresIn := "1h"
	if exp, ok := rawRequest["expiresIn"].(string); ok && exp != "" {
		expiresIn = exp
	}

	// Remove expiresIn from claims as it's not a JWT claim
	delete(rawRequest, "expiresIn")

	// Set default claims if the request is empty
	if len(rawRequest) == 0 {
		rawRequest = map[string]interface{}{
			"sub":   "test-user",
			"email": "test@example.com",
			"name":  "Test User",
			"roles": []string{"user"},
		}
	}

	// Get a random key for signing
	keyPair, err := h.keyManager.GetRandomKey()
	if err != nil {
		log.Printf("Error getting random key: %v", err)
		http.Error(w, `{"error": "Failed to get signing key"}`, http.StatusInternalServerError)
		return
	}

	// Calculate expiration
	exp := time.Now().Add(time.Hour) // Default 1 hour
	if expiresIn != "1h" {
		if seconds, err := strconv.Atoi(expiresIn); err == nil {
			exp = time.Now().Add(time.Duration(seconds) * time.Second)
		}
	}

	// Create claims starting with the dynamic claims from the request
	claims := jwt.MapClaims{}
	for key, value := range rawRequest {
		claims[key] = value
	}

	// Add standard JWT claims (these override any user-provided values for security)
	claims["iat"] = time.Now().Unix()
	claims["exp"] = exp.Unix()
	claims["iss"] = h.config.JWT.Issuer
	claims["aud"] = h.config.JWT.Audience

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyPair.Kid

	// Sign token
	tokenString, err := token.SignedString(keyPair.PrivateKey)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		http.Error(w, `{"error": "Failed to sign token"}`, http.StatusInternalServerError)
		return
	}

	response := TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		KeyID:       keyPair.Kid,
		RawRequest:  rawRequest, // Include all the dynamic request data
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateToken validates a JWT token
func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON request"}`, http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		response := ValidationResponse{
			Valid: false,
			Error: "Token is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse token to get the kid
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the kid from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing key ID in token header")
		}

		// Find the corresponding key
		keyPair, err := h.keyManager.GetKeyByID(kid)
		if err != nil {
			return nil, fmt.Errorf("key not found for kid: %s", kid)
		}

		return keyPair.PublicKey, nil
	})

	response := ValidationResponse{}

	if err != nil {
		response.Valid = false
		response.Error = err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
	} else if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Validate issuer and audience
		if claims["iss"] != h.config.JWT.Issuer {
			response.Valid = false
			response.Error = "invalid issuer"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
		} else if claims["aud"] != h.config.JWT.Audience {
			response.Valid = false
			response.Error = "invalid audience"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			response.Valid = true
			response.Decoded = claims
			response.KeyID = token.Header["kid"].(string)
			w.Header().Set("Content-Type", "application/json")
		}
	} else {
		response.Valid = false
		response.Error = "invalid token"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
	}

	json.NewEncoder(w).Encode(response)
}

// QuickToken generates a quick token for testing
func (h *Handler) QuickToken(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = "quick-user"
	}

	// Get a random key for signing
	keyPair, err := h.keyManager.GetRandomKey()
	if err != nil {
		log.Printf("Error getting random key: %v", err)
		http.Error(w, `{"error": "Failed to get signing key"}`, http.StatusInternalServerError)
		return
	}

	// Create claims
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": fmt.Sprintf("%s@example.com", userID),
		"name":  fmt.Sprintf("%s User", userID),
		"roles": []string{"user"},
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iss":   h.config.JWT.Issuer,
		"aud":   h.config.JWT.Audience,
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyPair.Kid

	// Sign token
	tokenString, err := token.SignedString(keyPair.PrivateKey)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		http.Error(w, `{"error": "Failed to sign token"}`, http.StatusInternalServerError)
		return
	}

	response := QuickTokenResponse{
		Token:       tokenString,
		KeyID:       keyPair.Kid,
		CurlExample: fmt.Sprintf(`curl -H "Authorization: Bearer %s" <your-api-endpoint>`, tokenString),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateInvalidToken generates an invalid token for testing
// GenerateInvalidToken generates an invalid JWT token for testing
func (h *Handler) GenerateInvalidToken(w http.ResponseWriter, r *http.Request) {
	// Parse the request body as a generic map to capture all fields
	var rawRequest map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawRequest); err != nil {
		http.Error(w, `{"error": "Invalid JSON request"}`, http.StatusBadRequest)
		return
	}

	// Extract expiresIn if present, default to "1h"
	expiresIn := "1h"
	if exp, ok := rawRequest["expiresIn"].(string); ok && exp != "" {
		expiresIn = exp
	}

	// Remove expiresIn from claims as it's not a JWT claim
	delete(rawRequest, "expiresIn")

	// Set default claims if the request is empty
	if len(rawRequest) == 0 {
		rawRequest = map[string]interface{}{
			"userId": "invalid-test-user",
			"email":  "invalid-test@example.com",
			"name":   "Invalid Test User",
			"roles":  []string{"user"},
		}
	}

	// Get a valid key to use its kid
	validKey, err := h.keyManager.GetRandomKey()
	if err != nil {
		log.Printf("Error getting random key: %v", err)
		http.Error(w, `{"error": "Failed to get signing key"}`, http.StatusInternalServerError)
		return
	}

	// Generate a temporary invalid key pair
	invalidPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Printf("Error generating invalid key: %v", err)
		http.Error(w, `{"error": "Failed to generate invalid key"}`, http.StatusInternalServerError)
		return
	}

	// Calculate expiration
	exp := time.Now().Add(time.Hour) // Default 1 hour
	if expiresIn != "1h" {
		if seconds, err := strconv.Atoi(expiresIn); err == nil {
			exp = time.Now().Add(time.Duration(seconds) * time.Second)
		}
	}

	// Create claims starting with the dynamic claims from the request
	claims := jwt.MapClaims{}
	for key, value := range rawRequest {
		claims[key] = value
	}

	// Add standard JWT claims (these override any user-provided values for security)
	claims["iat"] = time.Now().Unix()
	claims["exp"] = exp.Unix()
	claims["iss"] = h.config.JWT.Issuer
	claims["aud"] = h.config.JWT.Audience

	// Create token with valid kid but sign with invalid key
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = validKey.Kid

	// Sign token with invalid key
	tokenString, err := token.SignedString(invalidPrivateKey)
	if err != nil {
		log.Printf("Error signing invalid token: %v", err)
		http.Error(w, `{"error": "Failed to sign invalid token"}`, http.StatusInternalServerError)
		return
	}

	response := TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		KeyID:       validKey.Kid,
		RawRequest:  rawRequest, // Include all the dynamic request data
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// QuickInvalidToken generates a quick invalid token for testing
func (h *Handler) QuickInvalidToken(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = "quick-invalid-user"
	}

	// Get a valid key to use its kid
	validKey, err := h.keyManager.GetRandomKey()
	if err != nil {
		log.Printf("Error getting random key: %v", err)
		http.Error(w, `{"error": "Failed to get signing key"}`, http.StatusInternalServerError)
		return
	}

	// Generate a temporary invalid key pair
	invalidPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Printf("Error generating invalid key: %v", err)
		http.Error(w, `{"error": "Failed to generate invalid key"}`, http.StatusInternalServerError)
		return
	}

	// Create claims
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": fmt.Sprintf("%s@example.com", userID),
		"name":  fmt.Sprintf("%s User", userID),
		"roles": []string{"user"},
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iss":   h.config.JWT.Issuer,
		"aud":   h.config.JWT.Audience,
	}

	// Create token with valid kid but sign with invalid key
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = validKey.Kid

	// Sign token with invalid key
	tokenString, err := token.SignedString(invalidPrivateKey)
	if err != nil {
		log.Printf("Error signing invalid token: %v", err)
		http.Error(w, `{"error": "Failed to sign invalid token"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token":        tokenString,
		"key_id":       validKey.Kid,
		"note":         "Invalid token - same kid as valid key but signed with different key",
		"curl_example": fmt.Sprintf(`curl -H "Authorization: Bearer %s" <your-api-endpoint>`, tokenString),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Health returns the health status of the service
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:        "ok",
		Service:       "jwt-dev-service",
		AvailableKeys: h.keyManager.GetAllKeyIDs(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Keys returns information about available keys
func (h *Handler) Keys(w http.ResponseWriter, r *http.Request) {
	keyIDs := h.keyManager.GetAllKeyIDs()
	availableKeys := make([]map[string]interface{}, len(keyIDs))

	for i, kid := range keyIDs {
		availableKeys[i] = map[string]interface{}{
			"kid": kid,
			"alg": "RS256",
			"use": "sig",
		}
	}

	response := KeysResponse{
		TotalKeys:     h.keyManager.GetKeyCount(),
		AvailableKeys: availableKeys,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CORS middleware
func (h *Handler) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}