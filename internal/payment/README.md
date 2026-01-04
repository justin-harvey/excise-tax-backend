# XRPL Payment Service

Complete XRPL (XRP Ledger) blockchain payment integration for the Louisiana Excise Tax Portal. This service provides **99.9% cost reduction** for payment processing by using cryptocurrency payments instead of traditional payment processors.

## Architecture Overview

The XRPL payment service is structured in clean layers following Domain-Driven Design principles:

```
internal/payment/
├── xrpl/               # XRPL blockchain integration layer
│   ├── client.go       # WebSocket client for XRPL network
│   ├── oracle.go       # Multi-source price oracle (Coinbase, Binance, Kraken, Bitstamp)
│   ├── monitor.go      # Real-time payment monitoring service
│   └── payment.go      # Payment processing and QR code generation
├── model/              # Domain models
│   └── payment.go      # Payment, XRPLPayment, ExchangeRate models
├── repository/         # Data access layer
│   └── payment_repository.go  # Database operations
├── service/            # Business logic layer
│   └── payment_service.go     # Payment orchestration
├── handler/            # HTTP handlers
│   └── payment_handler.go     # REST API endpoints
└── README.md
```

## Key Features

### 1. XRPL Client (`xrpl/client.go`)

WebSocket-based XRPL network client with:
- **Auto-reconnection**: Exponential backoff, max 10 attempts
- **Network selection**: Supports testnet and mainnet
- **Health checks**: Periodic connection validation
- **Account monitoring**: Real-time transaction subscriptions
- **Transaction verification**: Blockchain confirmation checks

**Key Functions:**
```go
NewClient(cfg *Config, logger *zap.Logger) (*Client, error)
Initialize(ctx context.Context) error
GetAccountBalance(ctx context.Context, address string) (float64, error)
SubscribeToAccount(ctx context.Context, address string, handler TransactionHandler) error
VerifyTransaction(ctx context.Context, txHash string) (*TransactionVerification, error)
```

### 2. Price Oracle (`xrpl/oracle.go`)

Multi-source exchange rate aggregation:
- **Sources**: Coinbase, Binance, Kraken, Bitstamp
- **Median calculation**: Outlier-resistant pricing
- **Outlier detection**: ±10% threshold from median
- **Background refresh**: 30-second update interval
- **Redis caching**: 60-second TTL
- **Volatility buffer**: 2% added to protect against price fluctuations

**Exchange Sources:**
- Coinbase: `https://api.coinbase.com/v2/prices/XRP-USD/spot`
- Binance: `https://api.binance.com/api/v3/ticker/price?symbol=XRPUSDT`
- Kraken: `https://api.kraken.com/0/public/Ticker?pair=XRPUSD`
- Bitstamp: `https://www.bitstamp.net/api/v2/ticker/xrpusd/`

**Key Functions:**
```go
NewPriceOracle(cfg *OracleConfig, logger *zap.Logger) *PriceOracle
GetExchangeRate(ctx context.Context) (*ExchangeRate, error)
ConvertUSDtoXRP(ctx context.Context, usdAmount float64, includeBuffer bool) (float64, error)
ConvertXRPtoUSD(ctx context.Context, xrpAmount float64) (float64, error)
```

### 3. Payment Monitor (`xrpl/monitor.go`)

Real-time payment tracking:
- **Goroutine-based**: Non-blocking payment monitoring
- **Context cancellation**: Proper cleanup on timeout
- **30-minute timeout**: Automatic expiration handling
- **Callback system**: Status change notifications
- **Expiration checker**: Periodic cleanup of expired payments

**Key Functions:**
```go
NewMonitorService(xrplClient *Client, logger *zap.Logger) *MonitorService
MonitorPayment(ctx context.Context, paymentID string, destinationTag uint32, expectedAmount int64, expiresAt time.Time, callback PaymentCallback) error
StopMonitoring(paymentID string)
```

### 4. Payment Processor (`xrpl/payment.go`)

Payment request creation and verification:
- **Destination tag generation**: Based on manufacturer ID (1,000,000 - 9,999,999)
- **QR code generation**: Mobile wallet compatible (format: `xrpl:{address}?dt={tag}&amount={xrp}`)
- **Transaction verification**: Blockchain confirmation checks

**Key Functions:**
```go
GenerateDestinationTag(manufacturerID int64) (uint32, error)
CreatePaymentRequest(ctx context.Context, req *CreatePaymentRequest) (*PaymentResponse, error)
GenerateQRCode(payment *PaymentResponse) (string, error)
VerifyTransaction(ctx context.Context, txHash string) (*TransactionVerification, error)
```

