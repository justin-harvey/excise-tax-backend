-- =====================================================
-- Tax and Reporting Migration
-- Creates tables for tax reporting and rate management
-- =====================================================

-- =====================================================
-- 1. Tax Reports Table
-- Stores tax reports submitted by manufacturers
-- =====================================================
CREATE TABLE tax_reports (
    id BIGSERIAL PRIMARY KEY,
    manufacturer_id BIGINT NOT NULL REFERENCES manufacturers(id) ON DELETE RESTRICT,
    report_type VARCHAR(50) NOT NULL CHECK (report_type IN ('monthly', 'quarterly', 'annual')),
    form_number VARCHAR(20),
    report_period_start DATE NOT NULL,
    report_period_end DATE NOT NULL,
    submission_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('draft', 'pending', 'under_review', 'approved', 'needs_correction', 'rejected')),
    reviewer_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,
    review_notes TEXT,
    production_data JSONB,
    tax_amount_calculated DECIMAL(15, 2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Ensure valid date ranges
    CONSTRAINT valid_report_period CHECK (report_period_end >= report_period_start)
);

-- Indexes for tax_reports table
CREATE INDEX idx_tax_reports_manufacturer_id ON tax_reports(manufacturer_id);
CREATE INDEX idx_tax_reports_status ON tax_reports(status);
CREATE INDEX idx_tax_reports_report_type ON tax_reports(report_type);
CREATE INDEX idx_tax_reports_form_number ON tax_reports(form_number);
CREATE INDEX idx_tax_reports_report_period ON tax_reports(report_period_start, report_period_end);
CREATE INDEX idx_tax_reports_submission_date ON tax_reports(submission_date DESC);
CREATE INDEX idx_tax_reports_reviewer_id ON tax_reports(reviewer_id);
CREATE INDEX idx_tax_reports_reviewed_at ON tax_reports(reviewed_at DESC);
CREATE INDEX idx_tax_reports_created_at ON tax_reports(created_at DESC);

COMMENT ON TABLE tax_reports IS 'Tax reports submitted by manufacturers';
COMMENT ON COLUMN tax_reports.report_type IS 'Report frequency: monthly, quarterly, annual';
COMMENT ON COLUMN tax_reports.form_number IS 'Tax form number: 35-7136, 35-7131, etc.';
COMMENT ON COLUMN tax_reports.status IS 'Report status: draft, pending, under_review, approved, needs_correction, rejected';
COMMENT ON COLUMN tax_reports.production_data IS 'Production data as JSON (gallons, barrels, cases, etc.)';
COMMENT ON COLUMN tax_reports.tax_amount_calculated IS 'Calculated tax amount in USD';

-- =====================================================
-- 2. Tax Rates Table
-- Stores tax rates for different product types
-- =====================================================
CREATE TABLE tax_rates (
    id BIGSERIAL PRIMARY KEY,
    product_type VARCHAR(50) NOT NULL CHECK (product_type IN ('beer', 'wine', 'spirits', 'other')),
    rate_per_unit DECIMAL(10, 4) NOT NULL CHECK (rate_per_unit >= 0),
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('gallon', 'barrel', 'case', 'liter')),
    effective_date DATE NOT NULL,
    expiration_date DATE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Ensure valid date ranges
    CONSTRAINT valid_effective_dates CHECK (expiration_date IS NULL OR expiration_date > effective_date)
);

-- Indexes for tax_rates table
CREATE INDEX idx_tax_rates_product_type ON tax_rates(product_type);
CREATE INDEX idx_tax_rates_effective_date ON tax_rates(effective_date DESC);
CREATE INDEX idx_tax_rates_expiration_date ON tax_rates(expiration_date);
CREATE INDEX idx_tax_rates_product_type_effective ON tax_rates(product_type, effective_date DESC);
CREATE INDEX idx_tax_rates_created_at ON tax_rates(created_at DESC);

