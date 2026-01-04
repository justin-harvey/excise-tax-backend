-- =====================================================
-- Excise Tax Portal - PostgreSQL Initialization
-- =====================================================
-- This script is automatically executed when the database container starts
-- NOTE: The actual schema is managed by migrations in /migrations/*.sql
-- This file only sets up extensions and optimizes PostgreSQL settings
-- =====================================================

\echo 'Initializing excise_tax_db database...'

-- Create required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
\echo '✓ Created uuid-ossp extension'

-- Set timezone to UTC
ALTER DATABASE excise_tax_db SET timezone TO 'UTC';
\echo '✓ Set timezone to UTC'

-- Optimize PostgreSQL settings for performance
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
\echo '✓ Optimized PostgreSQL settings'

\echo ''
\echo '===================================================='
\echo 'Database initialization completed!'
\echo 'Database: excise_tax_db'
\echo 'Extensions: uuid-ossp'
\echo ''
\echo 'Run migrations to create schema:'
\echo '  docker exec -it excise-tax-dev migrate -path /app/migrations -database "postgresql://postgres:password@postgres:5432/excise_tax_db?sslmode=disable" up'
\echo '===================================================='
