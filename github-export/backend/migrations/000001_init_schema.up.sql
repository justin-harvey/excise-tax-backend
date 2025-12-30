-- =====================================================
-- Initial Schema Migration
-- Creates core tables for authentication, user management, and auditing
-- =====================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- 1. Users Table
-- Stores user accounts for manufacturers and administrators
-- =====================================================
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    role VARCHAR(50) NOT NULL CHECK (role IN ('manufacturer', 'admin', 'super_admin')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for users table
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

COMMENT ON TABLE users IS 'User accounts for authentication and authorization';
COMMENT ON COLUMN users.role IS 'User role: manufacturer, admin, super_admin';
COMMENT ON COLUMN users.is_active IS 'Flag to enable/disable user access';

-- =====================================================
-- 2. Manufacturers Table
-- Stores manufacturer business information and registration details
-- =====================================================
CREATE TABLE manufacturers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50) UNIQUE NOT NULL,
    license_number VARCHAR(100) UNIQUE NOT NULL,
    business_type VARCHAR(50) CHECK (business_type IN ('brewery', 'winery', 'distillery')),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(2),
    zip_code VARCHAR(10),
    phone VARCHAR(20),
    website VARCHAR(255),
    is_approved BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for manufacturers table
CREATE INDEX idx_manufacturers_user_id ON manufacturers(user_id);
CREATE INDEX idx_manufacturers_tax_id ON manufacturers(tax_id);
CREATE INDEX idx_manufacturers_license_number ON manufacturers(license_number);
CREATE INDEX idx_manufacturers_is_approved ON manufacturers(is_approved);
CREATE INDEX idx_manufacturers_business_type ON manufacturers(business_type);
CREATE INDEX idx_manufacturers_created_at ON manufacturers(created_at DESC);

COMMENT ON TABLE manufacturers IS 'Manufacturer business registration and details';
COMMENT ON COLUMN manufacturers.user_id IS 'Reference to associated user account';
COMMENT ON COLUMN manufacturers.business_type IS 'Type of beverage manufacturer: brewery, winery, distillery';
COMMENT ON COLUMN manufacturers.is_approved IS 'Admin approval status for manufacturer registration';

-- =====================================================
-- 3. Audit Log Table
-- Complete audit trail of all system actions
-- =====================================================
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id BIGINT,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for audit_log table
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_action ON audit_log(action);
CREATE INDEX idx_audit_log_resource_type ON audit_log(resource_type);
CREATE INDEX idx_audit_log_resource_id ON audit_log(resource_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);
CREATE INDEX idx_audit_log_ip_address ON audit_log(ip_address);

COMMENT ON TABLE audit_log IS 'Complete audit trail of all system actions';
COMMENT ON COLUMN audit_log.action IS 'Action performed: login, payment_created, report_approved, etc.';
COMMENT ON COLUMN audit_log.resource_type IS 'Type of resource affected: payment, report, user, etc.';
COMMENT ON COLUMN audit_log.metadata IS 'Additional JSON metadata about the action';

-- =====================================================
-- 4. Sessions Table
-- Session storage for authentication (Redis backup)
-- =====================================================
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for sessions table
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

COMMENT ON TABLE sessions IS 'Session storage for authentication (PostgreSQL backup for Redis)';
COMMENT ON COLUMN sessions.data IS 'Session data stored as JSON';

-- =====================================================
-- 5. Triggers for updated_at columns
-- =====================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to users table
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Apply updated_at trigger to manufacturers table
CREATE TRIGGER update_manufacturers_updated_at
    BEFORE UPDATE ON manufacturers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- Migration Complete
-- =====================================================
