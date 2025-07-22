# Build stage
FROM golang:1.21-alpine AS builder

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o jwks-mock-api \
    ./cmd/jwks-mock-api

# Final stage
FROM scratch

# Copy our static executable
COPY --from=builder /build/jwks-mock-api /jwks-mock-api

# Copy appuser from builder
COPY --from=builder /etc/passwd /etc/passwd

# Use an unprivileged user
USER appuser

# Expose port
EXPOSE 3000

# Run the binary
ENTRYPOINT ["/jwks-mock-api"]