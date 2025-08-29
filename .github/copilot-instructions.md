# JWKS Mock API Development Guide

**ALWAYS follow these instructions first and only search for additional context if the information here is incomplete or incorrect.**

This is a lightweight Go-based JSON Web Key Set (JWKS) mock service for JWT testing and development. Single binary (~10MB) with instant startup providing JWT generation, validation, and JWKS endpoints.

## Quick Setup and Build

**Prerequisites:** Go 1.23+ (use `go version` to check)

**Bootstrap the project:**
```bash
# Clone and setup (if starting fresh)
git clone https://github.com/shogotsuneto/jwks-mock-api.git
cd jwks-mock-api

# Download dependencies - takes ~1 second
make deps

# Build the application - takes ~11 seconds, NEVER CANCEL, use 60+ second timeout
make build
```

## Essential Commands

### Building
```bash
make build              # Standard build (~11 seconds)
make build-optimized    # Smaller binary (~11 seconds)
make clean              # Remove build artifacts
```
**NEVER CANCEL builds** - Set timeout to 60+ seconds. Build takes approximately 11 seconds to complete.

### Testing
```bash
# Primary test suite - Docker-based integration tests
make test-integration   # Takes ~38 seconds, NEVER CANCEL, use 120+ second timeout

# Minimal unit tests (limited coverage)
make test-unit          # Takes ~3 seconds

# External integration testing (for development)
make test-integration-external  # Takes ~45 seconds, NEVER CANCEL
```
**CRITICAL:** `make test-integration` is the primary test suite with comprehensive coverage. **NEVER CANCEL** - takes approximately 38 seconds. Always use 120+ second timeout.

### Running the Application
```bash
# Run directly (instant startup)
./jwks-mock-api

# Or run with Make
make run

# Run with config file
make run-config    # Uses config.yaml.example
```

### Docker Operations
```bash
# Build Docker image - takes ~10 seconds, NEVER CANCEL, use 60+ second timeout  
make docker

# Run Docker container
make docker-run         # Basic run on port 3000
make docker-run-env     # With custom environment variables
```

### Code Quality
```bash
# Format code (always run before committing)
make fmt                # Takes <1 second

# Vet code (always run before committing) 
make vet                # Takes ~2 seconds

# Lint code (requires golangci-lint installation)
make lint               # May need golangci-lint setup
```

## Manual Validation Requirements

**ALWAYS manually validate changes using these scenarios:**

### 1. Basic Health Check
```bash
# Start the application
./jwks-mock-api

# Test health endpoint (should return JSON with status "ok")
curl -s http://localhost:3000/health
```
Expected response: `{"status":"ok","service":"jwt-dev-service","available_keys":["key-1","key-2"]}`

### 2. JWKS Endpoint Test
```bash
# Get JWKS (should return JSON with "keys" array)
curl -s http://localhost:3000/.well-known/jwks.json
```
Expected: JSON response with `"keys"` array containing RSA public keys.

### 3. Token Generation Test
```bash
# Generate a JWT token
curl -s -X POST http://localhost:3000/generate-token \
  -H "Content-Type: application/json" \
  -d '{"claims": {"sub": "test-user", "role": "admin"}}'
```
Expected: JSON response with `"token"`, `"expires_in"`, and `"key_id"` fields.

### 4. Token Introspection Test
```bash
# First generate a token, then introspect it
TOKEN=$(curl -s -X POST http://localhost:3000/generate-token \
  -H "Content-Type: application/json" \
  -d '{"claims": {"sub": "test"}}' | jq -r '.token')

curl -s -X POST http://localhost:3000/introspect \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "token=$TOKEN"
```
Expected: JSON response with `"active": true` and token claims.

### 5. Key Management Test
```bash
# Add a new key
curl -s -X POST http://localhost:3000/keys \
  -H "Content-Type: application/json" \
  -d '{"kid": "test-key"}'

# Verify key was added
curl -s http://localhost:3000/keys | jq '.available_keys | map(.kid) | contains(["test-key"])'

# Remove the key
curl -s -X DELETE http://localhost:3000/keys/test-key
```

**VALIDATION REQUIREMENT:** After making ANY changes, run through scenarios 1-3 at minimum. For key management changes, run scenario 5.

## golangci-lint Setup

golangci-lint is **not installed by default**. Install if needed:
```bash
# Install golangci-lint (required for `make lint`)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Add to PATH for current session
export PATH=$PATH:$(go env GOPATH)/bin

# Now you can run
make lint
```

**Note:** `make lint` may show some import-related warnings but these don't affect functionality.

## CI/CD Validation

**Always run these before committing to ensure CI passes:**
```bash
# Code formatting (required for CI)
make fmt

# Code vetting (required for CI) 
make vet

# Integration tests (primary CI validation) - NEVER CANCEL, 120+ second timeout
make test-integration
```

The GitHub Actions CI (`.github/workflows/pr-test.yml`) runs these same commands.

## Project Structure and Navigation

### Key Directories
```
├── cmd/jwks-mock-api/        # Main application entry point (main.go)
├── internal/                 # Private packages
│   ├── keys/                 # Key management logic
│   └── server/               # HTTP server setup
├── pkg/                      # Public packages
│   ├── config/               # Configuration handling (config.go)
│   ├── handlers/             # HTTP handlers (handlers.go) 
│   └── logger/               # Logging utilities
├── test/integration/         # Docker-based integration tests
│   ├── endpoints/            # API endpoint tests
│   ├── scenarios/            # Complete workflow tests
│   └── common/               # Test utilities
├── .github/workflows/        # CI/CD pipelines
├── config.yaml.example      # Configuration template
├── Makefile                 # Build automation
└── docker-compose.test.yml  # Integration test setup
```

