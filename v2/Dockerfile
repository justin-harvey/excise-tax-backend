FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git make
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/xrpl-cli ./cmd/xrpl-cli

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app
WORKDIR /app
COPY --from=builder /app/bin/xrpl-cli ./
RUN chown -R app:app /app
USER app
ENTRYPOINT ["./xrpl-cli"]
