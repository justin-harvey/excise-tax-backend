-- =====================================================
-- Rollback XRPL Integration Migration
-- Drops all XRPL-related tables and functions
-- =====================================================

-- Drop functions
DROP FUNCTION IF EXISTS get_pending_xrpl_payment_by_tag(INT);

-- Drop triggers
DROP TRIGGER IF EXISTS update_xrpl_payments_updated_at ON xrpl_payments;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;

-- Drop tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS xrpl_settlement_batches CASCADE;
DROP TABLE IF EXISTS xrpl_transaction_log CASCADE;
DROP TABLE IF EXISTS exchange_rates CASCADE;
DROP TABLE IF EXISTS xrpl_payments CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
