# Build stage
FROM golang:1.23-alpine AS builder

# Install ca-certificates and git
RUN apk --no-cache add ca-certificates git

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# DO NOT copy vendor directory - always download dependencies fresh
# This ensures cross-platform compatibility and avoids stale dependencies
# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
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