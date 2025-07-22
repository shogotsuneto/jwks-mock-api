# Build stage
FROM golang:1.23-alpine AS builder

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /build

# Copy go mod files and source
COPY go.mod go.sum ./
COPY . .

# Create vendor directory
RUN go mod vendor

# Build with vendor
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