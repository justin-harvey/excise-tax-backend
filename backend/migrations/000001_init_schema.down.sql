-- =====================================================
-- Rollback Initial Schema Migration
-- Drops all core tables and functions
-- =====================================================

-- Drop triggers
DROP TRIGGER IF EXISTS update_manufacturers_updated_at ON manufacturers;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS audit_log CASCADE;
DROP TABLE IF EXISTS manufacturers CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop extensions (commented out to avoid affecting other schemas)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
