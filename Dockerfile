# Build stage
FROM golang:1.23-bullseye AS builder

# Update ca-certificates
RUN apt-get update && apt-get install -y ca-certificates git && rm -rf /var/lib/apt/lists/*

# Create appuser
RUN adduser --disabled-password --gecos '' appuser

WORKDIR /build

# Copy go mod files and vendor directory
COPY go.mod go.sum ./
COPY vendor/ vendor/

# Copy source code (excluding vendor directory via .dockerignore)
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -mod=vendor \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o jwks-mock-api \
    ./cmd/jwks-mock-api

# Final stage
FROM alpine:latest

# Copy our static executable
COPY --from=builder /build/jwks-mock-api /jwks-mock-api

# Create appuser for runtime
RUN adduser -D -g '' appuser

# Use an unprivileged user
USER appuser

# Expose port
EXPOSE 3000

# Run the binary
ENTRYPOINT ["/jwks-mock-api"]