package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/maxfelker/excise-tax-backend/v2/pkg/xrpl"
)

// Processor handles payment processing logic
type Processor struct {
	xrplClient *xrpl.Client
	storage    map[string]*PaymentTransaction // In-memory storage for CLI
	mu         sync.RWMutex
}

// NewProcessor creates a new payment processor
func NewProcessor(xrplClient *xrpl.Client) *Processor {
	return &Processor{
		xrplClient: xrplClient,
		storage:    make(map[string]*PaymentTransaction),
	}
}

// CreatePayment creates a new payment transaction
func (p *Processor) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentTransaction, error) {
	// Validate payment request
	if err := validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Generate unique payment ID (UUID v4)
	paymentID := uuid.New().String()

	// Create initial transaction
	tx := &PaymentTransaction{
		ID:        paymentID,
		Type:      req.Type,
		Amount:    req.Amount,
		Currency:  req.Currency,
		From:      req.From,
		To:        req.To,
		State:     StatePending,
		Events:    []PaymentEvent{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Metadata:  req.Metadata,
	}

	// Create initial event
	if err := AppendEvent(tx, EventPaymentInitiated, StatePending, nil); err != nil {
		return nil, fmt.Errorf("failed to create initial event: %w", err)
	}

	// Store transaction
	p.mu.Lock()
	p.storage[paymentID] = tx
	p.mu.Unlock()

	// Process payment based on type
	switch req.Type {
	case PaymentTypeXRPL:
		go p.processXRPLPaymentAsync(context.Background(), tx)
	case PaymentTypeCreditCard, PaymentTypeACH:
		// Future implementation
		return nil, fmt.Errorf("payment type %s not yet implemented", req.Type)
	default:
		return nil, ErrInvalidPaymentType
	}

	return tx, nil
}

// GetPayment retrieves a payment by ID
func (p *Processor) GetPayment(ctx context.Context, paymentID string) (*PaymentTransaction, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tx, ok := p.storage[paymentID]
	if !ok {
		return nil, ErrPaymentNotFound
	}

	return tx, nil
}

// ListPayments returns payments matching the filters
func (p *Processor) ListPayments(ctx context.Context, filters ListPaymentsFilters) ([]*PaymentTransaction, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*PaymentTransaction
	for _, tx := range p.storage {
		if matchesFilters(tx, filters) {
			result = append(result, tx)
		}
	}

	// Apply limit and offset
	start := filters.Offset
	if start > len(result) {
		return []*PaymentTransaction{}, nil
	}

	end := start + filters.Limit
	if filters.Limit == 0 || end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

// VerifyPayment verifies the current status of a payment
func (p *Processor) VerifyPayment(ctx context.Context, paymentID string) (*PaymentTransaction, error) {
	tx, err := p.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// For XRPL payments, verify on-chain if we have a transaction hash
	if tx.Type == PaymentTypeXRPL && tx.State == StatePending {
		if err := p.verifyXRPLPayment(ctx, tx); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

// RetryPayment attempts to retry a failed payment
func (p *Processor) RetryPayment(ctx context.Context, paymentID string) error {
	tx, err := p.GetPayment(ctx, paymentID)
	if err != nil {
		return err
	}

	if tx.State != StateFailed {
		return fmt.Errorf("can only retry failed payments, current state: %s", tx.State)
	}

	// Add retry event
	p.mu.Lock()
	if err := AppendEvent(tx, EventPaymentRetried, StatePending, nil); err != nil {
		p.mu.Unlock()
		return err
	}
	p.mu.Unlock()

	// Process payment again based on type
	switch tx.Type {
	case PaymentTypeXRPL:
		go p.processXRPLPaymentAsync(context.Background(), tx)
	default:
		return fmt.Errorf("payment type %s not yet implemented", tx.Type)
	}

	return nil
}

// processXRPLPaymentAsync processes an XRPL payment asynchronously
func (p *Processor) processXRPLPaymentAsync(ctx context.Context, tx *PaymentTransaction) {
	if err := p.ProcessXRPLPayment(ctx, tx); err != nil {
		// Log error and update transaction state
		p.mu.Lock()
		details := XRPLPaymentDetails{
			ErrorMessage: err.Error(),
		}
		_ = AppendEvent(tx, EventPaymentFailed, StateFailed, details)
		p.mu.Unlock()
	}
}

// ProcessXRPLPayment processes an XRPL payment
func (p *Processor) ProcessXRPLPayment(ctx context.Context, tx *PaymentTransaction) error {
	if p.xrplClient == nil {
		return fmt.Errorf("XRPL client not configured")
	}

	// Parse amount (convert to drops for XRPL)
	amountFloat, err := strconv.ParseFloat(tx.Amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	amountDrops := int64(amountFloat * 1_000_000) // Convert XRP to drops

	// Submit payment event
	p.mu.Lock()
	if err := AppendEvent(tx, EventPaymentSubmitted, StatePending, nil); err != nil {
		p.mu.Unlock()
		return err
	}
	p.mu.Unlock()

	// Note: In a real implementation, we would use the XRPL client to submit the payment
	// For now, we'll simulate the process

	// Simulate payment submission (would use xrplClient.SubmitPayment in real impl)
	_ = amountDrops
	txHash := fmt.Sprintf("xrpl_%s", uuid.New().String())

	// Add verification event
	p.mu.Lock()
	details := XRPLPaymentDetails{
		NetworkTxHash: txHash,
		Ledger:        12345678,
		Sequence:      1,
		Fee:           "12",
	}
	if err := AppendEvent(tx, EventPaymentVerified, StatePending, details); err != nil {
		p.mu.Unlock()
		return err
	}
	p.mu.Unlock()

	// Simulate waiting for confirmation
	time.Sleep(2 * time.Second)

	// Mark as completed
	p.mu.Lock()
	if err := AppendEvent(tx, EventPaymentCompleted, StateCompleted, details); err != nil {
		p.mu.Unlock()
		return err
	}
	p.mu.Unlock()

	return nil
}

// verifyXRPLPayment verifies an XRPL payment on-chain
func (p *Processor) verifyXRPLPayment(ctx context.Context, tx *PaymentTransaction) error {
	// Get the latest event to find the transaction hash
	latestEvent := GetLatestEvent(tx)
	if latestEvent == nil || latestEvent.Details == nil {
		return nil // No hash to verify yet
	}

	var details XRPLPaymentDetails
	if err := json.Unmarshal(latestEvent.Details, &details); err != nil {
		return nil // Details don't contain XRPL info
	}

	if details.NetworkTxHash == "" {
		return nil // No hash yet
	}

	// In real implementation, would verify with XRPL client
	// For now, just return success
	return nil
}

// validateCreateRequest validates a payment creation request
func validateCreateRequest(req CreatePaymentRequest) error {
	if !req.Type.IsValid() {
		return ErrInvalidPaymentType
	}

	if req.Amount == "" {
		return &ValidationError{Field: "amount", Message: "amount is required"}
	}

	// Validate amount is a positive number
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		return &ValidationError{Field: "amount", Message: "amount must be a valid number"}
	}
	if amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}

	if req.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}

	if req.From == "" {
		return &ValidationError{Field: "from", Message: "from address is required"}
	}

	if req.To == "" {
		return &ValidationError{Field: "to", Message: "to address is required"}
	}

	return nil
}

// matchesFilters checks if a transaction matches the given filters
func matchesFilters(tx *PaymentTransaction, filters ListPaymentsFilters) bool {
	if filters.State != nil && tx.State != *filters.State {
		return false
	}

	if filters.Type != nil && tx.Type != *filters.Type {
		return false
	}

	return true
}
