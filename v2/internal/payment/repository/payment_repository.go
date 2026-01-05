package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/internal/payment"
)

// PaymentRepository handles database operations for payments
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}

// Create inserts a new payment into the database
func (r *PaymentRepository) Create(ctx context.Context, tx *payment.PaymentTransaction) error {
	query := `
		INSERT INTO payments (
			id, type, amount, currency, from_address, to_address, 
			state, transaction_data, events, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	// Marshal transaction data and events to JSONB
	txData, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction data: %w", err)
	}

	eventsData, err := json.Marshal(tx.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tx.ID,
		tx.Type,
		tx.Amount,
		tx.Currency,
		tx.From,
		tx.To,
		tx.State,
		txData,
		eventsData,
		tx.CreatedAt,
		tx.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

// Get retrieves a payment by ID
func (r *PaymentRepository) Get(ctx context.Context, id string) (*payment.PaymentTransaction, error) {
	query := `
		SELECT transaction_data, events
		FROM payments
		WHERE id = $1
	`

	var txData, eventsData []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(&txData, &eventsData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, payment.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Unmarshal transaction data
	var tx payment.PaymentTransaction
	if err := json.Unmarshal(txData, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction data: %w", err)
	}

	// Unmarshal events
	if err := json.Unmarshal(eventsData, &tx.Events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	return &tx, nil
}

// Update updates an existing payment
func (r *PaymentRepository) Update(ctx context.Context, tx *payment.PaymentTransaction) error {
	query := `
		UPDATE payments
		SET state = $2,
		    transaction_data = $3,
		    events = $4,
		    updated_at = $5
		WHERE id = $1
	`

	// Marshal transaction data and events
	txData, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction data: %w", err)
	}

	eventsData, err := json.Marshal(tx.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	tx.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query,
		tx.ID,
		tx.State,
		txData,
		eventsData,
		tx.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return payment.ErrPaymentNotFound
	}

	return nil
}

// List retrieves payments with filters
func (r *PaymentRepository) List(ctx context.Context, filters payment.ListPaymentsFilters) ([]*payment.PaymentTransaction, error) {
	query := `
		SELECT transaction_data, events
		FROM payments
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	// Add filters
	if filters.State != nil {
		query += fmt.Sprintf(" AND state = $%d", argCount)
		args = append(args, *filters.State)
		argCount++
	}

	if filters.Type != nil {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, *filters.Type)
		argCount++
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add limit and offset
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
		argCount++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	var payments []*payment.PaymentTransaction

	for rows.Next() {
		var txData, eventsData []byte
		if err := rows.Scan(&txData, &eventsData); err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		var tx payment.PaymentTransaction
		if err := json.Unmarshal(txData, &tx); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction data: %w", err)
		}

		if err := json.Unmarshal(eventsData, &tx.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}

		payments = append(payments, &tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return payments, nil
}

// AppendEvent appends a new event to a payment's event history
// This is done atomically using JSONB array append
func (r *PaymentRepository) AppendEvent(ctx context.Context, txID string, event payment.PaymentEvent) error {
	query := `
		UPDATE payments
		SET events = events || $2::jsonb,
		    state = $3,
		    updated_at = NOW()
		WHERE id = $1
	`

	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Wrap event in array for JSONB append
	eventArray := fmt.Sprintf("[%s]", string(eventData))

	result, err := r.db.ExecContext(ctx, query, txID, eventArray, event.State)
	if err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return payment.ErrPaymentNotFound
	}

	return nil
}

// GetByState retrieves all payments with a specific state
func (r *PaymentRepository) GetByState(ctx context.Context, state payment.TransactionState) ([]*payment.PaymentTransaction, error) {
	filters := payment.ListPaymentsFilters{
		State: &state,
	}
	return r.List(ctx, filters)
}

// GetByType retrieves all payments of a specific type
func (r *PaymentRepository) GetByType(ctx context.Context, paymentType payment.PaymentType) ([]*payment.PaymentTransaction, error) {
	filters := payment.ListPaymentsFilters{
		Type: &paymentType,
	}
	return r.List(ctx, filters)
}

// QueryByMetadata queries payments by metadata fields using JSONB operators
func (r *PaymentRepository) QueryByMetadata(ctx context.Context, jsonPath string, value interface{}) ([]*payment.PaymentTransaction, error) {
	query := `
		SELECT transaction_data, events
		FROM payments
		WHERE transaction_data->'metadata'->$1 = $2
		ORDER BY created_at DESC
	`

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, jsonPath, valueJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()

	var payments []*payment.PaymentTransaction

	for rows.Next() {
		var txData, eventsData []byte
		if err := rows.Scan(&txData, &eventsData); err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		var tx payment.PaymentTransaction
		if err := json.Unmarshal(txData, &tx); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction data: %w", err)
		}

		if err := json.Unmarshal(eventsData, &tx.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}

		payments = append(payments, &tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return payments, nil
}