COMMENT ON TABLE tax_rates IS 'Tax rates for different product types and time periods';
COMMENT ON COLUMN tax_rates.product_type IS 'Type of product: beer, wine, spirits, other';
COMMENT ON COLUMN tax_rates.rate_per_unit IS 'Tax rate per unit in USD';
COMMENT ON COLUMN tax_rates.unit_type IS 'Unit of measurement: gallon, barrel, case, liter';
COMMENT ON COLUMN tax_rates.effective_date IS 'Date when the rate becomes effective';
COMMENT ON COLUMN tax_rates.expiration_date IS 'Date when the rate expires (NULL if current)';

-- =====================================================
-- 3. Add report_id foreign key to payments table
-- =====================================================
ALTER TABLE payments
ADD CONSTRAINT fk_payments_report_id
FOREIGN KEY (report_id) REFERENCES tax_reports(id) ON DELETE SET NULL;

CREATE INDEX idx_payments_report_id_fk ON payments(report_id) WHERE report_id IS NOT NULL;

-- =====================================================
-- 4. Triggers for updated_at columns
-- =====================================================

-- Apply updated_at trigger to tax_reports table
CREATE TRIGGER update_tax_reports_updated_at
    BEFORE UPDATE ON tax_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- 5. Helper Functions
-- =====================================================

-- Function: Get current tax rate for a product type
CREATE OR REPLACE FUNCTION get_current_tax_rate(
    p_product_type VARCHAR(50),
    p_unit_type VARCHAR(50),
    p_date DATE DEFAULT CURRENT_DATE
)
RETURNS DECIMAL(10, 4) AS $$
DECLARE
    v_rate DECIMAL(10, 4);
BEGIN
    SELECT rate_per_unit INTO v_rate
    FROM tax_rates
    WHERE product_type = p_product_type
    AND unit_type = p_unit_type
    AND effective_date <= p_date
    AND (expiration_date IS NULL OR expiration_date > p_date)
    ORDER BY effective_date DESC
    LIMIT 1;

    RETURN COALESCE(v_rate, 0);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_current_tax_rate IS 'Retrieves the current tax rate for a product type and unit';

-- Function: Calculate tax for a report
CREATE OR REPLACE FUNCTION calculate_report_tax(
    p_report_id BIGINT
)
RETURNS DECIMAL(15, 2) AS $$
DECLARE
    v_production_data JSONB;
    v_total_tax DECIMAL(15, 2) := 0;
    v_product_type VARCHAR(50);
    v_item JSONB;
    v_quantity DECIMAL(15, 2);
    v_unit_type VARCHAR(50);
    v_rate DECIMAL(10, 4);
BEGIN
    -- Get production data from the report
    SELECT production_data INTO v_production_data
    FROM tax_reports
    WHERE id = p_report_id;

    IF v_production_data IS NULL THEN
        RETURN 0;
    END IF;

    -- Iterate through production items and calculate tax
    -- This is a simplified example - actual implementation would depend on JSON structure
    -- You may need to adjust based on your production_data JSON schema

    RETURN v_total_tax;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION calculate_report_tax IS 'Calculates total tax for a report based on production data';

-- =====================================================
-- 6. Insert Default Tax Rates
-- =====================================================

-- Insert default tax rates for beer (example rates - adjust as needed)
INSERT INTO tax_rates (product_type, rate_per_unit, unit_type, effective_date, description)
VALUES
    ('beer', 0.11, 'gallon', '2024-01-01', 'Standard beer tax rate per gallon'),
    ('beer', 3.50, 'barrel', '2024-01-01', 'Standard beer tax rate per barrel (31 gallons)'),
    ('wine', 0.67, 'gallon', '2024-01-01', 'Standard wine tax rate per gallon'),
    ('spirits', 13.50, 'gallon', '2024-01-01', 'Standard spirits tax rate per gallon (distilled spirits)')
ON CONFLICT DO NOTHING;

-- =====================================================
-- Migration Complete
-- =====================================================
