#!/bin/bash

# =====================================================
# Migration Verification Runner
# Runs verification script and checks for issues
# =====================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERIFY_SQL="$SCRIPT_DIR/verify_migrations.sql"

# Default database connection
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-excise_tax_portal}"

# Check if PostgreSQL client is installed
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql (PostgreSQL client) is not installed.${NC}"
    exit 1
fi

# Check if verification script exists
if [ ! -f "$VERIFY_SQL" ]; then
    echo -e "${RED}Error: Verification script not found at $VERIFY_SQL${NC}"
    exit 1
fi

echo -e "${GREEN}Running migration verification...${NC}"
echo ""

# Run verification script
PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$VERIFY_SQL"

exit_code=$?

echo ""
if [ $exit_code -eq 0 ]; then
    echo -e "${GREEN}Verification completed successfully!${NC}"
else
    echo -e "${RED}Verification failed with exit code $exit_code${NC}"
fi

exit $exit_code
