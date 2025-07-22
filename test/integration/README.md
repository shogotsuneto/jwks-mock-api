# Docker Integration Tests

This directory contains Docker-based integration tests that run against real API endpoints in containerized environments.

## What's Different from Unit Tests

- **Unit Tests** (`pkg/handlers/*_test.go`): Use in-memory `httptest.Server`, fast execution, isolated testing
- **Integration Tests** (`test/integration/*_test.go`): Use real Docker containers, actual HTTP requests, production-like environment

## Test Structure

### API Tests (`api_test.go`)
- Tests all endpoints with real HTTP requests
- Validates CORS, JWKS, token generation, and introspection
- Tests error handling with actual network calls

### Scenario Tests (`scenarios_test.go`)
- Complete JWT workflow testing
- Microservices communication simulation
- Key rotation testing
- High-volume token generation
- API endpoint authentication testing

## Running Tests

### Option 1: Full Docker Setup (Recommended for CI/CD)
```bash
# Run integration tests in Docker containers
make test-integration

# This will:
# 1. Build Docker images for API server and test runner
# 2. Start API server in container
# 3. Wait for health check to pass
# 4. Run tests in separate container
# 5. Clean up containers
```

### Option 2: External Docker Server (For Development)
```bash
# Run tests against external Docker server
make test-integration-external

# This will:
# 1. Start API server in Docker on port 3002
# 2. Run tests locally against the Docker server
# 3. Clean up server container
```

### Option 3: Manual Setup
```bash
# Terminal 1: Start the server
docker-compose -f docker-compose.test.yml up jwks-api

# Terminal 2: Run tests
JWKS_API_URL=http://localhost:3001 go test -v ./test/integration/...

# Terminal 1: Clean up
docker-compose -f docker-compose.test.yml down
```

## Environment Variables

- `JWKS_API_URL`: API server URL (default: `http://localhost:3001`)
- `TEST_TIMEOUT`: Timeout for HTTP requests (default: `30s`)

## Test Configuration

The integration tests use a separate configuration:
- **Port**: 3001 (to avoid conflicts)
- **Issuer**: `http://jwks-api:3000` (container networking)
- **Audience**: `integration-test-api`
- **Keys**: 3 keys for rotation testing
- **Key IDs**: `integration-key-1`, `integration-key-2`, `integration-key-3`

## CI/CD Integration

### GitHub Actions Example
```yaml
- name: Run Integration Tests
  run: |
    make test-integration
```

### GitLab CI Example
```yaml
integration_tests:
  script:
    - make test-integration
  services:
    - docker:dind
```

## Benefits of Docker Integration Tests

1. **Real Environment**: Tests actual containerized deployment
2. **Network Testing**: Validates container networking and communication
3. **Production Parity**: Matches production Docker environment
4. **State Persistence**: Ready for testing future persistence features
5. **Load Testing**: Tests under realistic conditions
6. **Debugging**: Can inspect running containers and logs

## Debugging Failed Tests

```bash
# View logs from the API server container
docker-compose -f docker-compose.test.yml logs jwks-api

# Run tests with verbose output
JWKS_API_URL=http://localhost:3001 go test -v ./test/integration/... -run TestSpecificTest

# Keep containers running for debugging
docker-compose -f docker-compose.test.yml up -d jwks-api
# Run individual tests manually
# Clean up when done
docker-compose -f docker-compose.test.yml down
```

## Test Coverage

The integration tests cover:
- ✅ All API endpoints with real HTTP requests
- ✅ Complete JWT workflows (generate → validate → introspect)
- ✅ Microservices authentication scenarios  
- ✅ Key rotation simulation
- ✅ High-volume token generation
- ✅ CORS validation
- ✅ Error handling and edge cases
- ✅ Performance under load

This complements the unit tests by validating that the service works correctly when deployed as a container.