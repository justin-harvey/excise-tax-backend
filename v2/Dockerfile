# Production Dockerfile - API Server
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build optimized binaries with stripped symbols
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags='-s -w -extldflags "-static"' -o /api-server ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags='-s -w -extldflags "-static"' -o /xrpl-cli ./cmd/xrpl-cli/main.go

# Production runtime - using scratch for minimal image size
FROM scratch

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binaries
COPY --from=builder /api-server /api-server
COPY --from=builder /xrpl-cli /xrpl-cli

# Copy OpenAPI spec to /api-spec/ to avoid conflict with binary
COPY --from=builder /app/api/openapi.yaml /api-spec/openapi.yaml

# Expose API port
EXPOSE 8080

# Run API server
CMD ["/api-server"]

