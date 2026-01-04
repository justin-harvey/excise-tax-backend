#!/bin/bash

# =====================================================
# Database Migration Script
# Uses golang-migrate to manage PostgreSQL migrations
# =====================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATIONS_DIR="$PROJECT_ROOT/migrations"

# Default database connection (can be overridden by environment variables)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-excise_tax_portal}"

# Construct database URL
DATABASE_URL="${DATABASE_URL:-postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable}"

# Check if golang-migrate is installed
if ! command -v migrate &> /dev/null; then
    echo -e "${RED}Error: golang-migrate is not installed.${NC}"
    echo "Please install it using one of the following methods:"
    echo ""
    echo "macOS (Homebrew):"
    echo "  brew install golang-migrate"
    echo ""
    echo "Linux (curl):"
    echo "  curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz"
    echo "  sudo mv migrate /usr/local/bin/"
    echo ""
    echo "Windows (Chocolatey):"
    echo "  choco install migrate"
    echo ""
    echo "Go install:"
    echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Print usage information
usage() {
    echo "Usage: $0 {up|down|drop|force|version|create} [arguments]"
    echo ""
    echo "Commands:"
    echo "  up [N]          Apply all or N up migrations"
    echo "  down [N]        Apply all or N down migrations"
    echo "  drop            Drop everything inside database"
    echo "  force VERSION   Set version V but don't run migration (ignores dirty state)"
    echo "  version         Print current migration version"
    echo "  create NAME     Create a new migration with the given name"
    echo ""
    echo "Environment Variables:"
    echo "  DATABASE_URL    PostgreSQL connection string (default: constructed from below)"
    echo "  DB_HOST         Database host (default: localhost)"
    echo "  DB_PORT         Database port (default: 5432)"
    echo "  DB_USER         Database user (default: postgres)"
    echo "  DB_PASSWORD     Database password (default: postgres)"
    echo "  DB_NAME         Database name (default: excise_tax_portal)"
    echo ""
    echo "Examples:"
    echo "  $0 up                    # Apply all pending migrations"
    echo "  $0 up 1                  # Apply next migration"
    echo "  $0 down 1                # Rollback last migration"
    echo "  $0 version               # Show current version"
    echo "  $0 create add_new_table  # Create new migration files"
    exit 1
}

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${RED}Error: Migrations directory not found at $MIGRATIONS_DIR${NC}"
    exit 1
fi

# Main command handler
case "${1:-}" in
    up)
        echo -e "${GREEN}Applying migrations...${NC}"
        if [ -n "${2:-}" ]; then
            migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up "$2"
        else
            migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
        fi
        echo -e "${GREEN}Migrations applied successfully!${NC}"
        ;;

    down)
        echo -e "${YELLOW}Rolling back migrations...${NC}"
        if [ -n "${2:-}" ]; then
            migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down "$2"
        else
            echo -e "${RED}Warning: This will rollback ALL migrations!${NC}"
            read -p "Are you sure? (yes/no): " confirm
            if [ "$confirm" = "yes" ]; then
                migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down
            else
                echo "Aborted."
                exit 0
            fi
        fi
        echo -e "${GREEN}Rollback completed!${NC}"
        ;;

    drop)
        echo -e "${RED}WARNING: This will drop all tables and data!${NC}"
        read -p "Are you absolutely sure? Type 'DROP ALL DATA' to confirm: " confirm
        if [ "$confirm" = "DROP ALL DATA" ]; then
            migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" drop
            echo -e "${GREEN}Database dropped successfully!${NC}"
        else
            echo "Aborted."
            exit 0
        fi
        ;;

    force)
        if [ -z "${2:-}" ]; then
            echo -e "${RED}Error: Version number required${NC}"
            echo "Usage: $0 force VERSION"
            exit 1
        fi
        echo -e "${YELLOW}Forcing version to $2...${NC}"
        migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" force "$2"
        echo -e "${GREEN}Version forced successfully!${NC}"
        ;;

    version)
        echo -e "${GREEN}Current migration version:${NC}"
        migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version
        ;;

    create)
        if [ -z "${2:-}" ]; then
            echo -e "${RED}Error: Migration name required${NC}"
            echo "Usage: $0 create NAME"
            exit 1
        fi

        # Get the next migration number
        LAST_MIGRATION=$(ls -1 "$MIGRATIONS_DIR" | grep -E '^[0-9]+_' | tail -n 1 | cut -d'_' -f1)
        if [ -z "$LAST_MIGRATION" ]; then
            NEXT_NUMBER="000001"
        else
            NEXT_NUMBER=$(printf "%06d" $((10#$LAST_MIGRATION + 1)))
        fi

        MIGRATION_NAME="${2}"
        UP_FILE="$MIGRATIONS_DIR/${NEXT_NUMBER}_${MIGRATION_NAME}.up.sql"
        DOWN_FILE="$MIGRATIONS_DIR/${NEXT_NUMBER}_${MIGRATION_NAME}.down.sql"

        # Create up migration
        cat > "$UP_FILE" << EOF
-- =====================================================
-- Migration: ${MIGRATION_NAME}
-- =====================================================

-- Add your migration SQL here

-- =====================================================
-- Migration Complete
-- =====================================================
EOF

        # Create down migration
        cat > "$DOWN_FILE" << EOF
-- =====================================================
-- Rollback: ${MIGRATION_NAME}
-- =====================================================

-- Add your rollback SQL here

-- =====================================================
-- Rollback Complete
-- =====================================================
EOF

        echo -e "${GREEN}Created migration files:${NC}"
        echo "  - $UP_FILE"
        echo "  - $DOWN_FILE"
        ;;

    *)
        usage
        ;;
esac
