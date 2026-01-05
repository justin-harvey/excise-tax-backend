# Payment CLI

A command-line tool for processing payments with support for XRPL, credit card, and ACH payment methods.

## Installation

```bash
cd v2
go build -o payment-cli ./cmd/payment-cli
```

## Usage

### Create a Payment

From flags:
```bash
./payment-cli create --type xrpl --amount 100 \
  --from rSender... --to rReceiver...
```

From JSON file:
```bash
./payment-cli create example_payment.json
```

From stdin:
```bash
cat payment.json | ./payment-cli create -
```

Example payment JSON:
```json
{
  "type": "xrpl",
  "amount": "100",
  "currency": "XRP",
  "from": "rSenderAddress...",
  "to": "rReceiverAddress...",
  "metadata": {
    "order_id": "12345"
  }
}
```

### Get Payment Details

```bash
./payment-cli get pay_abc123
```

With JSON output:
```bash
./payment-cli get pay_abc123 --json
```

### List Payments

List all payments:
```bash
./payment-cli list
```

Filter by state:
```bash
./payment-cli list --state pending
```

Filter by type:
```bash
./payment-cli list --type xrpl
```

With pagination:
```bash
./payment-cli list --limit 20 --offset 10
```

### Verify Payment Status

```bash
./payment-cli verify pay_abc123
```

### Retry Failed Payment

```bash
./payment-cli retry pay_abc123
```

### View Event History

```bash
./payment-cli events pay_abc123
```

## Payment Types

- **xrpl**: XRP Ledger payments (supported)
- **credit_card**: Credit card payments (coming soon)
- **ach**: ACH bank transfers (coming soon)

## Transaction States

- **pending**: Payment initiated, awaiting confirmation
- **completed**: Payment successfully processed
- **in_review**: Payment requires manual review
- **failed**: Payment failed with reason

## Event Types

Events are immutable and track the lifecycle of a payment:

- **payment_initiated**: Payment created
- **payment_submitted**: Submitted to payment network
- **payment_verified**: Payment verified on network
- **payment_retried**: Automatic retry attempted
- **payment_completed**: Payment successfully completed
- **payment_failed**: Payment failed with reason
- **payment_reviewed**: Manual review occurred

## Examples

### Complete Payment Workflow

```bash
# Create payment
PAYMENT_ID=$(./payment-cli create --type xrpl --amount 100 \
  --from rSender... --to rReceiver... --json | jq -r '.id')

# Wait a moment for processing
sleep 3

# Verify payment
./payment-cli verify $PAYMENT_ID

# View event history
./payment-cli events $PAYMENT_ID

# Get full details
./payment-cli get $PAYMENT_ID --json
```

### List and Filter

```bash
# Show all pending payments
./payment-cli list --state pending

# Show completed XRPL payments
./payment-cli list --state completed --type xrpl

# Export to JSON for analysis
./payment-cli list --json > payments.json
```

## Architecture

The payment CLI follows the Unix philosophy:

- **Small and focused**: One tool, one purpose
- **Composable**: Works with pipes and files
- **Plain text**: JSON input/output for scripting
- **Exit codes**: 0 (success), 1 (error), 2 (usage)

## Development

Run tests:
```bash
go test ./internal/payment/...
```

Build:
```bash
go build -o payment-cli ./cmd/payment-cli
```

## Next Steps

- [ ] Implement credit card payment processing
- [ ] Implement ACH payment processing
- [ ] Add database persistence (PostgreSQL)
- [ ] Build HTTP API wrapper
- [ ] Add webhook notifications
- [ ] Implement payment refunds
