package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/excise-tax-portal/backend/internal/payment/model"
	"github.com/excise-tax-portal/backend/pkg/database"

	"github.com/jackc/pgx/v5"
)

// PaymentRepository handles payment data access.
type PaymentRepository struct {
	db *database.PostgresDB
}

// NewPaymentRepository creates a new payment repository.
func NewPaymentRepository(db *database.PostgresDB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// CreatePayment creates a new payment with XRPL details in a transaction.
func (r *PaymentRepository) CreatePayment(ctx context.Context, payment *model.Payment, xrplPayment *model.XRPLPayment) error {
	return r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Insert payment
		query := `
			INSERT INTO payments (
				manufacturer_id, report_id, payment_method, amount_usd, status,
				metadata, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, created_at, updated_at
		`

		err := tx.QueryRow(ctx, query,
			payment.ManufacturerID,
			payment.ReportID,
			payment.PaymentMethod,
			payment.AmountUSD,
			payment.Status,
			payment.Metadata,
			time.Now(),
			time.Now(),
		).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)

		if err != nil {
			return fmt.Errorf("failed to insert payment: %w", err)
		}

		// Insert XRPL payment if provided
		if xrplPayment != nil {
			xrplPayment.PaymentID = payment.ID

			xrplQuery := `
				INSERT INTO xrpl_payments (
					payment_id, xrp_amount, exchange_rate, destination_address,
					destination_tag, status, qr_code, expires_at, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
				RETURNING id, created_at, updated_at
			`

			err = tx.QueryRow(ctx, xrplQuery,
				xrplPayment.PaymentID,
				xrplPayment.XRPAmount,
				xrplPayment.ExchangeRate,
				xrplPayment.DestinationAddress,
				xrplPayment.DestinationTag,
				xrplPayment.Status,
				xrplPayment.QRCode,
				xrplPayment.ExpiresAt,
				time.Now(),
				time.Now(),
			).Scan(&xrplPayment.ID, &xrplPayment.CreatedAt, &xrplPayment.UpdatedAt)

			if err != nil {
				return fmt.Errorf("failed to insert xrpl payment: %w", err)
			}
		}

		return nil
	})
}

// GetPaymentByID retrieves a payment by ID with XRPL details.
func (r *PaymentRepository) GetPaymentByID(ctx context.Context, id int64) (*model.PaymentWithXRPL, error) {
	query := `
		SELECT
			p.id, p.manufacturer_id, p.report_id, p.payment_method, p.amount_usd,
			p.status, p.transaction_id, p.confirmation_number, p.payment_date,
			p.processed_at, p.metadata, p.created_at, p.updated_at,
			xp.id, xp.payment_id, xp.xrp_amount, xp.exchange_rate,
			xp.destination_address, xp.destination_tag, xp.source_address,
			xp.tx_hash, xp.ledger_index, xp.fee_xrp, xp.status,
			xp.qr_code, xp.expires_at, xp.confirmed_at, xp.created_at, xp.updated_at
		FROM payments p
		LEFT JOIN xrpl_payments xp ON p.id = xp.payment_id
		WHERE p.id = $1
	`

	var result model.PaymentWithXRPL
	var xrpl model.XRPLPayment
	var hasXRPL bool

	err := r.db.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.ManufacturerID, &result.ReportID, &result.PaymentMethod,
		&result.AmountUSD, &result.Status, &result.TransactionID, &result.ConfirmationNumber,
		&result.PaymentDate, &result.ProcessedAt, &result.Metadata, &result.CreatedAt, &result.UpdatedAt,
		&xrpl.ID, &xrpl.PaymentID, &xrpl.XRPAmount, &xrpl.ExchangeRate,
		&xrpl.DestinationAddress, &xrpl.DestinationTag, &xrpl.SourceAddress,
		&xrpl.TxHash, &xrpl.LedgerIndex, &xrpl.FeeXRP, &xrpl.Status,
		&xrpl.QRCode, &xrpl.ExpiresAt, &xrpl.ConfirmedAt, &xrpl.CreatedAt, &xrpl.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("payment not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if xrpl.ID.Valid {
		hasXRPL = true
		result.XRPLPayment = &xrpl
	}

	if !hasXRPL && result.PaymentMethod == model.PaymentMethodXRPL {
		return nil, fmt.Errorf("xrpl payment data missing for payment: %d", id)
	}

	return &result, nil
}

