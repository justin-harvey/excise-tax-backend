-- =====================================================
-- Database Views Migration
-- Creates views for dashboard metrics and reporting
-- =====================================================

-- =====================================================
-- 1. Dashboard Metrics View
-- Provides aggregated metrics for admin dashboard
-- =====================================================
CREATE OR REPLACE VIEW dashboard_metrics AS
SELECT
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_reports,
    COUNT(CASE WHEN status = 'needs_correction' THEN 1 END) as needs_correction,
    COUNT(CASE WHEN status = 'under_review' THEN 1 END) as under_review,
    COUNT(CASE WHEN status = 'approved' AND DATE_TRUNC('month', reviewed_at) = DATE_TRUNC('month', CURRENT_DATE) THEN 1 END) as approved_this_month,
    COUNT(CASE WHEN status = 'approved' AND EXTRACT(YEAR FROM reviewed_at) = EXTRACT(YEAR FROM CURRENT_DATE) THEN 1 END) as approved_this_year,
    SUM(CASE WHEN status = 'approved' AND DATE_TRUNC('month', reviewed_at) = DATE_TRUNC('month', CURRENT_DATE) THEN tax_amount_calculated ELSE 0 END) as tax_collected_this_month,
    SUM(CASE WHEN status = 'approved' AND EXTRACT(YEAR FROM reviewed_at) = EXTRACT(YEAR FROM CURRENT_DATE) THEN tax_amount_calculated ELSE 0 END) as tax_collected_this_year,
    COUNT(CASE WHEN status = 'rejected' THEN 1 END) as rejected_reports,
    AVG(CASE WHEN status = 'approved' AND reviewed_at IS NOT NULL THEN EXTRACT(EPOCH FROM (reviewed_at - submission_date)) / 86400 END) as avg_review_time_days
FROM tax_reports;

COMMENT ON VIEW dashboard_metrics IS 'Aggregated metrics for admin dashboard';

-- =====================================================
-- 2. Payment Summary View
-- Provides payment statistics grouped by manufacturer
-- =====================================================
CREATE OR REPLACE VIEW payment_summary AS
SELECT
    p.manufacturer_id,
    m.company_name,
    m.tax_id,
    m.business_type,
    COUNT(p.id) as total_payments,
    SUM(CASE WHEN p.status = 'completed' THEN p.amount_usd ELSE 0 END) as total_paid,
    SUM(CASE WHEN p.status = 'pending' THEN p.amount_usd ELSE 0 END) as pending_amount,
    SUM(CASE WHEN p.status = 'processing' THEN p.amount_usd ELSE 0 END) as processing_amount,
    SUM(CASE WHEN p.status = 'failed' THEN p.amount_usd ELSE 0 END) as failed_amount,
    COUNT(CASE WHEN p.payment_method = 'xrpl' THEN 1 END) as xrpl_payment_count,
    COUNT(CASE WHEN p.payment_method = 'ach' THEN 1 END) as ach_payment_count,
    COUNT(CASE WHEN p.payment_method = 'credit_card' THEN 1 END) as credit_card_payment_count,
    AVG(CASE WHEN p.payment_method = 'xrpl' AND p.status = 'completed' THEN EXTRACT(EPOCH FROM (p.processed_at - p.created_at)) END) as avg_xrpl_processing_time_seconds,
    MAX(p.payment_date) as last_payment_date,
    MIN(p.payment_date) as first_payment_date
FROM payments p
JOIN manufacturers m ON p.manufacturer_id = m.id
GROUP BY p.manufacturer_id, m.company_name, m.tax_id, m.business_type;

COMMENT ON VIEW payment_summary IS 'Payment statistics grouped by manufacturer';

-- =====================================================
-- 3. XRPL Payment Statistics View
-- Daily statistics for XRPL payments
-- =====================================================
CREATE OR REPLACE VIEW xrpl_payment_stats AS
SELECT
    DATE(xp.created_at) as payment_date,
    xp.status,
    COUNT(*) as payment_count,
    SUM(p.amount_usd) as total_usd,
    SUM(xp.xrp_amount) as total_xrp,
    AVG(xp.exchange_rate) as avg_exchange_rate,
    MIN(xp.exchange_rate) as min_exchange_rate,
    MAX(xp.exchange_rate) as max_exchange_rate,
    AVG(CASE WHEN xp.status = 'confirmed' AND xp.confirmed_at IS NOT NULL
        THEN EXTRACT(EPOCH FROM (xp.confirmed_at - xp.created_at)) / 60 END) as avg_confirmation_time_minutes
