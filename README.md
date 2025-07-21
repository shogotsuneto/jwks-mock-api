# JWKS Mock API

*This project is bootstrapped by GitHub Copilot, including this documentation.*

A lightweight mock JSON Web Key Set (JWKS) service for backend API development and testing. This service provides JWT token generation, validation, and JWKS endpoints for development environments.

## Features

- **Lightweight**: Single binary (~10MB) with minimal resource usage
- **Fast startup**: Ready in milliseconds
- **Multiple RSA key pairs**: Configurable number of keys for testing key rotation
- **JWT generation**: Create valid and invalid tokens for testing
- **JWT validation**: Validate tokens against the managed keys
- **JWKS endpoint**: Standard `.well-known/jwks.json` endpoint
- **Flexible configuration**: Environment variables and config file support
- **Docker support**: Small container image for easy deployment
- **CORS enabled**: Ready for browser-based testing

## Use Cases

### Backend API Development
Use this service to provide JWT tokens for your development APIs without needing a full authentication system.

### Testing JWT Validation
Generate both valid and invalid tokens to test your JWT validation logic.

### Integration Testing
Include this service in your test environments to provide consistent JWT tokens.

### Local Development
Run locally to avoid dependencies on external authentication services during development.

## Quick Start

### Using Docker (Recommended)

```bash
# Pull and run the container
docker run -p 3000:3000 jwks-mock-api:latest

# Or build locally
docker build -t jwks-mock-api .
docker run -p 3000:3000 jwks-mock-api
```

### Using Go

```bash
# Clone the repository
git clone https://github.com/shogotsuneto/jwks-mock-api.git
cd jwks-mock-api

# Build and run
make build
./jwks-mock-api

# Or run directly
make run
```

## API Endpoints

### JWKS Endpoint
- **GET** `/.well-known/jwks.json` - Returns the JSON Web Key Set

### Token Generation
- **POST** `/generate-token` - Generate a JWT token
- **GET** `/quick-token?userId=user123` - Quick token generation
- **POST** `/generate-invalid-token` - Generate an invalid token for testing
- **GET** `/quick-invalid-token?userId=user123` - Quick invalid token generation

### Token Validation
- **POST** `/validate-token` - Validate a JWT token

### Service Information
- **GET** `/health` - Health check endpoint
- **GET** `/keys` - Available keys information

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Server port |
| `HOST` | `0.0.0.0` | Server host |
| `JWT_ISSUER` | `http://localhost:3000` | JWT issuer claim |
| `JWT_AUDIENCE` | `dev-api` | JWT audience claim |
| `KEY_COUNT` | `2` | Number of RSA key pairs to generate |
| `KEY_IDS` | `key-1,key-2` | Comma-separated key IDs |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

### Configuration File

Create a `config.yaml` file (see `config.yaml.example`):

```yaml
server:
  port: 3000
  host: "0.0.0.0"

jwt:
  issuer: "http://localhost:3000"
  audience: "dev-api"

keys:
  count: 2
  key_ids:
    - "key-1"
    - "key-2"

log_level: "info"
```

Run with config file:
```bash
./jwks-mock-api -config config.yaml
```

## API Usage Examples

### Generate a Token

```bash
curl -X POST http://localhost:3000/generate-token \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user123",
    "email": "user123@example.com",
    "name": "Test User",
    "roles": ["user", "admin"],
    "expiresIn": "1h"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0xIiwiaW5hLWwdM0NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": "1h",
  "key_id": "key-1",
  "user": {
    "id": "user123",
    "email": "user123@example.com",
    "name": "Test User",
    "roles": ["user", "admin"]
  }
}
```

### Quick Token Generation

```bash
curl "http://localhost:3000/quick-token?userId=testuser"
```

### Validate a Token

```bash
curl -X POST http://localhost:3000/validate-token \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0xIiwiaW5hLWwdM0NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### Get JWKS

```bash
curl http://localhost:3000/.well-known/jwks.json
```

### Health Check

```bash
curl http://localhost:3000/health
```

## Development

### Building

```bash
# Standard build
make build

# Optimized build (smaller binary)
make build-optimized

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o jwks-mock-api-linux ./cmd/jwks-mock-api
GOOS=windows GOARCH=amd64 go build -o jwks-mock-api-windows.exe ./cmd/jwks-mock-api
GOOS=darwin GOARCH=amd64 go build -o jwks-mock-api-macos ./cmd/jwks-mock-api
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Vet code
make vet
```

### Docker

```bash
# Build image
make docker

# Run container
make docker-run

# Run with custom environment
make docker-run-env
```

## Project Structure

```
.
├── cmd/
│   └── jwks-mock-api/     # Main application entry point
├── internal/
│   ├── keys/              # Key management
│   └── server/            # HTTP server implementation
├── pkg/
│   ├── config/            # Configuration management
│   └── handlers/          # HTTP handlers
├── config.yaml.example   # Example configuration
├── Dockerfile            # Container image definition
├── Makefile              # Build automation
└── README.md             # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is provided "as is" without warranty of any kind, express or implied. Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files, to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.