package model

import (
	"database/sql"
	"time"
)

// Payment represents a payment record in the database.
type Payment struct {
	ID                 int64          `json:"id" db:"id"`
	ManufacturerID     int64          `json:"manufacturer_id" db:"manufacturer_id"`
	ReportID           sql.NullInt64  `json:"report_id,omitempty" db:"report_id"`
	PaymentMethod      string         `json:"payment_method" db:"payment_method"`
	AmountUSD          float64        `json:"amount_usd" db:"amount_usd"`
	Status             string         `json:"status" db:"status"`
	TransactionID      sql.NullString `json:"transaction_id,omitempty" db:"transaction_id"`
	ConfirmationNumber sql.NullString `json:"confirmation_number,omitempty" db:"confirmation_number"`
	PaymentDate        sql.NullTime   `json:"payment_date,omitempty" db:"payment_date"`
	ProcessedAt        sql.NullTime   `json:"processed_at,omitempty" db:"processed_at"`
	Metadata           []byte         `json:"metadata,omitempty" db:"metadata"`
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
}

// XRPLPayment represents an XRPL-specific payment record.
type XRPLPayment struct {
	// ID and PaymentID are nullable: GetPaymentByID and GetPaymentsByManufacturer
	// read this struct back from a LEFT JOIN against payments, which is NULL
	// on every row that isn't an XRPL payment.
	ID                 sql.NullInt64   `json:"id" db:"id"`
	PaymentID          sql.NullInt64   `json:"payment_id" db:"payment_id"`
	XRPAmount          float64         `json:"xrp_amount" db:"xrp_amount"`
	ExchangeRate       float64         `json:"exchange_rate" db:"exchange_rate"`
	DestinationAddress string          `json:"destination_address" db:"destination_address"`
	DestinationTag     sql.NullInt32   `json:"destination_tag,omitempty" db:"destination_tag"`
	SourceAddress      sql.NullString  `json:"source_address,omitempty" db:"source_address"`
	TxHash             sql.NullString  `json:"tx_hash,omitempty" db:"tx_hash"`
	LedgerIndex        sql.NullInt64   `json:"ledger_index,omitempty" db:"ledger_index"`
	FeeXRP             sql.NullFloat64 `json:"fee_xrp,omitempty" db:"fee_xrp"`
	Status             string          `json:"status" db:"status"`
	QRCode             sql.NullString  `json:"qr_code,omitempty" db:"qr_code"`
	ExpiresAt          sql.NullTime    `json:"expires_at,omitempty" db:"expires_at"`
	ConfirmedAt        sql.NullTime    `json:"confirmed_at,omitempty" db:"confirmed_at"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

// ExchangeRate represents an exchange rate record.
type ExchangeRate struct {
	ID         int64           `json:"id" db:"id"`
	Source     string          `json:"source" db:"source"`
	XRPUSDRate float64         `json:"xrp_usd_rate" db:"xrp_usd_rate"`
	Bid        sql.NullFloat64 `json:"bid,omitempty" db:"bid"`
	Ask        sql.NullFloat64 `json:"ask,omitempty" db:"ask"`
	Volume24h  sql.NullFloat64 `json:"volume_24h,omitempty" db:"volume_24h"`
	Timestamp  time.Time       `json:"timestamp" db:"timestamp"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// XRPLTransactionLog represents a transaction log entry.
type XRPLTransactionLog struct {
	ID              int64          `json:"id" db:"id"`
	PaymentID       sql.NullInt64  `json:"payment_id,omitempty" db:"payment_id"`
	TxHash          string         `json:"tx_hash" db:"tx_hash"`
	TxType          string         `json:"tx_type" db:"tx_type"`
	FromAddress     sql.NullString `json:"from_address,omitempty" db:"from_address"`
	ToAddress       sql.NullString `json:"to_address,omitempty" db:"to_address"`
	AmountDrops     sql.NullInt64  `json:"amount_drops,omitempty" db:"amount_drops"`
	DestinationTag  sql.NullInt32  `json:"destination_tag,omitempty" db:"destination_tag"`
	LedgerIndex     sql.NullInt64  `json:"ledger_index,omitempty" db:"ledger_index"`
	LedgerHash      sql.NullString `json:"ledger_hash,omitempty" db:"ledger_hash"`
	TransactionDate sql.NullTime   `json:"transaction_date,omitempty" db:"transaction_date"`
	RawTransaction  []byte         `json:"raw_transaction" db:"raw_transaction"`
	Notes           sql.NullString `json:"notes,omitempty" db:"notes"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
}

// PaymentWithXRPL represents a joined payment and XRPL payment record.
type PaymentWithXRPL struct {
	Payment
	XRPLPayment *XRPLPayment `json:"xrpl_payment,omitempty"`
}

// PaymentStatus constants.
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusCompleted  = "completed"
	PaymentStatusFailed     = "failed"
	PaymentStatusRefunded   = "refunded"
	PaymentStatusExpired    = "expired"
)

// XRPLPaymentStatus constants.
const (
	XRPLPaymentStatusPending   = "pending"
	XRPLPaymentStatusConfirmed = "confirmed"
	XRPLPaymentStatusExpired   = "expired"
	XRPLPaymentStatusFailed    = "failed"
	XRPLPaymentStatusRefunded  = "refunded"
)

// PaymentMethod constants.
const (
	PaymentMethodXRPL       = "xrpl"
	PaymentMethodACH        = "ach"
	PaymentMethodCreditCard = "credit_card"
	PaymentMethodWire       = "wire"
)

// ExchangeRateSource constants.
const (
	ExchangeSourceCoinbase   = "coinbase"
	ExchangeSourceBinance    = "binance"
	ExchangeSourceKraken     = "kraken"
	ExchangeSourceBitstamp   = "bitstamp"
	ExchangeSourceAggregated = "aggregated"
)