// GetPaymentByDestinationTag retrieves a pending payment by destination tag.
func (r *PaymentRepository) GetPaymentByDestinationTag(ctx context.Context, tag uint32) (*model.PaymentWithXRPL, error) {
	query := `
		SELECT
			p.id, p.manufacturer_id, p.report_id, p.payment_method, p.amount_usd,
			p.status, p.transaction_id, p.confirmation_number, p.payment_date,
			p.processed_at, p.metadata, p.created_at, p.updated_at,
			xp.id, xp.payment_id, xp.xrp_amount, xp.exchange_rate,
			xp.destination_address, xp.destination_tag, xp.source_address,
			xp.tx_hash, xp.ledger_index, xp.fee_xrp, xp.status,
			xp.qr_code, xp.expires_at, xp.confirmed_at, xp.created_at, xp.updated_at
		FROM payments p
		JOIN xrpl_payments xp ON p.id = xp.payment_id
		WHERE xp.destination_tag = $1
		AND xp.status = 'pending'
		AND xp.expires_at > NOW()
		LIMIT 1
	`

	var result model.PaymentWithXRPL
	var xrpl model.XRPLPayment

	err := r.db.QueryRow(ctx, query, tag).Scan(
		&result.ID, &result.ManufacturerID, &result.ReportID, &result.PaymentMethod,
		&result.AmountUSD, &result.Status, &result.TransactionID, &result.ConfirmationNumber,
		&result.PaymentDate, &result.ProcessedAt, &result.Metadata, &result.CreatedAt, &result.UpdatedAt,
		&xrpl.ID, &xrpl.PaymentID, &xrpl.XRPAmount, &xrpl.ExchangeRate,
		&xrpl.DestinationAddress, &xrpl.DestinationTag, &xrpl.SourceAddress,
		&xrpl.TxHash, &xrpl.LedgerIndex, &xrpl.FeeXRP, &xrpl.Status,
		&xrpl.QRCode, &xrpl.ExpiresAt, &xrpl.ConfirmedAt, &xrpl.CreatedAt, &xrpl.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("pending payment not found for tag: %d", tag)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by tag: %w", err)
	}

	result.XRPLPayment = &xrpl
	return &result, nil
}

// UpdatePaymentStatus updates the payment status.
func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, id int64, status string) error {
	query := `
		UPDATE payments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	tag, err := r.db.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("payment not found: %d", id)
	}

	return nil
}

// UpdateXRPLPaymentStatus updates the XRPL payment status and transaction details.
func (r *PaymentRepository) UpdateXRPLPaymentStatus(ctx context.Context, paymentID int64, status string, txHash *string, ledgerIndex *int64, confirmedAt *time.Time) error {
	query := `
		UPDATE xrpl_payments
		SET status = $1, tx_hash = $2, ledger_index = $3, confirmed_at = $4, updated_at = $5
		WHERE payment_id = $6
	`

	tag, err := r.db.Exec(ctx, query, status, txHash, ledgerIndex, confirmedAt, time.Now(), paymentID)
	if err != nil {
		return fmt.Errorf("failed to update xrpl payment status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("xrpl payment not found for payment: %d", paymentID)
	}

	return nil
}

// GetPaymentsByManufacturer retrieves all payments for a manufacturer.
func (r *PaymentRepository) GetPaymentsByManufacturer(ctx context.Context, manufacturerID int64, limit, offset int) ([]*model.PaymentWithXRPL, error) {
	query := `
		SELECT
			p.id, p.manufacturer_id, p.report_id, p.payment_method, p.amount_usd,
			p.status, p.transaction_id, p.confirmation_number, p.payment_date,
			p.processed_at, p.metadata, p.created_at, p.updated_at,
			xp.id, xp.payment_id, xp.xrp_amount, xp.exchange_rate,
			xp.destination_address, xp.destination_tag, xp.source_address,
			xp.tx_hash, xp.ledger_index, xp.fee_xrp, xp.status,
			xp.qr_code, xp.expires_at, xp.confirmed_at, xp.created_at, xp.updated_at
		FROM payments p
		LEFT JOIN xrpl_payments xp ON p.id = xp.payment_id
		WHERE p.manufacturer_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, manufacturerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}
	defer rows.Close()

	var payments []*model.PaymentWithXRPL
	for rows.Next() {
		var p model.PaymentWithXRPL
		var xrpl model.XRPLPayment

		err := rows.Scan(
			&p.ID, &p.ManufacturerID, &p.ReportID, &p.PaymentMethod,
			&p.AmountUSD, &p.Status, &p.TransactionID, &p.ConfirmationNumber,
			&p.PaymentDate, &p.ProcessedAt, &p.Metadata, &p.CreatedAt, &p.UpdatedAt,
			&xrpl.ID, &xrpl.PaymentID, &xrpl.XRPAmount, &xrpl.ExchangeRate,
			&xrpl.DestinationAddress, &xrpl.DestinationTag, &xrpl.SourceAddress,
			&xrpl.TxHash, &xrpl.LedgerIndex, &xrpl.FeeXRP, &xrpl.Status,
			&xrpl.QRCode, &xrpl.ExpiresAt, &xrpl.ConfirmedAt, &xrpl.CreatedAt, &xrpl.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		if xrpl.ID.Valid {
			p.XRPLPayment = &xrpl
		}

		payments = append(payments, &p)
	}

	return payments, nil
}

