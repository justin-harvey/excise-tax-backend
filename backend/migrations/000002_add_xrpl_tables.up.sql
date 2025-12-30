-- =====================================================
-- XRPL Integration Migration
-- Creates tables for XRP Ledger payment processing and exchange rate tracking
-- =====================================================

-- =====================================================
-- 1. Payments Table
-- Stores all payment requests across all payment methods
-- =====================================================
CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    manufacturer_id BIGINT NOT NULL REFERENCES manufacturers(id) ON DELETE RESTRICT,
    report_id BIGINT,
    payment_method VARCHAR(50) NOT NULL CHECK (payment_method IN ('xrpl', 'ach', 'credit_card', 'wire')),
    amount_usd DECIMAL(15, 2) NOT NULL CHECK (amount_usd > 0),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'refunded', 'expired')),
    transaction_id VARCHAR(255),
    confirmation_number VARCHAR(100),
    payment_date TIMESTAMP,
    processed_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for payments table
CREATE INDEX idx_payments_manufacturer_id ON payments(manufacturer_id);
CREATE INDEX idx_payments_report_id ON payments(report_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_payment_method ON payments(payment_method);
CREATE INDEX idx_payments_payment_date ON payments(payment_date);
CREATE INDEX idx_payments_transaction_id ON payments(transaction_id);
CREATE INDEX idx_payments_created_at ON payments(created_at DESC);

COMMENT ON TABLE payments IS 'Payment records for all payment methods';
COMMENT ON COLUMN payments.payment_method IS 'Payment method: xrpl, ach, credit_card, wire';
COMMENT ON COLUMN payments.status IS 'Payment status: pending, processing, completed, failed, refunded, expired';
COMMENT ON COLUMN payments.metadata IS 'Additional payment metadata as JSON';

-- =====================================================
-- 2. XRPL Payments Table
-- Specific details for XRP Ledger payments
-- =====================================================
CREATE TABLE xrpl_payments (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    xrp_amount DECIMAL(20, 6) NOT NULL CHECK (xrp_amount > 0),
    exchange_rate DECIMAL(10, 4) NOT NULL CHECK (exchange_rate > 0),
    destination_address VARCHAR(100) NOT NULL,
    destination_tag INT,
    source_address VARCHAR(100),
    tx_hash VARCHAR(100) UNIQUE,
    ledger_index BIGINT,
    fee_xrp DECIMAL(10, 6),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'expired', 'failed', 'refunded')),
    qr_code TEXT,
    expires_at TIMESTAMP,
    confirmed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for xrpl_payments table
CREATE INDEX idx_xrpl_payments_payment_id ON xrpl_payments(payment_id);
CREATE INDEX idx_xrpl_payments_tx_hash ON xrpl_payments(tx_hash);
CREATE INDEX idx_xrpl_payments_status ON xrpl_payments(status);
CREATE INDEX idx_xrpl_payments_destination_tag ON xrpl_payments(destination_tag);
CREATE INDEX idx_xrpl_payments_destination_address ON xrpl_payments(destination_address);
CREATE INDEX idx_xrpl_payments_source_address ON xrpl_payments(source_address);
CREATE INDEX idx_xrpl_payments_expires_at ON xrpl_payments(expires_at) WHERE status = 'pending';
CREATE INDEX idx_xrpl_payments_confirmed_at ON xrpl_payments(confirmed_at DESC);
CREATE INDEX idx_xrpl_payments_created_at ON xrpl_payments(created_at DESC);

COMMENT ON TABLE xrpl_payments IS 'XRP Ledger payment-specific details and transaction data';
COMMENT ON COLUMN xrpl_payments.xrp_amount IS 'Amount in XRP (6 decimal places)';
COMMENT ON COLUMN xrpl_payments.exchange_rate IS 'XRP/USD exchange rate at time of payment creation';
COMMENT ON COLUMN xrpl_payments.destination_tag IS 'XRPL destination tag for payment identification';
COMMENT ON COLUMN xrpl_payments.tx_hash IS 'XRPL transaction hash after confirmation';
COMMENT ON COLUMN xrpl_payments.qr_code IS 'Base64 encoded QR code for payment';

