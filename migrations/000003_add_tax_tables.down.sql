-- =====================================================
-- Rollback Tax and Reporting Migration
-- Drops all tax-related tables and functions
-- =====================================================

-- Drop functions
DROP FUNCTION IF EXISTS calculate_report_tax(BIGINT);
DROP FUNCTION IF EXISTS get_current_tax_rate(VARCHAR(50), VARCHAR(50), DATE);

-- Drop triggers
DROP TRIGGER IF EXISTS update_tax_reports_updated_at ON tax_reports;

-- Remove foreign key constraint from payments table
ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_report_id;

-- Drop tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS tax_rates CASCADE;
DROP TABLE IF EXISTS tax_reports CASCADE;
