package payment

import (
	"encoding/json"
	"time"
)

// PaymentType represents the type of payment method
type PaymentType string

const (
	PaymentTypeCreditCard PaymentType = "credit_card"
	PaymentTypeACH        PaymentType = "ach"
	PaymentTypeXRPL       PaymentType = "xrpl"
)

// TransactionState represents the current state of a payment transaction
type TransactionState string

const (
	StatePending   TransactionState = "pending"
	StateCompleted TransactionState = "completed"
	StateInReview  TransactionState = "in_review"
	StateFailed    TransactionState = "failed"
)

// EventType represents the type of payment event
type EventType string

const (
	EventPaymentInitiated EventType = "payment_initiated"
	EventPaymentSubmitted EventType = "payment_submitted"
	EventPaymentVerified  EventType = "payment_verified"
	EventPaymentRetried   EventType = "payment_retried"
	EventPaymentCompleted EventType = "payment_completed"
	EventPaymentFailed    EventType = "payment_failed"
	EventPaymentReviewed  EventType = "payment_reviewed"
)

// PaymentTransaction represents a complete payment transaction with its event history
type PaymentTransaction struct {
	ID        string           `json:"id"`
	Type      PaymentType      `json:"type"`
	Amount    string           `json:"amount"`
	Currency  string           `json:"currency"`
	From      string           `json:"from"`
	To        string           `json:"to"`
	State     TransactionState `json:"state"`
	Events    []PaymentEvent   `json:"events"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Metadata  json.RawMessage  `json:"metadata,omitempty"`
}

// PaymentEvent represents an immutable event in a payment's lifecycle
type PaymentEvent struct {
	ID            string           `json:"id"`
	TransactionID string           `json:"transaction_id"`
	Type          EventType        `json:"type"`
	State         TransactionState `json:"state"`
	Details       json.RawMessage  `json:"details"`
	Timestamp     time.Time        `json:"timestamp"`
}

// XRPLPaymentDetails contains XRPL-specific payment information
type XRPLPaymentDetails struct {
	NetworkTxHash string `json:"network_tx_hash,omitempty"`
	Ledger        int64  `json:"ledger,omitempty"`
	Sequence      int64  `json:"sequence,omitempty"`
	Fee           string `json:"fee,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// CreatePaymentRequest represents a request to create a new payment
type CreatePaymentRequest struct {
	Type     PaymentType     `json:"type"`
	Amount   string          `json:"amount"`
	Currency string          `json:"currency"`
	From     string          `json:"from"`
	To       string          `json:"to"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// ListPaymentsFilters contains filters for listing payments
type ListPaymentsFilters struct {
	State  *TransactionState
	Type   *PaymentType
	Limit  int
	Offset int
}

// ValidPaymentTypes returns all valid payment types
func ValidPaymentTypes() []PaymentType {
	return []PaymentType{PaymentTypeCreditCard, PaymentTypeACH, PaymentTypeXRPL}
}

// ValidTransactionStates returns all valid transaction states
func ValidTransactionStates() []TransactionState {
	return []TransactionState{StatePending, StateCompleted, StateInReview, StateFailed}
}

// IsValid checks if a PaymentType is valid
func (pt PaymentType) IsValid() bool {
	for _, valid := range ValidPaymentTypes() {
		if pt == valid {
			return true
		}
	}
	return false
}

// IsValid checks if a TransactionState is valid
func (ts TransactionState) IsValid() bool {
	for _, valid := range ValidTransactionStates() {
		if ts == valid {
			return true
		}
	}
	return false
}
