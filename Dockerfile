# Build stage
FROM golang:1.23 AS builder

# Install ca-certificates for TLS
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Create appuser
RUN useradd -m -s /bin/bash appuser

WORKDIR /build

# Copy go mod files and vendor directory
COPY go.mod go.sum ./
COPY vendor/ vendor/

# Copy source code
COPY . .

# Build with vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o jwks-mock-api \
    ./cmd/jwks-mock-api

# Final stage
FROM debian:bookworm-slim

# Install ca-certificates for runtime
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy our static executable
COPY --from=builder /build/jwks-mock-api /jwks-mock-api

# Create appuser for runtime
RUN useradd -m -s /bin/bash appuser

# Use an unprivileged user
USER appuser

# Expose port
EXPOSE 3000

# Run the binary
ENTRYPOINT ["/jwks-mock-api"]