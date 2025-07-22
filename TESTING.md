# API Testing Documentation

This document describes the comprehensive test suite for the JWKS Mock API, designed specifically for backend API developers.

## Test Coverage

### Unit Test Coverage

The unit test suite provides **80.8%+ coverage** for core components:

- **`pkg/handlers`**: 80.8% - Complete API endpoint testing
- **`internal/keys`**: 80.8% - Key management and JWKS generation
- **`pkg/config`**: 100% - Configuration loading and environment handling
- **`internal/server`**: 34.1% - Server setup and routing

### Docker Integration Tests

The Docker integration test suite provides **real-world scenario testing**:

- **Real API Endpoints**: Tests against actual containerized service
- **Production Environment**: Validates Docker deployment scenarios
- **Network Communication**: Tests container-to-container communication
- **State Persistence**: Ready for testing future persistence features
- **Load Testing**: Performance validation under realistic conditions

#### Integration Test Structure (`test/integration/`)

##### API Tests (`api_test.go`)
- Real HTTP requests to containerized API
- JWKS endpoint validation with actual network calls
- Token generation and introspection via HTTP
- CORS validation with real browsers
- Error handling with network timeouts

##### Real-World Scenarios (`scenarios_test.go`)
- **Complete JWT Workflow**: Generate → Fetch JWKS → Validate → Introspect
- **Microservices Communication**: Service-to-service token testing
- **Key Rotation Simulation**: Multiple key usage in production-like environment
- **High-Volume Testing**: Load testing with 50+ concurrent requests
- **API Authentication**: Scope-based authorization testing

## Unit Test Details

## Test Structure

The testing strategy includes two complementary approaches:

### 1. Unit/Integration Tests (Fast Development Cycle)
- **Location**: `pkg/handlers/*_test.go`, `internal/*/test.go`
- **Method**: In-memory `httptest.Server`
- **Speed**: Fast (< 10 seconds)
- **Purpose**: Development, debugging, code coverage

### 2. Docker Integration Tests (Production-Like Testing)
- **Location**: `test/integration/*_test.go`
- **Method**: Real HTTP requests to containerized API
- **Speed**: Moderate (30-60 seconds)
- **Purpose**: CI/CD, deployment validation, real-world scenarios

### Unit Test Coverage

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

### Unit Tests (Fast Development)
```bash
# Run all unit tests
make test-unit

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./pkg/handlers/...
go test ./internal/keys/...
go test ./pkg/config/...
go test ./internal/server/...
```

### Docker Integration Tests (Production-Like)
```bash
# Run Docker-based integration tests
make test-integration

# Run integration tests against external Docker server (for development)
make test-integration-external

# Run all tests (unit + integration)
make test-all
```

### Basic Test Execution (Legacy)
```bash
# Run all tests (includes both unit and integration if Docker is available)
make test

# Verbose output
go test -v ./...

### Specific Test Execution
```bash
# Run with verbose output (unit tests)
go test -v ./pkg/... ./internal/...

# Run specific test function (unit tests)
go test -v ./pkg/handlers -run TestHealthEndpoint

# Run scenario tests only (unit tests)
go test -v ./pkg/handlers -run TestRealWorldScenarios

# Run specific integration test
JWKS_API_URL=http://localhost:3001 go test -v ./test/integration -run TestIntegrationCompleteJWTWorkflow

# Run integration scenarios only
JWKS_API_URL=http://localhost:3001 go test -v ./test/integration -run TestIntegrationMicroservices
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

### 1. Fast Development Cycle (Unit Tests)
```bash
# Quick feedback during development
make test-unit

# Test specific handlers
go test -v ./pkg/handlers -run TestHealthEndpoint

# Test with coverage
make test-coverage
```

### 2. Production-Ready Testing (Docker Integration Tests)
```bash
# Full integration testing with Docker
make test-integration

# Development workflow with external Docker server
make test-integration-external

# Run specific integration scenarios
JWKS_API_URL=http://localhost:3001 go test -v ./test/integration -run TestIntegrationCompleteJWTWorkflow
```

### 3. JWT Validation Testing (Both Approaches Available)
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

### 3. JWT Validation Testing (Both Approaches Available)
```bash
# Generate a token for testing (works with both unit and Docker tests)
curl -X POST http://localhost:3000/generate-token \
  -H "Content-Type: application/json" \
  -d '{"claims": {"sub": "test-user", "role": "admin"}}'

# Validate the token using introspection
curl -X POST http://localhost:3000/introspect \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "token=<generated_token>"
```

### 4. JWKS Integration Testing
```bash
# Fetch JWKS for JWT library configuration
curl http://localhost:3000/.well-known/jwks.json

# Test token validation against JWKS
# (Use in your JWT library configuration)
```

### 5. Error Scenario Testing
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

### When to Use Which Test Type

- **Unit Tests**: Fast feedback during development, testing specific components, code coverage
- **Docker Integration Tests**: Pre-deployment validation, testing real-world scenarios, CI/CD pipelines
- **Both**: JWT workflows, API endpoint testing, error scenario validation

## Test Summary

This testing suite provides both **fast development feedback** and **production-ready validation**:

| Test Type | Purpose | Execution Time | Use Case |
|-----------|---------|---------------|----------|
| **Unit Tests** | Fast development feedback | < 10 seconds | Development, debugging, code coverage |
| **Docker Integration** | Production validation | 30-60 seconds | CI/CD, deployment testing, real-world scenarios |

Both test types cover the same scenarios but with different approaches:
- Unit tests use in-memory servers for speed
- Docker tests use real containers for production parity

## Continuous Integration

### Unit Tests
The unit test suite is designed for fast CI/CD pipelines:

- **Fast execution**: Most tests complete in under 10 seconds
- **No external dependencies**: Self-contained test server
- **Deterministic**: Tests use fixed seeds and configurations
- **Parallel safe**: No shared state between tests
- **Coverage reporting**: Compatible with codecov, coveralls

### Docker Integration Tests
The integration test suite is designed for comprehensive CI/CD validation:

- **Production parity**: Tests actual containerized deployment
- **Network validation**: Tests container networking and communication
- **State persistence**: Ready for testing future persistence features like key storage
- **Load testing**: Validates performance under realistic conditions
- **Real HTTP**: Tests actual HTTP clients and network behavior

#### CI/CD Pipeline Example
```yaml
# GitHub Actions
name: Tests
on: [push, pull_request]
jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.21
      - run: make test-unit
      
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: make test-integration
```

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