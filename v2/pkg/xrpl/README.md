# XRPL Client Library

A clean, idiomatic Go client library for interacting with the XRP Ledger (XRPL) WebSocket API.

This package offers a clean, idiomatic Go interface to the XRP Ledger WebSocket API. It follows Go best practices including standard library first approach, functional options pattern for configuration, context-aware operations, explicit error handling with sentinel errors, and thread-safe concurrent access.

## Features

- **Standard Library First**: Minimal external dependencies
- **Functional Options**: Flexible configuration pattern
- **Context-Aware**: All blocking operations accept `context.Context`
- **Thread-Safe**: Safe for concurrent use by multiple goroutines
- **Explicit Errors**: Sentinel errors for common error conditions
- **Real-time Subscriptions**: Subscribe to account transaction streams
- **Well Documented**: Comprehensive godoc and examples
- **Tested**: Unit tests, integration tests, and benchmarks

## Installation

```bash
go get github.com/maxfelker/excise-tax-backend/v2/pkg/xrpl
```

## Basic Usage

Create a client and connect to XRPL:

```go
client := xrpl.New("wss://s.altnet.rippletest.net:51233")
ctx := context.Background()

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## Configuration

Use functional options to configure the client:

```go
import (
    "log/slog"
    "time"
)

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

client := xrpl.New("wss://s.altnet.rippletest.net:51233",
    xrpl.WithTimeout(15*time.Second),           // Request timeout
    xrpl.WithHandshakeTimeout(10*time.Second),  // Connection timeout
    xrpl.WithLogger(logger),                    // Structured logging
    xrpl.WithMaxRetries(5),                     // Retry attempts
)
```

## Usage Examples

### Get Account Information

```go
info, err := client.GetAccountInfo(ctx, address)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Account: %s\n", info.Account)
fmt.Printf("Balance: %s drops\n", info.Balance)
fmt.Printf("Sequence: %d\n", info.Sequence)
```

### Get Transaction History

```go
// Get last 10 transactions
txs, err := client.GetAccountTransactions(ctx, address, 10)
if err != nil {
    log.Fatal(err)
}

for _, tx := range txs {
    fmt.Printf("Hash: %s\n", tx.Hash)
    fmt.Printf("Type: %s\n", tx.TransactionType)
    fmt.Printf("Validated: %v\n", tx.Validated)
}
```

### Get and Verify Transaction

```go
// Get transaction by hash
tx, err := client.GetTransaction(ctx, hash)
if err != nil {
    log.Fatal(err)
}

// Check validation status
if tx.Validated {
    fmt.Println("Transaction is validated on ledger")
} else {
    fmt.Println("Transaction is pending")
}

// Or use convenience method
result, err := client.VerifyTransaction(ctx, hash)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %s\n", result.Status) // "validated" or "pending"
```

### Subscribe to Real-time Transactions

```go
// Subscribe to account
stream, err := client.SubscribeToAccount(ctx, address)
if err != nil {
    log.Fatal(err)
}
defer client.Unsubscribe(ctx, address)

// Listen for transactions
for msg := range stream {
    if msg.Transaction != nil {
        fmt.Printf("New transaction: %s\n", msg.Transaction.Hash)
        fmt.Printf("Type: %s\n", msg.Transaction.TransactionType)
        fmt.Printf("Validated: %v\n", msg.Validated)
    }
}
```

### Error Handling

Use `errors.Is()` to check for specific error types:

```go
import "errors"

info, err := client.GetAccountInfo(ctx, address)
if err != nil {
    switch {
    case errors.Is(err, xrpl.ErrNotConnected):
        // Reconnect and retry
        client.Connect(ctx)
    case errors.Is(err, xrpl.ErrInvalidAddress):
        // Handle invalid address
        fmt.Println("Invalid address format")
    case errors.Is(err, xrpl.ErrTimeout):
        // Handle timeout
        fmt.Println("Request timed out")
    default:
        // Handle other errors
        log.Fatal(err)
    }
}
```

## Thread Safety

Client is safe for concurrent use by multiple goroutines. All operations can be called simultaneously from different goroutines:

```go
var wg sync.WaitGroup

// Query multiple accounts concurrently
for _, address := range addresses {
    wg.Add(1)
    go func(addr string) {
        defer wg.Done()
        info, err := client.GetAccountInfo(ctx, addr)
        // Handle result...
    }(address)
}

wg.Wait()
```

### Utility Functions

```go
// Convert drops to XRP
xrp, err := xrpl.DropsToXRP("1000000")
fmt.Println(xrp) // "1.000000"

// Convert XRP to drops
drops, err := xrpl.XRPToDrops(1.5)
fmt.Println(drops) // "1500000"

// Validate address
if xrpl.IsValidAddress("rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY") {
    fmt.Println("Valid address")
}

// Validate transaction hash
if xrpl.IsValidTxHash(hash) {
    fmt.Println("Valid hash")
}
```

## Error Types

The library defines sentinel errors for common conditions:

| Error | Description |
|-------|-------------|
| `ErrNotConnected` | Client is not connected to XRPL |
| `ErrInvalidAddress` | Invalid XRPL address format |
| `ErrInvalidHash` | Invalid transaction hash format |
| `ErrTimeout` | Request timeout |
| `ErrRequestFailed` | XRPL server rejected the request |
| `ErrInvalidResponse` | Unexpected response format |
| `ErrAlreadyConnected` | Client is already connected |
| `ErrConnectionClosed` | Connection was closed |

## Network Endpoints

### Testnet

```go
client := xrpl.New("wss://s.altnet.rippletest.net:51233")
```

### Mainnet

```go
client := xrpl.New("wss://s1.ripple.com")      // Ripple's server
client := xrpl.New("wss://xrplcluster.com")    // Public cluster
```

### Devnet

```go
client := xrpl.New("wss://s.devnet.rippletest.net:51233")
```

## Testing
    }(address)
}

wg.Wait()
```

## Testing

Run unit tests:

```bash
go test ./pkg/xrpl
```

Run with race detector:

```bash
go test -race ./pkg/xrpl
```

Run benchmarks:

```bash
go test -bench=. ./pkg/xrpl
```

Run integration tests (requires network):

```bash
go test -tags=integration ./pkg/xrpl
```

## Design Principles

This library follows the principles from "Let's Go" and "Let's Go Further" by Alex Edwards:

1. **Standard library first**: Minimal dependencies
2. **Explicit over implicit**: Clear error handling, no magic
3. **Idiomatic Go**: Follow Go conventions and patterns
4. **Testable design**: Dependency injection, interfaces for mocking
5. **Graceful operations**: Proper startup, shutdown, and error recovery
6. **Context propagation**: All blocking operations accept `context.Context`

## Examples

See [example_test.go](example_test.go) for comprehensive examples that demonstrate all features.

## Documentation

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev) or by running:

```bash
go doc -all github.com/yourusername/excise-tax-backend/v2/pkg/xrpl
```

## License

See LICENSE file in the repository root.

## Contributing

Contributions are welcome! Please ensure:

- Code follows Go conventions (`go fmt`, `go vet`)
- All tests pass (`go test ./...`)
- New features include tests and examples
- Public APIs are documented with godoc comments

## Resources

- [XRP Ledger Documentation](https://xrpl.org/)
- [WebSocket API Reference](https://xrpl.org/websocket-api.html)
- [Let's Go by Alex Edwards](https://lets-go.alexedwards.net/)
- [Effective Go](https://go.dev/doc/effective_go)