## Database Schema

### `payments` Table
```sql
id, manufacturer_id, report_id, payment_method, amount_usd,
status, transaction_id, confirmation_number, payment_date,
processed_at, metadata, created_at, updated_at
```

### `xrpl_payments` Table
```sql
id, payment_id, xrp_amount, exchange_rate, destination_address,
destination_tag, source_address, tx_hash, ledger_index, fee_xrp,
status, qr_code, expires_at, confirmed_at, created_at, updated_at
```

### `exchange_rates` Table
```sql
id, source, xrp_usd_rate, bid, ask, volume_24h, timestamp, created_at
```

### `xrpl_transaction_log` Table
```sql
id, payment_id, tx_hash, tx_type, from_address, to_address,
amount_drops, destination_tag, ledger_index, ledger_hash,
transaction_date, raw_transaction, notes, created_at
```

## API Endpoints

### Create XRPL Payment
```http
POST /api/v1/payments/xrpl/create
Content-Type: application/json

{
  "manufacturer_id": 123,
  "report_id": "456",
  "amount_usd": 1000.00,
  "description": "Excise tax payment",
  "manufacturer_email": "manufacturer@example.com"
}

Response:
{
  "success": true,
  "data": {
    "payment_id": 789,
    "xrpl_payment_id": 101,
    "manufacturer_id": 123,
    "amount_usd": 1000.00,
    "xrp_amount": 2000.40,
    "exchange_rate": 0.50,
    "destination_address": "rN7n7otQDd6FczFgLdlqtyMVrn3HMfXy4G",
    "destination_tag": 1000123,
    "qr_code": "data:image/png;base64,...",
    "status": "pending",
    "expires_at": "2025-12-29T15:30:00Z",
    "instructions": {
      "destination": "rN7n7otQDd6FczFgLdlqtyMVrn3HMfXy4G",
      "amount": 2000.40,
      "destination_tag": 1000123,
      "currency": "XRP",
      "network": "testnet"
    }
  }
}
```

### Get Payment Status
```http
GET /api/v1/payments/xrpl/:paymentId/status

Response:
{
  "success": true,
  "data": {
    "payment_id": 789,
    "status": "completed",
    "amount_usd": 1000.00,
    "payment_method": "xrpl",
    "xrpl": {
      "xrp_amount": 2000.40,
      "exchange_rate": 0.50,
      "destination_tag": 1000123,
      "status": "confirmed",
      "tx_hash": "ABC123...",
      "confirmed_at": "2025-12-29T15:25:00Z"
    }
  }
}
```

### Verify Payment
```http
POST /api/v1/payments/xrpl/:paymentId/verify
Content-Type: application/json

{
  "tx_hash": "ABC123..."
}

Response:
{
  "success": true,
  "data": {
    "payment_id": 789,
    "tx_hash": "ABC123...",
    "verified": true,
    "confirmed_at": "2025-12-29T15:25:00Z",
    "message": "Payment verified successfully"
  }
}
```

### Get Exchange Rate
```http
GET /api/v1/payments/exchange-rate

Response:
{
  "success": true,
  "data": {
    "xrp_usd": 0.50,
    "usd_xrp": 2.00,
    "sources": [
      {
        "name": "coinbase",
        "rate": 0.501,
        "timestamp": "2025-12-29T15:00:00Z"
      },
      {
        "name": "binance",
        "rate": 0.499,
        "timestamp": "2025-12-29T15:00:00Z"
      }
    ],
    "updated_at": "2025-12-29T15:00:00Z"
  }
}
```

### Convert Currency
```http
POST /api/v1/payments/convert
Content-Type: application/json

{
  "amount": 1000.00,
  "from": "USD",
  "to": "XRP"
}

Response:
{
  "success": true,
  "data": {
    "amount": 1000.00,
    "from_currency": "USD",
    "to_currency": "XRP",
    "result": 2000.00,
    "rate": 0.50,
    "timestamp": "2025-12-29T15:00:00Z"
  }
}
```

### List Payments
```http
GET /api/v1/payments?manufacturer_id=123&limit=50&offset=0

Response:
{
  "success": true,
  "data": {
    "payments": [...],
    "limit": 50,
    "offset": 0,
    "count": 25
  }
}
```

## Environment Variables

