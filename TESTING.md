# API Testing Documentation

This document describes the comprehensive test suite for the JWKS Mock API, designed specifically for backend API developers.

## Test Coverage

The test suite provides **80.8%+ coverage** for core components:

- **`pkg/handlers`**: 80.8% - Complete API endpoint testing
- **`internal/keys`**: 80.8% - Key management and JWKS generation
- **`pkg/config`**: 100% - Configuration loading and environment handling
- **`internal/server`**: 34.1% - Server setup and routing

## Test Structure

### 1. API Integration Tests (`pkg/handlers/handlers_test.go`)

Tests all API endpoints with various input scenarios:

#### Health Endpoint (`/health`)
- Validates service status and available keys
- Tests response format and required fields

#### Keys Endpoint (`/keys`) 
- Validates key information structure
- Tests key metadata (algorithm, usage, key IDs)

#### JWKS Endpoint (`/.well-known/jwks.json`)
- Tests RFC 7517 compliance
- Validates JWKS structure and required fields
- Tests caching headers
- Verifies RSA key format (kty, use, kid, alg, n, e)

#### Token Generation (`/generate-token`)
- Tests custom claims handling
- Tests various token expiration times
- Validates JWT structure and signing
- Tests complex nested claims
- Error handling for invalid JSON

#### Invalid Token Generation (`/generate-invalid-token`)
- Tests generation of tokens signed with wrong keys
- Useful for testing token validation failure scenarios

#### Token Introspection (`/introspect`)
- Tests OAuth 2.0 RFC 7662 compliance
- Validates active/inactive token responses
- Tests with valid, invalid, and malformed tokens
- Tests form data parsing
- Custom claims preservation in introspection

#### CORS Support
- Tests preflight OPTIONS requests
- Validates CORS headers for cross-origin requests

### 2. Real-World Scenarios (`pkg/handlers/scenarios_test.go`)

Comprehensive scenarios that backend developers encounter:

#### Complete JWT Workflow
- Generate token → Validate token → Introspect token
- Tests end-to-end JWT lifecycle
- Validates custom user claims (roles, permissions, metadata)

#### API Testing Scenario
- Token generation for API endpoint testing
- Scope-based authorization validation
- Client identification in tokens

#### Microservices Communication
- Service-to-service authentication tokens
- Service identity and version tracking
- Inter-service authorization scopes

#### Key Rotation Simulation
- Multiple key usage testing
- Validates that different keys work correctly
- Simulates production key rotation scenarios

#### JWKS Validation Workflow
- Fetches JWKS endpoint like JWT libraries do
- Validates RFC 7517 compliance
- Tests RSA key field validation

#### Error Handling
- Malformed token introspection
- Invalid token scenarios
- Network error simulation

#### Load Testing Support
- High-volume token generation
- Performance validation
- Batch token processing

### 3. Key Manager Tests (`internal/keys/manager_test.go`)

Tests the core cryptographic functionality:

#### Key Generation
- RSA key pair generation
- Multiple key ID support
- JWK format compliance

#### Key Retrieval
- Random key selection for load balancing
- Key lookup by ID
- Error handling for missing keys

#### JWKS Generation
- RFC 7517 compliant JWKS creation
- Public key extraction
- Key metadata preservation

#### PEM Conversion
- Private key to PEM format
- Public key to PEM format
- Format validation

#### Performance Tests
- Benchmarks for key operations
- Memory usage validation

### 4. Configuration Tests (`pkg/config/config_test.go`)

Comprehensive configuration testing:

#### Default Configuration
- Default values validation
- Configuration structure

#### Environment Variable Override
- PORT, HOST, JWT_ISSUER, JWT_AUDIENCE
- KEY_COUNT and KEY_IDS handling
- Invalid value fallback behavior

#### File-based Configuration
- YAML file loading
- Environment variable precedence
- Error handling for invalid files

#### Edge Cases
- Empty configuration files
- Invalid YAML format
- Missing files
- Mixed configuration sources

### 5. Server Tests (`internal/server/server_test.go`)

Server initialization and routing:

#### Server Creation
- Configuration validation
- Component initialization
- Error handling

#### Route Setup
- All endpoint availability
- HTTP method validation
- Request/response handling

## Running Tests

### Basic Test Execution
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./pkg/handlers/...
go test ./internal/keys/...
go test ./pkg/config/...
go test ./internal/server/...
```

### Verbose Output
```bash
# Run with verbose output
go test -v ./...

# Run specific test function
go test -v ./pkg/handlers -run TestHealthEndpoint

# Run scenario tests only
go test -v ./pkg/handlers -run TestRealWorldScenarios
```

### Coverage Analysis
```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Coverage by function
go tool cover -func=coverage.out
```

## Test Data and Fixtures

### Test Configuration
- **Port**: Random/3000 for parallel testing
- **Host**: localhost/0.0.0.0
- **Issuer**: http://localhost:3000
- **Audience**: test-api
- **Keys**: test-key-1, test-key-2 (2048-bit RSA)

### Sample Claims Structures

#### User Authentication Token
```json
{
  "sub": "user-12345",
  "email": "john.doe@company.com", 
  "roles": ["developer", "admin"],
  "permissions": ["read:projects", "write:projects"],
  "metadata": {
    "last_login": "2024-01-15T10:30:00Z",
    "login_count": 127
  }
}
```

#### Service Token
```json
{
  "sub": "service-payment",
  "client_type": "service",
  "service_id": "payment-service-v1",
  "scopes": ["payments:read", "payments:write"]
}
```

#### API Testing Token
```json
{
  "sub": "api-test-user",
  "scope": "read:api write:api",
  "client": "test-client",
  "env": "testing"
}
```

## Backend Developer Use Cases

### 1. JWT Validation Testing
```bash
# Generate a token for testing
curl -X POST http://localhost:3000/generate-token \
  -H "Content-Type: application/json" \
  -d '{"claims": {"sub": "test-user", "role": "admin"}}'

# Validate the token using introspection
curl -X POST http://localhost:3000/introspect \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "token=<generated_token>"
```

### 2. JWKS Integration Testing
```bash
# Fetch JWKS for JWT library configuration
curl http://localhost:3000/.well-known/jwks.json

# Test token validation against JWKS
# (Use in your JWT library configuration)
```

### 3. Error Scenario Testing
```bash
# Generate invalid token for failure testing
curl -X POST http://localhost:3000/generate-invalid-token \
  -H "Content-Type: application/json" \
  -d '{"claims": {"sub": "test"}}'

# Test introspection with invalid token
curl -X POST http://localhost:3000/introspect \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "token=<invalid_token>"
```

## Continuous Integration

The test suite is designed for CI/CD pipelines:

- **Fast execution**: Most tests complete in under 10 seconds
- **No external dependencies**: Self-contained test server
- **Deterministic**: Tests use fixed seeds and configurations
- **Parallel safe**: No shared state between tests
- **Coverage reporting**: Compatible with codecov, coveralls

## Adding New Tests

### For New Endpoints
1. Add handler tests in `pkg/handlers/handlers_test.go`
2. Add scenario tests in `pkg/handlers/scenarios_test.go`
3. Update server route tests in `internal/server/server_test.go`

### For New Features
1. Add unit tests for the specific component
2. Add integration tests for the full workflow
3. Add real-world scenario tests for developer use cases

### Test Guidelines
- Test both success and failure cases
- Include edge cases and boundary conditions
- Test with realistic data structures
- Validate response formats and HTTP status codes
- Test error handling and recovery
- Include performance considerations for high-load scenarios