# V2 Platform - Excise Tax Backend

Lightweight, composable services following Unix philosophy and Go best practices.

## Project Structure

```
v2/
├── cmd/
│   └── xrpl-cli/          # XRPL command-line tool
├── internal/
│   └── xrpl/              # XRPL client library
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.22 or later
- Access to XRPL testnet/mainnet

### Installation

```bash
cd v2
go mod download
go build -o xrpl-cli ./cmd/xrpl-cli
```

### Usage

```bash
# Get account balance on testnet (default)
./xrpl-cli balance rN7n7otQDd6FczFgLdlqtyMVrn3NnrcVc3

# Get detailed account info
./xrpl-cli info rN7n7otQDd6FczFgLdlqtyMVrn3NnrcVc3

# Use mainnet
./xrpl-cli -network mainnet balance rN7n7otQDd6FczFgLdlqtyMVrn3NnrcVc3

# Unix filter pattern - extract just the balance number
./xrpl-cli balance rN7n7otQDd6FczFgLdlqtyMVrn3NnrcVc3 | cut -d' ' -f1
```

## Development

### Run without installing

```bash
go run ./cmd/xrpl-cli balance rN7n7otQDd6FczFgLdlqtyMVrn3NnrcVc3
```

### Run tests

```bash
go test ./...
```

### Format code

```bash
go fmt ./...
```

## Architecture

Following the Unix Philosophy:
- **Small is beautiful**: Focused CLI tool for XRPL operations
- **Make each program a filter**: Accepts input, produces output
- **Build prototypes quickly**: Start simple, iterate
- **Standard library first**: Minimal dependencies

Following Let's Go principles:
- Explicit error handling
- Context for cancellation
- No global state
- Idiomatic Go

## Next Steps

See [MIGRATION_PLAN.md](MIGRATION_PLAN.md) for the full migration roadmap.
