# Payment Repository

Database layer for payment persistence using PostgreSQL with JSONB storage.

## Overview

The payment repository provides a clean interface for storing and retrieving payment transactions with their complete event history. All data is stored in a single `payments` table using PostgreSQL's JSONB type for flexible schema and efficient querying.

## Schema Design

### Payments Table

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    state VARCHAR(20) NOT NULL,
    transaction_data JSONB NOT NULL,  -- Full PaymentTransaction
    events JSONB NOT NULL,            -- Array of PaymentEvents
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

### Indexes

- **B-tree indexes**: Fast lookups on state, type, addresses, timestamps
- **GIN indexes**: Efficient JSONB queries on transaction_data and events
- **Composite indexes**: Optimized for common filter combinations

## Usage

### Initialize Repository

```go
import (
    "database/sql"
    "github.com/maxfelker/excise-tax-backend/v2/internal/payment/repository"
)

db, err := sql.Open("postgres", connStr)
if err != nil {
    log.Fatal(err)
}

repo := repository.NewPaymentRepository(db)
```

### Create Payment

```go
tx := &payment.PaymentTransaction{
    ID:        "pay_abc123",
    Type:      payment.PaymentTypeXRPL,
    Amount:    "100",
    Currency:  "XRP",
    From:      "rSender...",
    To:        "rReceiver...",
    State:     payment.StatePending,
    Events:    []payment.PaymentEvent{...},
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

err := repo.Create(ctx, tx)
```

### Get Payment

```go
tx, err := repo.Get(ctx, "pay_abc123")
if err != nil {
    if errors.Is(err, payment.ErrPaymentNotFound) {
        // Handle not found
    }
}
```

### Update Payment

```go
tx.State = payment.StateCompleted
err := repo.Update(ctx, tx)
```

### List Payments with Filters

```go
filters := payment.ListPaymentsFilters{
    State:  &statePending,
    Type:   &typeXRPL,
    Limit:  10,
    Offset: 0,
}

payments, err := repo.List(ctx, filters)
```

### Append Event (Atomic)

```go
event := payment.PaymentEvent{
    ID:            "evt_123",
    TransactionID: "pay_abc123",
    Type:          payment.EventPaymentCompleted,
    State:         payment.StateCompleted,
    Timestamp:     time.Now(),
}

err := repo.AppendEvent(ctx, "pay_abc123", event)
```

### Query by Metadata

```go
// Find payments by order_id in metadata
payments, err := repo.QueryByMetadata(ctx, "order_id", "12345")
```

## JSONB Query Examples

### Query by Event Type

```sql
-- Find payments with failed events
SELECT * FROM payments
WHERE events @> '[{"type": "payment_failed"}]';
```

### Query by Metadata Field

```sql
-- Find payments by order ID
SELECT * FROM payments
WHERE transaction_data->'metadata'->>'order_id' = '12345';
```

### Count Events

```sql
-- Count total events per payment
SELECT 
    id,
    jsonb_array_length(events) as event_count
FROM payments
ORDER BY event_count DESC;
```

### Filter by XRPL Transaction Hash

```sql
-- Find payment by XRPL transaction hash in event details
SELECT * FROM payments
WHERE events @> '[{"details": {"network_tx_hash": "ABC123..."}}]';
```

## Performance Considerations

1. **GIN Indexes**: Enable fast JSONB queries but increase write overhead
2. **Event Appending**: Use `AppendEvent()` for atomic JSONB array operations
3. **Pagination**: Always use LIMIT and OFFSET for large result sets
4. **Metadata Queries**: Create specific GIN indexes for frequently queried paths

## Best Practices

1. **Immutable Events**: Never modify events, only append new ones
2. **Atomic Updates**: Use `AppendEvent()` instead of reading, modifying, updating
3. **Error Handling**: Check for `ErrPaymentNotFound` on Get/Update operations
4. **Context Propagation**: Always pass context for cancellation support
5. **Transaction Wrapping**: Wrap multiple operations in database transactions

## Migration

Run the migration to create the payments table:

```bash
# Up migration
psql -U user -d dbname -f migrations/000005_add_payments.up.sql

# Down migration (rollback)
psql -U user -d dbname -f migrations/000005_add_payments.down.sql
```

## Testing

Unit tests use an in-memory SQLite database or test containers:

```go
func TestPaymentRepository(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repository.NewPaymentRepository(db)
    
    // Test Create
    tx := createTestPayment()
    err := repo.Create(context.Background(), tx)
    assert.NoError(t, err)
    
    // Test Get
    retrieved, err := repo.Get(context.Background(), tx.ID)
    assert.NoError(t, err)
    assert.Equal(t, tx.ID, retrieved.ID)
}
```

## Future Enhancements

- [ ] Soft delete support
- [ ] Archive old payments to separate table
- [ ] Full-text search on metadata
- [ ] Webhook event tracking
- [ ] Payment refund support
- [ ] Multi-currency support with exchange rates