// GetPendingPayments retrieves all pending XRPL payments.
func (r *PaymentRepository) GetPendingPayments(ctx context.Context) ([]*model.PaymentWithXRPL, error) {
	query := `
		SELECT
			p.id, p.manufacturer_id, p.report_id, p.payment_method, p.amount_usd,
			p.status, p.transaction_id, p.confirmation_number, p.payment_date,
			p.processed_at, p.metadata, p.created_at, p.updated_at,
			xp.id, xp.payment_id, xp.xrp_amount, xp.exchange_rate,
			xp.destination_address, xp.destination_tag, xp.source_address,
			xp.tx_hash, xp.ledger_index, xp.fee_xrp, xp.status,
			xp.qr_code, xp.expires_at, xp.confirmed_at, xp.created_at, xp.updated_at
		FROM payments p
		JOIN xrpl_payments xp ON p.id = xp.payment_id
		WHERE xp.status = 'pending'
		AND xp.expires_at > NOW()
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending payments: %w", err)
	}
	defer rows.Close()

	var payments []*model.PaymentWithXRPL
	for rows.Next() {
		var p model.PaymentWithXRPL
		var xrpl model.XRPLPayment

		err := rows.Scan(
			&p.ID, &p.ManufacturerID, &p.ReportID, &p.PaymentMethod,
			&p.AmountUSD, &p.Status, &p.TransactionID, &p.ConfirmationNumber,
			&p.PaymentDate, &p.ProcessedAt, &p.Metadata, &p.CreatedAt, &p.UpdatedAt,
			&xrpl.ID, &xrpl.PaymentID, &xrpl.XRPAmount, &xrpl.ExchangeRate,
			&xrpl.DestinationAddress, &xrpl.DestinationTag, &xrpl.SourceAddress,
			&xrpl.TxHash, &xrpl.LedgerIndex, &xrpl.FeeXRP, &xrpl.Status,
			&xrpl.QRCode, &xrpl.ExpiresAt, &xrpl.ConfirmedAt, &xrpl.CreatedAt, &xrpl.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		p.XRPLPayment = &xrpl
		payments = append(payments, &p)
	}

	return payments, nil
}

// SaveExchangeRate saves an exchange rate to the database.
func (r *PaymentRepository) SaveExchangeRate(ctx context.Context, rate *model.ExchangeRate) error {
	query := `
		INSERT INTO exchange_rates (
			source, xrp_usd_rate, bid, ask, volume_24h, timestamp, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		rate.Source,
		rate.XRPUSDRate,
		rate.Bid,
		rate.Ask,
		rate.Volume24h,
		rate.Timestamp,
		time.Now(),
	).Scan(&rate.ID, &rate.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to save exchange rate: %w", err)
	}

	return nil
}

// LogTransaction logs an XRPL transaction.
func (r *PaymentRepository) LogTransaction(ctx context.Context, log *model.XRPLTransactionLog) error {
	query := `
		INSERT INTO xrpl_transaction_log (
			payment_id, tx_hash, tx_type, from_address, to_address,
			amount_drops, destination_tag, ledger_index, ledger_hash,
			transaction_date, raw_transaction, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		log.PaymentID,
		log.TxHash,
		log.TxType,
		log.FromAddress,
		log.ToAddress,
		log.AmountDrops,
		log.DestinationTag,
		log.LedgerIndex,
		log.LedgerHash,
		log.TransactionDate,
		log.RawTransaction,
		log.Notes,
		time.Now(),
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to log transaction: %w", err)
	}

	return nil
}

// GetLatestExchangeRate retrieves the most recent exchange rate.
func (r *PaymentRepository) GetLatestExchangeRate(ctx context.Context, source string) (*model.ExchangeRate, error) {
	query := `
		SELECT id, source, xrp_usd_rate, bid, ask, volume_24h, timestamp, created_at
		FROM exchange_rates
		WHERE source = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var rate model.ExchangeRate
	err := r.db.QueryRow(ctx, query, source).Scan(
		&rate.ID,
		&rate.Source,
		&rate.XRPUSDRate,
		&rate.Bid,
		&rate.Ask,
		&rate.Volume24h,
		&rate.Timestamp,
		&rate.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest exchange rate: %w", err)
	}

	return &rate, nil
}

// SaveAggregatedExchangeRate saves the aggregated exchange rate with source details.
func (r *PaymentRepository) SaveAggregatedExchangeRate(ctx context.Context, rate float64, sources []map[string]interface{}) error {
	// Convert sources to JSON
	sourcesJSON, err := json.Marshal(sources)
	if err != nil {
		return fmt.Errorf("failed to marshal sources: %w", err)
	}

	query := `
		INSERT INTO exchange_rates (
			source, xrp_usd_rate, timestamp, created_at
		) VALUES ($1, $2, $3, $4)
	`

	_, err = r.db.Exec(ctx, query,
		model.ExchangeSourceAggregated,
		rate,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save aggregated exchange rate: %w", err)
	}

	return nil
}
