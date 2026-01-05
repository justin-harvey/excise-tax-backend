-- Drop trigger
DROP TRIGGER IF EXISTS update_payments_updated_at_trigger ON payments;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_payments_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_payments_event_types;
DROP INDEX IF EXISTS idx_payments_metadata;
DROP INDEX IF EXISTS idx_payments_events_gin;
DROP INDEX IF EXISTS idx_payments_transaction_data_gin;
DROP INDEX IF EXISTS idx_payments_state_type;
DROP INDEX IF EXISTS idx_payments_to_address;
DROP INDEX IF EXISTS idx_payments_from_address;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_type;
DROP INDEX IF EXISTS idx_payments_state;

-- Drop table
DROP TABLE IF EXISTS payments;