-- =====================================================
-- 3. Exchange Rates Table
-- Historical XRP/USD exchange rates from multiple sources
-- =====================================================
CREATE TABLE exchange_rates (
    id BIGSERIAL PRIMARY KEY,
    source VARCHAR(50) NOT NULL CHECK (source IN ('coinbase', 'binance', 'kraken', 'bitstamp', 'aggregated')),
    xrp_usd_rate DECIMAL(10, 6) NOT NULL CHECK (xrp_usd_rate > 0),
    bid DECIMAL(10, 6),
    ask DECIMAL(10, 6),
    volume_24h DECIMAL(20, 2),
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for exchange_rates table
CREATE INDEX idx_exchange_rates_timestamp ON exchange_rates(timestamp DESC);
CREATE INDEX idx_exchange_rates_source ON exchange_rates(source);
CREATE INDEX idx_exchange_rates_source_timestamp ON exchange_rates(source, timestamp DESC);
CREATE INDEX idx_exchange_rates_created_at ON exchange_rates(created_at DESC);

COMMENT ON TABLE exchange_rates IS 'Historical XRP/USD exchange rates from multiple sources';
COMMENT ON COLUMN exchange_rates.source IS 'Exchange source: coinbase, binance, kraken, bitstamp, aggregated';
COMMENT ON COLUMN exchange_rates.xrp_usd_rate IS 'XRP to USD exchange rate';
COMMENT ON COLUMN exchange_rates.timestamp IS 'Timestamp when the rate was fetched';

-- =====================================================
-- 4. XRPL Transaction Log Table
-- Complete audit trail of all XRPL transactions
-- =====================================================
CREATE TABLE xrpl_transaction_log (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT REFERENCES payments(id) ON DELETE SET NULL,
    tx_hash VARCHAR(100) NOT NULL,
    tx_type VARCHAR(50) NOT NULL,
    from_address VARCHAR(100),
    to_address VARCHAR(100),
    amount_drops BIGINT,
    destination_tag INT,
    ledger_index BIGINT,
    ledger_hash VARCHAR(100),
    transaction_date TIMESTAMP,
    raw_transaction JSONB NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for xrpl_transaction_log table
CREATE INDEX idx_xrpl_transaction_log_payment_id ON xrpl_transaction_log(payment_id);
CREATE INDEX idx_xrpl_transaction_log_tx_hash ON xrpl_transaction_log(tx_hash);
CREATE INDEX idx_xrpl_transaction_log_destination_tag ON xrpl_transaction_log(destination_tag);
CREATE INDEX idx_xrpl_transaction_log_ledger_index ON xrpl_transaction_log(ledger_index);
CREATE INDEX idx_xrpl_transaction_log_transaction_date ON xrpl_transaction_log(transaction_date DESC);
CREATE INDEX idx_xrpl_transaction_log_created_at ON xrpl_transaction_log(created_at DESC);

COMMENT ON TABLE xrpl_transaction_log IS 'Complete audit trail of all XRPL transactions';
COMMENT ON COLUMN xrpl_transaction_log.amount_drops IS 'Amount in XRP drops (1 XRP = 1,000,000 drops)';
COMMENT ON COLUMN xrpl_transaction_log.raw_transaction IS 'Complete raw transaction data from XRPL';

-- =====================================================
-- 5. XRPL Settlement Batches Table
-- Tracks daily settlement batches for XRP to USD conversion
-- =====================================================
CREATE TABLE xrpl_settlement_batches (
    id BIGSERIAL PRIMARY KEY,
    batch_date DATE NOT NULL UNIQUE,
    total_xrp DECIMAL(20, 6) NOT NULL DEFAULT 0,
    total_usd DECIMAL(15, 2) NOT NULL DEFAULT 0,
    average_exchange_rate DECIMAL(10, 4) NOT NULL,
    payment_count INT NOT NULL DEFAULT 0,
    exchange_name VARCHAR(100),
    exchange_tx_id VARCHAR(200),
    exchange_fee DECIMAL(10, 2),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    processed_by VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Indexes for xrpl_settlement_batches table
CREATE INDEX idx_xrpl_settlement_batches_batch_date ON xrpl_settlement_batches(batch_date DESC);
CREATE INDEX idx_xrpl_settlement_batches_status ON xrpl_settlement_batches(status);
CREATE INDEX idx_xrpl_settlement_batches_created_at ON xrpl_settlement_batches(created_at DESC);

COMMENT ON TABLE xrpl_settlement_batches IS 'Daily settlement batches for converting XRP to USD';
COMMENT ON COLUMN xrpl_settlement_batches.batch_date IS 'Date of the settlement batch';
COMMENT ON COLUMN xrpl_settlement_batches.exchange_tx_id IS 'Transaction ID from the exchange where XRP was converted';

-- =====================================================
-- 6. Triggers for updated_at columns
-- =====================================================

-- Apply updated_at trigger to payments table
CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Apply updated_at trigger to xrpl_payments table
CREATE TRIGGER update_xrpl_payments_updated_at
    BEFORE UPDATE ON xrpl_payments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- 7. Helper Functions
-- =====================================================

-- Function: Get pending payment by destination tag
CREATE OR REPLACE FUNCTION get_pending_xrpl_payment_by_tag(tag INT)
RETURNS TABLE (
    payment_id BIGINT,
    xrpl_payment_id BIGINT,
    manufacturer_id BIGINT,
    amount_usd DECIMAL(15, 2),
    xrp_amount DECIMAL(20, 6),
    exchange_rate DECIMAL(10, 4),
    destination_tag INT,
    destination_address VARCHAR(100),
    expires_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        p.id,
        xp.id,
        p.manufacturer_id,
        p.amount_usd,
        xp.xrp_amount,
        xp.exchange_rate,
        xp.destination_tag,
        xp.destination_address,
        xp.expires_at
    FROM xrpl_payments xp
    JOIN payments p ON xp.payment_id = p.id
    WHERE xp.destination_tag = tag
    AND xp.status = 'pending'
    AND xp.expires_at > CURRENT_TIMESTAMP
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_pending_xrpl_payment_by_tag IS 'Retrieves pending XRPL payment by destination tag';

-- =====================================================
-- Migration Complete
-- =====================================================
