# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all services
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/api-gateway ./cmd/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/payment-service ./cmd/payment-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/tax-service ./cmd/tax-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/reporting-service ./cmd/reporting-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/notification-service ./cmd/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/auth-service ./cmd/auth-service

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/ ./
COPY --from=builder /app/migrations/ ./migrations/
COPY --from=builder /app/configs/ ./configs/

RUN chown -R app:app /app

USER app

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --spider -q http://localhost:8080/health || exit 1

# Default to api-gateway, can be overridden via CMD
ENTRYPOINT ["./api-gateway"]
