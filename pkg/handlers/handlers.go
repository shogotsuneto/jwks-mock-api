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
	ExpiresIn   string                 `json:"expires_in"`
	KeyID       string                 `json:"key_id"`
	RawRequest  map[string]interface{} `json:"raw_request"`
}

// IntrospectionResponse represents an OAuth 2.0 token introspection response (RFC 7662)
type IntrospectionResponse struct {
	Active    bool                   `json:"active"`
	TokenType string                 `json:"token_type,omitempty"`
	Scope     string                 `json:"scope,omitempty"`
	ClientID  string                 `json:"client_id,omitempty"`
	Username  string                 `json:"username,omitempty"`
	Exp       int64                  `json:"exp,omitempty"`
	Iat       int64                  `json:"iat,omitempty"`
	Nbf       int64                  `json:"nbf,omitempty"`
	Sub       string                 `json:"sub,omitempty"`
	Aud       string                 `json:"aud,omitempty"`
	Iss       string                 `json:"iss,omitempty"`
	Jti       string                 `json:"jti,omitempty"`
	// Additional claims from the original token
	Claims map[string]interface{} `json:"-"` // Use custom marshaling to flatten
}

// MarshalJSON implements custom JSON marshaling to flatten claims into the response
func (r IntrospectionResponse) MarshalJSON() ([]byte, error) {
	// Create a map with the standard fields
	result := map[string]interface{}{
		"active": r.Active,
	}
	
	// Add optional standard fields
	if r.TokenType != "" {
		result["token_type"] = r.TokenType
	}
	if r.Scope != "" {
		result["scope"] = r.Scope
	}
	if r.ClientID != "" {
		result["client_id"] = r.ClientID
	}
	if r.Username != "" {
		result["username"] = r.Username
	}
	if r.Exp != 0 {
		result["exp"] = r.Exp
	}
	if r.Iat != 0 {
		result["iat"] = r.Iat
	}
	if r.Nbf != 0 {
		result["nbf"] = r.Nbf
	}
	if r.Sub != "" {
		result["sub"] = r.Sub
	}
	if r.Aud != "" {
		result["aud"] = r.Aud
	}
	if r.Iss != "" {
		result["iss"] = r.Iss
	}
	if r.Jti != "" {
		result["jti"] = r.Jti
	}
	
	// Add additional claims, avoiding overwriting standard fields
	standardFields := map[string]bool{
		"active": true, "token_type": true, "scope": true, "client_id": true,
		"username": true, "exp": true, "iat": true, "nbf": true,
		"sub": true, "aud": true, "iss": true, "jti": true,
	}
	
	for key, value := range r.Claims {
		if !standardFields[key] {
			result[key] = value
		}
	}
	
	return json.Marshal(result)
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

// TokenRequest represents the structure expected for token generation
type TokenRequest struct {
	Claims    map[string]interface{} `json:"claims"`
	ExpiresIn *int                   `json:"expiresIn,omitempty"` // seconds
}

// GenerateToken generates a new JWT token with dynamic claims
func (h *Handler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	// Parse the request body with the new structure
	var request TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, `{"error": "Invalid JSON request"}`, http.StatusBadRequest)
		return
	}

	// Extract expiresIn if present, default to 3600 seconds (1 hour)
	expiresInSeconds := 3600
	if request.ExpiresIn != nil {
		expiresInSeconds = *request.ExpiresIn
	}

	// Set default claims if none provided
	claims := request.Claims
	if len(claims) == 0 {
		claims = map[string]interface{}{
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

	// Calculate expiration based on seconds
	exp := time.Now().Add(time.Duration(expiresInSeconds) * time.Second)

	// Create JWT claims starting with the dynamic claims from the request
	jwtClaims := jwt.MapClaims{}
	for key, value := range claims {
		jwtClaims[key] = value
	}

	// Add standard JWT claims (these override any user-provided values for security)
	jwtClaims["iat"] = time.Now().Unix()
	jwtClaims["exp"] = exp.Unix()
	jwtClaims["iss"] = h.config.JWT.Issuer
	jwtClaims["aud"] = h.config.JWT.Audience

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
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
		ExpiresIn:   strconv.Itoa(expiresInSeconds),
		KeyID:       keyPair.Kid,
		RawRequest:  claims, // Include all the dynamic request claims
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Introspect implements OAuth 2.0 Token Introspection (RFC 7662)
func (h *Handler) Introspect(w http.ResponseWriter, r *http.Request) {
	// Parse form data (RFC 7662 requires application/x-www-form-urlencoded)
	if err := r.ParseForm(); err != nil {
		response := IntrospectionResponse{Active: false}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	token := r.FormValue("token")
	if token == "" {
		response := IntrospectionResponse{Active: false}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // RFC 7662: return 200 even for missing token
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse token to get the kid
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
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

	response := IntrospectionResponse{}

	if err != nil || !parsedToken.Valid {
		// Token is not active (invalid, expired, etc.)
		response.Active = false
	} else if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		// Validate issuer and audience
		if claims["iss"] != h.config.JWT.Issuer || claims["aud"] != h.config.JWT.Audience {
			response.Active = false
		} else {
			// Token is active - populate response with claims
			response.Active = true
			response.TokenType = "Bearer"
			
			// Map standard JWT claims to introspection response
			if exp, ok := claims["exp"].(float64); ok {
				response.Exp = int64(exp)
			}
			if iat, ok := claims["iat"].(float64); ok {
				response.Iat = int64(iat)
			}
			if nbf, ok := claims["nbf"].(float64); ok {
				response.Nbf = int64(nbf)
			}
			if sub, ok := claims["sub"].(string); ok {
				response.Sub = sub
				response.Username = sub // Use sub as username
			}
			if aud, ok := claims["aud"].(string); ok {
				response.Aud = aud
			}
			if iss, ok := claims["iss"].(string); ok {
				response.Iss = iss
			}
			if jti, ok := claims["jti"].(string); ok {
				response.Jti = jti
			}
			
			// Add all other claims
			response.Claims = make(map[string]interface{})
			for key, value := range claims {
				response.Claims[key] = value
			}
		}
	} else {
		response.Active = false
	}

	// Always return 200 OK per RFC 7662
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}



// GenerateInvalidToken generates an invalid JWT token for testing
func (h *Handler) GenerateInvalidToken(w http.ResponseWriter, r *http.Request) {
	// Parse the request body with the new structure
	var request TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, `{"error": "Invalid JSON request"}`, http.StatusBadRequest)
		return
	}

	// Extract expiresIn if present, default to 3600 seconds (1 hour)
	expiresInSeconds := 3600
	if request.ExpiresIn != nil {
		expiresInSeconds = *request.ExpiresIn
	}

	// Set default claims if none provided
	claims := request.Claims
	if len(claims) == 0 {
		claims = map[string]interface{}{
			"sub":   "invalid-test-user",
			"email": "invalid-test@example.com",
			"name":  "Invalid Test User",
			"roles": []string{"user"},
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

	// Calculate expiration based on seconds
	exp := time.Now().Add(time.Duration(expiresInSeconds) * time.Second)

	// Create JWT claims starting with the dynamic claims from the request
	jwtClaims := jwt.MapClaims{}
	for key, value := range claims {
		jwtClaims[key] = value
	}

	// Add standard JWT claims (these override any user-provided values for security)
	jwtClaims["iat"] = time.Now().Unix()
	jwtClaims["exp"] = exp.Unix()
	jwtClaims["iss"] = h.config.JWT.Issuer
	jwtClaims["aud"] = h.config.JWT.Audience

	// Create token with valid kid but sign with invalid key
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
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
		ExpiresIn:   strconv.Itoa(expiresInSeconds),
		KeyID:       validKey.Kid,
		RawRequest:  claims, // Include all the dynamic request claims
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