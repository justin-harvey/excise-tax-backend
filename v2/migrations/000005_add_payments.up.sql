-- Create payments table with JSONB storage for transactions and events
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL CHECK (type IN ('credit_card', 'ach', 'xrpl')),
    amount DECIMAL(20, 8) NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    state VARCHAR(20) NOT NULL CHECK (state IN ('pending', 'completed', 'in_review', 'failed')),
    
    -- Store full transaction and events as JSONB
    transaction_data JSONB NOT NULL,
    events JSONB NOT NULL DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_payments_state ON payments(state);
CREATE INDEX IF NOT EXISTS idx_payments_type ON payments(type);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_from_address ON payments(from_address);
CREATE INDEX IF NOT EXISTS idx_payments_to_address ON payments(to_address);
CREATE INDEX IF NOT EXISTS idx_payments_state_type ON payments(state, type);

-- Create GIN indexes for JSONB columns to enable efficient queries
CREATE INDEX IF NOT EXISTS idx_payments_transaction_data_gin ON payments USING gin(transaction_data);
CREATE INDEX IF NOT EXISTS idx_payments_events_gin ON payments USING gin(events);

-- Create specific JSONB path indexes for common queries
CREATE INDEX IF NOT EXISTS idx_payments_metadata ON payments USING gin((transaction_data->'metadata'));
CREATE INDEX IF NOT EXISTS idx_payments_event_types ON payments USING gin((events));

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_payments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_payments_updated_at_trigger
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_payments_updated_at();

-- Add comments for documentation
COMMENT ON TABLE payments IS 'Payment transactions with full event history stored as JSONB';
COMMENT ON COLUMN payments.transaction_data IS 'Full PaymentTransaction object stored as JSONB';
COMMENT ON COLUMN payments.events IS 'Array of immutable PaymentEvent objects stored as JSONB';
COMMENT ON COLUMN payments.state IS 'Current state of the payment: pending, completed, in_review, or failed';
COMMENT ON COLUMN payments.type IS 'Payment method type: credit_card, ach, or xrpl';
