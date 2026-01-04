// Package database provides PostgreSQL database connection management with connection pooling,
// health checks, and transaction support.
//
// Example usage:
//
//	config := &database.Config{
//		Host:     "localhost",
//		Port:     5432,
//		User:     "postgres",
//		Password: "password",
//		DBName:   "excise_tax",
//		SSLMode:  "disable",
//	}
//
//	db, err := database.NewPostgresDB(ctx, config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Use the database
//	var count int
//	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds the PostgreSQL database configuration parameters.
type Config struct {
	Host              string
	Port              int
	User              string
	Password          string
	DBName            string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() *Config {
	return &Config{
		Host:              "localhost",
		Port:              5432,
		SSLMode:           "disable",
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}
}

// PostgresDB wraps pgxpool.Pool to provide additional functionality.
type PostgresDB struct {
	Pool *pgxpool.Pool
}

// NewPostgresDB creates a new PostgreSQL database connection pool.
// It returns an error if the connection cannot be established or if the initial health check fails.
func NewPostgresDB(ctx context.Context, cfg *Config) (*PostgresDB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config cannot be nil")
	}

	// Build connection string
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Parse connection string and configure pool
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set pool configuration
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	db := &PostgresDB{Pool: pool}

	// Perform initial health check
	if err := db.HealthCheck(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("initial health check failed: %w", err)
	}

	return db, nil
}

// HealthCheck verifies that the database connection is alive and functional.
func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	if db.Pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result int
	err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("health check query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("health check returned unexpected result: %d", result)
	}

	return nil
}

// Close gracefully closes the database connection pool.
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Query executes a query that returns rows.
func (db *PostgresDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.Pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (db *PostgresDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.Pool.QueryRow(ctx, sql, args...)
}

// Exec executes a query without returning any rows.
func (db *PostgresDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.Pool.Exec(ctx, sql, args...)
}

// TxFunc is a function that executes within a database transaction.
type TxFunc func(pgx.Tx) error

// WithTransaction executes the provided function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
//
// Example usage:
//
//	err := db.WithTransaction(ctx, func(tx pgx.Tx) error {
//		_, err := tx.Exec(ctx, "INSERT INTO users (email) VALUES ($1)", "test@example.com")
//		if err != nil {
//			return err
//		}
//		_, err = tx.Exec(ctx, "INSERT INTO audit_log (action) VALUES ($1)", "user_created")
//		return err
//	})
func (db *PostgresDB) WithTransaction(ctx context.Context, fn TxFunc) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BeginTx starts a new transaction with the provided options.
func (db *PostgresDB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return db.Pool.BeginTx(ctx, txOptions)
}

// Ping verifies a connection to the database is still alive, establishing a connection if necessary.
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Stats returns database statistics.
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}
