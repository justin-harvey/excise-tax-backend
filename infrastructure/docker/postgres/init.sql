-- Excise Tax Portal Database Initialization Script
-- This script runs automatically when the PostgreSQL container is first created

\echo 'Creating excise_tax_db database and extensions...'

-- Connect to the default database
\c excise_tax_db;

-- Create UUID extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
\echo 'Created uuid-ossp extension'

-- Create pgcrypto extension for encryption functions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
\echo 'Created pgcrypto extension'

-- Create pg_trgm extension for fuzzy text search
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
\echo 'Created pg_trgm extension'

-- Create citext extension for case-insensitive text
CREATE EXTENSION IF NOT EXISTS "citext";
\echo 'Created citext extension'

-- Create btree_gin extension for better indexing
CREATE EXTENSION IF NOT EXISTS "btree_gin";
\echo 'Created btree_gin extension'

-- Create additional schemas for organization
CREATE SCHEMA IF NOT EXISTS audit;
\echo 'Created audit schema'

CREATE SCHEMA IF NOT EXISTS reporting;
\echo 'Created reporting schema'

CREATE SCHEMA IF NOT EXISTS integration;
\echo 'Created integration schema'

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE excise_tax_db TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA audit TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA reporting TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA integration TO postgres;

-- Create audit log trigger function (for future use)
CREATE OR REPLACE FUNCTION audit.audit_trigger_func()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit.audit_log (
            table_name,
            operation,
            new_data,
            changed_by,
            changed_at
        ) VALUES (
            TG_TABLE_NAME,
            TG_OP,
            row_to_json(NEW),
            current_user,
            current_timestamp
        );
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit.audit_log (
            table_name,
            operation,
            old_data,
            new_data,
            changed_by,
            changed_at
        ) VALUES (
            TG_TABLE_NAME,
            TG_OP,
            row_to_json(OLD),
            row_to_json(NEW),
            current_user,
            current_timestamp
        );
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit.audit_log (
            table_name,
            operation,
            old_data,
            changed_by,
            changed_at
        ) VALUES (
            TG_TABLE_NAME,
            TG_OP,
            row_to_json(OLD),
            current_user,
            current_timestamp
        );
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

\echo 'Created audit trigger function'

-- Create audit log table
CREATE TABLE IF NOT EXISTS audit.audit_log (
    id BIGSERIAL PRIMARY KEY,
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    old_data JSONB,
    new_data JSONB,
    changed_by TEXT NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_log_table_name ON audit.audit_log(table_name);
CREATE INDEX IF NOT EXISTS idx_audit_log_changed_at ON audit.audit_log(changed_at);
CREATE INDEX IF NOT EXISTS idx_audit_log_changed_by ON audit.audit_log(changed_by);

\echo 'Created audit log table'

-- Set timezone to UTC
ALTER DATABASE excise_tax_db SET timezone TO 'UTC';

-- Update PostgreSQL configuration for better performance
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
ALTER SYSTEM SET work_mem = '4MB';
ALTER SYSTEM SET min_wal_size = '1GB';
ALTER SYSTEM SET max_wal_size = '4GB';

\echo 'Database initialization completed successfully!'
\echo 'Database: excise_tax_db'
\echo 'Extensions: uuid-ossp, pgcrypto, pg_trgm, citext, btree_gin'
\echo 'Schemas: public, audit, reporting, integration'