### Important Files to Know
- **`pkg/handlers/handlers.go`** - All HTTP endpoint implementations
- **`pkg/config/config.go`** - Configuration and environment variable handling
- **`cmd/jwks-mock-api/main.go`** - Application entry point
- **`Makefile`** - All build commands and workflows
- **`test/integration/`** - Comprehensive test coverage

### When You Modify Code
- **Handler changes**: Check `pkg/handlers/handlers.go`, always test with scenarios 1-3
- **Configuration changes**: Check `pkg/config/config.go`, test with different env vars
- **Key management**: Check `internal/keys/`, test with scenario 5
- **Docker changes**: Test with `make docker && make docker-run`

## Configuration

### Environment Variables (runtime config)
```bash
PORT=3000                               # Server port
HOST=0.0.0.0                           # Server host  
JWT_ISSUER=http://localhost:3000        # JWT issuer claim
JWT_AUDIENCE=dev-api                    # JWT audience claim
KEY_COUNT=2                             # Number of RSA key pairs
KEY_IDS=key-1,key-2                     # Comma-separated key IDs
LOG_LEVEL=info                          # Log level (debug,info,warn,error)
```

### Config File (optional)
Copy `config.yaml.example` to `config.yaml` and run with:
```bash
./jwks-mock-api -config config.yaml
```

## Docker Usage

### Published Images  
```bash
# Run latest stable
docker run -p 3000:3000 ghcr.io/shogotsuneto/jwks-mock-api:latest

# Run development build
docker run -p 3000:3000 ghcr.io/shogotsuneto/jwks-mock-api:develop-latest
```

### Local Development
```bash
# Build and run local image
make docker && make docker-run

# Run with custom environment
make docker-run-env
```

## Build Times and Timeout Guidance

| Command | Expected Time | Timeout Setting | Never Cancel |
|---------|---------------|-----------------|--------------|
| `make deps` | ~1 second | 30 seconds | No |
| `make build` | ~11 seconds | 60+ seconds | **YES** |
| `make test-unit` | ~3 seconds | 30 seconds | No |
| `make test-integration` | ~38 seconds | 120+ seconds | **YES** |
| `make docker` | ~10 seconds | 60+ seconds | **YES** |
| `make fmt` | <1 second | 10 seconds | No |
| `make vet` | ~2 seconds | 30 seconds | No |

**CRITICAL:** Always set timeouts to 2x the expected time. Some operations like `make test-integration` are essential for validation and must complete.

## Troubleshooting

### "golangci-lint not found"
```bash
# Install golangci-lint as shown in setup section above
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
export PATH=$PATH:$(go env GOPATH)/bin
```

### "Docker daemon not running" (for integration tests)
Integration tests require Docker. Ensure Docker is running:
```bash
docker ps  # Should show running containers or empty list, not error
```

### "Port 3000 already in use"
Kill any existing processes:
```bash
pkill -f jwks-mock-api
# Or use different port
PORT=3001 ./jwks-mock-api
```

### Integration tests failing
1. Ensure Docker is running
2. Clean Docker state: `docker system prune -f`
3. Re-run: `make test-integration`

### Build failures
1. Check Go version: `go version` (need 1.23+)
2. Clean and rebuild: `make clean && make deps && make build`

## Common Development Workflows

### Adding New API Endpoint
1. Modify `pkg/handlers/handlers.go` to add handler function
2. Update routing in `pkg/handlers/handlers.go` 
3. Run `make fmt && make vet`
4. Test manually with curl scenarios
5. Run `make test-integration` to ensure existing functionality works
6. Add integration test in `test/integration/endpoints/` if needed

### Changing Configuration
1. Modify `pkg/config/config.go` for new config options
2. Update `config.yaml.example` with new settings
3. Test with various environment variables
4. Run full validation scenarios
5. Update documentation if needed

### Performance/Security Changes
1. Make changes
2. Run `make build-optimized` for optimized binary
3. Test with `make test-integration` - NEVER CANCEL
4. Run manual scenarios 1-5 for comprehensive validation
5. Test Docker build: `make docker`

## Common Command Outputs (for Reference)

To save time, here are outputs from frequently used commands:

### Repository Root Structure
```
ls -la
.dockerignore
.git/
.github/
.gitignore
Dockerfile
Dockerfile.integration-tests  
LICENSE
Makefile
README.md
cmd/
config.yaml.example
docker-compose.test.yml
docker-compose.yml
docs/
go.mod
go.sum
internal/
pkg/
test/
```

### Health Check Response
```json
{"status":"ok","service":"jwt-dev-service","available_keys":["key-1","key-2"]}
```

### JWKS Endpoint Response Structure
```json
{"keys":[{"alg":"RS256","e":"AQAB","kid":"key-1","kty":"RSA","n":"...","use":"sig"},...]}
```

### Token Generation Response
```json
{"token":"eyJhbGciOiJSUzI1NiIs...","expires_in":3600,"key_id":"key-1","raw_request":{"sub":"test-user"}}
```

### Keys Endpoint Response
```json
{"total_keys":2,"available_keys":[{"alg":"RS256","kid":"key-1","use":"sig"},{"alg":"RS256","kid":"key-2","use":"sig"}]}
```

### Package Structure
```
pkg/
├── config/      # Configuration handling (Load, environment variables)
├── handlers/    # HTTP handlers (JWKS, token generation, introspection)  
└── logger/      # Logging utilities
```

Always remember: **Integration tests are the primary validation mechanism.** The unit test coverage is minimal - rely on the comprehensive Docker-based integration test suite for validation.