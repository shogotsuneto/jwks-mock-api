# JWKS Mock API

*This project is bootstrapped by GitHub Copilot, including this documentation.*

A lightweight mock JSON Web Key Set (JWKS) service for backend API development and testing. Single binary (~10MB) with millisecond startup time, providing JWT generation, validation, and JWKS endpoints.

## Features

- **🚀 Lightweight & Fast**: ~10MB binary, instant startup, minimal resource usage
- **🔐 Dynamic JWT Claims**: Generate tokens with any custom JSON structure as claims
- **🔄 Multiple Keys**: Configurable RSA key pairs for testing key rotation
- **⚙️ Flexible Config**: Environment variables and YAML config file support  
- **🐳 Docker Ready**: Small container image for easy deployment
- **🧪 Testing Support**: Generate both valid and invalid tokens for comprehensive testing

## Quick Start

```bash
# Docker (Recommended)
docker run -p 3000:3000 jwks-mock-api:latest

# Or build from source
git clone https://github.com/shogotsuneto/jwks-mock-api.git
cd jwks-mock-api && make build && ./jwks-mock-api
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/.well-known/jwks.json` | Standard JWKS endpoint |
| POST | `/generate-token` | Generate JWT with **dynamic claims** |
| POST | `/generate-invalid-token` | Invalid token for testing |
| POST | `/introspect` | OAuth 2.0 Token Introspection (RFC 7662) |
| GET | `/health` | Health check |
| GET | `/keys` | Available keys info |
| POST | `/keys` | Add a new key |
| DELETE | `/keys/{kid}` | Remove a key by ID |

## Configuration

**Environment Variables:**
- `PORT=3000` - Server port
- `JWT_ISSUER=http://localhost:3000` - JWT issuer
- `JWT_AUDIENCE=dev-api` - JWT audience  
- `KEY_COUNT=2` - Number of RSA key pairs
- `KEY_IDS=key-1,key-2` - Comma-separated key IDs

**Config File:** Create `config.yaml` (see `config.yaml.example`):
```yaml
server:
  port: 3000
jwt:
  issuer: "http://localhost:3000"
  audience: "dev-api"
keys:
  count: 2
  key_ids: ["key-1", "key-2"]
```

Run with: `./jwks-mock-api -config config.yaml`

## Dynamic Claims Support

**The `/generate-token` endpoint accepts a structured request with claims nested under a `claims` key.** This separates configuration options (like `expiresIn`) from actual JWT claims, enabling flexible token generation for various testing scenarios.

### Examples

**Basic Token:**
```bash
curl -X POST http://localhost:3000/generate-token \
  -H "Content-Type: application/json" \
  -d '{"claims": {"sub": "user123", "role": "admin"}}'
```

**Complex Claims:**
```bash
curl -X POST http://localhost:3000/generate-token \
  -H "Content-Type: application/json" \
  -d '{
    "claims": {
      "sub": "user123",
      "profile": {
        "name": "John Doe",
        "email": "john@example.com",
        "department": "Engineering"
      },
      "permissions": ["read", "write", "admin"],
      "metadata": {
        "loginCount": 42,
        "lastLogin": "2024-01-15T10:30:00Z"
      }
    },
    "expiresIn": 3600
  }'
```

**Response Format:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0xIiwidHlwIjoiSldUIn0...",
  "expires_in": 3600,
  "key_id": "key-1"
}
```

> **Note:** Standard JWT fields (`iat`, `exp`, `iss`, `aud`) are automatically added. The `expiresIn` field (in seconds) controls token expiration and is not included as a claim.

### Other Examples

**Introspect Token (OAuth 2.0 RFC 7662):**
```bash
curl -X POST http://localhost:3000/introspect \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "token=eyJhbGciOiJSUzI1NiIs..."
```

**Get JWKS:** `curl http://localhost:3000/.well-known/jwks.json`

**Add Key:** 
```bash
curl -X POST http://localhost:3000/keys \
  -H "Content-Type: application/json" \
  -d '{"kid": "new-key-id"}'
```

**Remove Key:**
```bash
curl -X DELETE http://localhost:3000/keys/key-to-remove
```

## Development

```bash
# Build
make build                    # Standard build
make build-optimized         # Smaller binary

# Test  
make test-integration       # Docker-based integration tests
make test-integration-external  # Integration tests with external server
make test-coverage          # With coverage

# Docker
make docker                 # Build image
make docker-run            # Run container

# Code quality
make fmt && make vet       # Format and vet
```

### Integration Testing

The project includes comprehensive Docker-based integration tests that validate real API endpoints in containerized environments. These tests cover:

- **Endpoint Testing**: All API endpoints with real HTTP requests
- **JWT Workflows**: Complete token generation, validation, and introspection
- **Microservices Scenarios**: Service-to-service authentication patterns
- **Key Rotation**: Multiple key usage simulation

See `test/integration/README.md` for detailed testing documentation.

## Project Structure

```
├── cmd/jwks-mock-api/     # Main application
├── internal/              # Private packages
│   ├── keys/              # Key management  
│   └── server/            # HTTP server
├── pkg/                   # Public packages
│   ├── config/            # Configuration
│   └── handlers/          # HTTP handlers
├── config.yaml.example   # Example config
├── Dockerfile & Makefile # Build automation
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.