FROM xrpl_payments xp
JOIN payments p ON xp.payment_id = p.id
GROUP BY DATE(xp.created_at), xp.status
ORDER BY payment_date DESC, xp.status;

COMMENT ON VIEW xrpl_payment_stats IS 'Daily statistics for XRPL payments';

-- =====================================================
-- 4. Manufacturer Report Summary View
-- Report statistics grouped by manufacturer
-- =====================================================
CREATE OR REPLACE VIEW manufacturer_report_summary AS
SELECT
    tr.manufacturer_id,
    m.company_name,
    m.business_type,
    m.is_approved,
    COUNT(tr.id) as total_reports,
    COUNT(CASE WHEN tr.status = 'pending' THEN 1 END) as pending_reports,
    COUNT(CASE WHEN tr.status = 'under_review' THEN 1 END) as under_review_reports,
    COUNT(CASE WHEN tr.status = 'approved' THEN 1 END) as approved_reports,
    COUNT(CASE WHEN tr.status = 'needs_correction' THEN 1 END) as needs_correction_reports,
    COUNT(CASE WHEN tr.status = 'rejected' THEN 1 END) as rejected_reports,
    SUM(CASE WHEN tr.status = 'approved' THEN tr.tax_amount_calculated ELSE 0 END) as total_tax_approved,
    SUM(CASE WHEN tr.status = 'approved' AND EXTRACT(YEAR FROM tr.reviewed_at) = EXTRACT(YEAR FROM CURRENT_DATE)
        THEN tr.tax_amount_calculated ELSE 0 END) as tax_this_year,
    MAX(tr.submission_date) as last_submission_date,
    MIN(tr.submission_date) as first_submission_date
FROM tax_reports tr
JOIN manufacturers m ON tr.manufacturer_id = m.id
GROUP BY tr.manufacturer_id, m.company_name, m.business_type, m.is_approved;

COMMENT ON VIEW manufacturer_report_summary IS 'Report statistics grouped by manufacturer';

-- =====================================================
-- 5. Recent Activity View
-- Recent activity across the system for audit dashboard
-- =====================================================
CREATE OR REPLACE VIEW recent_activity AS
SELECT
    al.id,
    al.user_id,
    u.email as user_email,
    CONCAT(u.first_name, ' ', u.last_name) as user_name,
    u.role as user_role,
    al.action,
    al.resource_type,
    al.resource_id,
    al.ip_address,
    al.created_at
FROM audit_log al
LEFT JOIN users u ON al.user_id = u.id
ORDER BY al.created_at DESC
LIMIT 100;

COMMENT ON VIEW recent_activity IS 'Recent activity across the system (last 100 entries)';

-- =====================================================
-- 6. Tax Rate History View
-- Current and historical tax rates with status
-- =====================================================
CREATE OR REPLACE VIEW tax_rate_history AS
SELECT
    id,
    product_type,
    rate_per_unit,
    unit_type,
    effective_date,
    expiration_date,
    description,
    created_at,
    CASE
        WHEN effective_date > CURRENT_DATE THEN 'future'
        WHEN expiration_date IS NULL OR expiration_date > CURRENT_DATE THEN 'current'
        ELSE 'expired'
    END as status,
    CASE
        WHEN expiration_date IS NOT NULL THEN (expiration_date - effective_date)
        ELSE (CURRENT_DATE - effective_date)
    END as duration_days
FROM tax_rates
ORDER BY product_type, effective_date DESC;

COMMENT ON VIEW tax_rate_history IS 'Current and historical tax rates with calculated status';

-- =====================================================
-- 7. Daily Settlement Summary View
-- Summary of XRPL settlement batches
-- =====================================================
CREATE OR REPLACE VIEW daily_settlement_summary AS
SELECT
    batch_date,
    payment_count,
    total_xrp,
    total_usd,
    average_exchange_rate,
    CASE
        WHEN exchange_fee IS NOT NULL AND total_usd > 0 THEN
            ROUND(((exchange_fee / total_usd) * 100)::numeric, 4)
        ELSE NULL
    END as fee_percentage,
    status,
    exchange_name,
    processed_at,
    completed_at,
    CASE
        WHEN completed_at IS NOT NULL AND created_at IS NOT NULL THEN
            EXTRACT(EPOCH FROM (completed_at - created_at)) / 3600
        ELSE NULL
    END as processing_time_hours
FROM xrpl_settlement_batches
ORDER BY batch_date DESC;

COMMENT ON VIEW daily_settlement_summary IS 'Summary of XRPL settlement batches with calculated metrics';

-- =====================================================
-- Migration Complete
-- =====================================================
