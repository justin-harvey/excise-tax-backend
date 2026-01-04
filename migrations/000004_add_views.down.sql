-- =====================================================
-- Rollback Database Views Migration
-- Drops all views
-- =====================================================

-- Drop all views
DROP VIEW IF EXISTS daily_settlement_summary;
DROP VIEW IF EXISTS tax_rate_history;
DROP VIEW IF EXISTS recent_activity;
DROP VIEW IF EXISTS manufacturer_report_summary;
DROP VIEW IF EXISTS xrpl_payment_stats;
DROP VIEW IF EXISTS payment_summary;
DROP VIEW IF EXISTS dashboard_metrics;