```bash
# Service
ENVIRONMENT=development
PORT=8082
LOG_LEVEL=info

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=excise_tax
DB_SSL_MODE=disable
DB_MAX_CONNS=25
DB_MIN_CONNS=5

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# XRPL Configuration
XRPL_NETWORK=testnet

# Testnet
XRPL_TESTNET_SERVER=wss://s.altnet.rippletest.net:51233
XRPL_testnet_ADDRESS=rN7n7otQDd6FczFgLdlqtyMVrn3HMfXy4G

# Mainnet (production)
XRPL_MAINNET_SERVER=wss://xrplcluster.com
XRPL_mainnet_ADDRESS=
```

## Configuration

### XRPL Network Settings

**Testnet:**
- Server: `wss://s.altnet.rippletest.net:51233`
- Fallback: `wss://testnet.xrpl-labs.com`
- Explorer: https://testnet.xrpl.org

**Mainnet:**
- Server: `wss://xrplcluster.com`
- Fallback: `wss://s1.ripple.com`, `wss://s2.ripple.com`
- Explorer: https://livenet.xrpl.org

### Payment Settings

- **Volatility Buffer**: 2% (protects against price fluctuations)
- **Payment Timeout**: 30 minutes
- **Destination Tag Range**: 1,000,000 - 9,999,999
- **Min Confirmations**: 1 (XRP transactions are final after 1 confirmation)

### Price Oracle Settings

- **Update Interval**: 30 seconds
- **Cache Duration**: 60 seconds
- **Outlier Threshold**: 10%
- **Min Sources**: 2 (out of 4)

## Running the Service

### Development
```bash
cd backend
go run cmd/payment-service/main.go
```

### Production
```bash
cd backend
go build -o payment-service cmd/payment-service/main.go
./payment-service
```

### Docker
```bash
docker build -t payment-service -f Dockerfile.payment .
docker run -p 8082:8082 --env-file .env payment-service
```

## Testing

### Unit Tests
```bash
go test ./internal/payment/...
```

### Integration Tests
```bash
go test -tags=integration ./internal/payment/...
```

### Manual Testing (Testnet)

1. Start the service
2. Create a payment request
3. Send XRP to the destination address with the specified destination tag
4. Monitor the payment status
5. Verify the transaction on the blockchain

## Cost Savings Analysis

**Traditional Payment Processing:**
- Credit Card Fee: 2.9% + $0.30 per transaction
- ACH Fee: $0.50 - $1.50 per transaction
- Wire Transfer Fee: $15 - $50 per transaction

**XRPL Payment Processing:**
- Transaction Fee: ~0.00001 XRP (~$0.000005 USD)
- **Cost Reduction**: 99.9%

**Example:**
- $1,000 payment via credit card: $29.30 fee
- $1,000 payment via XRPL: $0.000005 fee
- **Savings per transaction**: $29.30

For 10,000 payments/year at average $1,000:
- Traditional fees: $293,000/year
- XRPL fees: $0.05/year
- **Total annual savings**: $292,999.95

## Security Considerations

1. **Multi-signature Wallet**: State wallet requires 3 of 5 signers for withdrawals
2. **Destination Tags**: Unique per manufacturer to prevent payment misrouting
3. **Transaction Verification**: All payments verified on blockchain before confirmation
4. **Rate Limiting**: Protects against DDoS attacks
5. **Input Validation**: All inputs sanitized and validated
6. **Audit Trail**: Complete transaction log in database

## Dependencies

```go
github.com/gin-gonic/gin              // HTTP framework
github.com/jackc/pgx/v5               // PostgreSQL driver
github.com/redis/go-redis/v9          // Redis client
github.com/skip2/go-qrcode            // QR code generation
go.uber.org/zap                       // Structured logging
```

## Future Enhancements

1. **WebSocket Support**: Real-time payment notifications to frontend
2. **Email Notifications**: Automatic email on payment confirmation/expiration
3. **Refund Processing**: Automated refund system
4. **Settlement Automation**: Daily batch settlement to fiat
5. **Multi-currency Support**: Additional cryptocurrencies (BTC, ETH)
6. **Analytics Dashboard**: Payment metrics and reporting

## Support

For issues or questions:
- GitHub Issues: [excise-tax-portal/issues](https://github.com/excise-tax-portal/issues)
- Documentation: [BACKEND_INFRASTRUCTURE_SPEC.md](../../BACKEND_INFRASTRUCTURE_SPEC.md)

## License

Proprietary - Louisiana Cannabis Control (LCC)
