package xrpl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MonitorService monitors XRPL payments in real-time.
type MonitorService struct {
	xrplClient *Client
	logger     *zap.Logger
	monitors   map[string]*paymentMonitor
	mu         sync.RWMutex
	stopChan   chan struct{}
}

// paymentMonitor tracks a single payment.
type paymentMonitor struct {
	paymentID      string
	destinationTag uint32
	expectedAmount int64 // in drops
	expiresAt      time.Time
	ctx            context.Context
	cancel         context.CancelFunc
	callback       PaymentCallback
}

// PaymentCallback is called when payment status changes.
type PaymentCallback func(ctx context.Context, result *PaymentResult) error

// PaymentResult represents the result of payment monitoring.
type PaymentResult struct {
	PaymentID    string
	Status       PaymentStatus
	TxHash       string
	LedgerIndex  int64
	ActualAmount int64 // in drops
	ConfirmedAt  time.Time
	Error        error
}

// PaymentStatus represents payment status.
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	PaymentStatusExpired   PaymentStatus = "expired"
	PaymentStatusFailed    PaymentStatus = "failed"
)

// NewMonitorService creates a new payment monitor service.
func NewMonitorService(xrplClient *Client, logger *zap.Logger) *MonitorService {
	return &MonitorService{
		xrplClient: xrplClient,
		logger:     logger.With(zap.String("component", "payment-monitor")),
		monitors:   make(map[string]*paymentMonitor),
		stopChan:   make(chan struct{}),
	}
}

// Start starts the payment monitor service.
func (s *MonitorService) Start(ctx context.Context) error {
	s.logger.Info("starting payment monitor service")

	// Subscribe to state wallet transactions
	err := s.xrplClient.SubscribeToAccount(ctx, s.xrplClient.stateAddress, s.handleTransaction)
	if err != nil {
		return fmt.Errorf("failed to subscribe to account: %w", err)
	}

	// Start expiration checker
	go s.checkExpirations(ctx)

	return nil
}

// Stop stops the payment monitor service.
func (s *MonitorService) Stop(ctx context.Context) error {
	s.logger.Info("stopping payment monitor service")

	close(s.stopChan)

	// Cancel all active monitors
	s.mu.Lock()
	for _, monitor := range s.monitors {
		monitor.cancel()
	}
	s.monitors = make(map[string]*paymentMonitor)
	s.mu.Unlock()

	// Unsubscribe from account
	err := s.xrplClient.UnsubscribeFromAccount(ctx, s.xrplClient.stateAddress)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from account: %w", err)
	}

	return nil
}

// MonitorPayment starts monitoring a payment.
func (s *MonitorService) MonitorPayment(
	ctx context.Context,
	paymentID string,
	destinationTag uint32,
	expectedAmount int64,
	expiresAt time.Time,
	callback PaymentCallback,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already monitoring
	if _, exists := s.monitors[paymentID]; exists {
		return fmt.Errorf("already monitoring payment: %s", paymentID)
	}

	// Create cancellable context with timeout
	timeout := time.Until(expiresAt)
	if timeout < 0 {
		return fmt.Errorf("payment already expired")
	}

	monitorCtx, cancel := context.WithTimeout(ctx, timeout)

	monitor := &paymentMonitor{
		paymentID:      paymentID,
		destinationTag: destinationTag,
		expectedAmount: expectedAmount,
		expiresAt:      expiresAt,
		ctx:            monitorCtx,
		cancel:         cancel,
		callback:       callback,
	}

	s.monitors[paymentID] = monitor

	s.logger.Info("started monitoring payment",
		zap.String("payment_id", paymentID),
		zap.Uint32("destination_tag", destinationTag),
		zap.Int64("expected_amount", expectedAmount),
		zap.Time("expires_at", expiresAt),
	)

	return nil
}

// StopMonitoring stops monitoring a payment.
func (s *MonitorService) StopMonitoring(paymentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if monitor, exists := s.monitors[paymentID]; exists {
		monitor.cancel()
		delete(s.monitors, paymentID)

		s.logger.Info("stopped monitoring payment",
			zap.String("payment_id", paymentID),
		)
	}
}

// handleTransaction processes incoming transactions.
func (s *MonitorService) handleTransaction(tx *Transaction) error {
	// Only process Payment transactions
	if tx.TransactionType != "Payment" {
		return nil
	}

	// Check if transaction is successful
	if tx.TransactionResult != "tesSUCCESS" {
		s.logger.Warn("transaction failed",
			zap.String("tx_hash", tx.Hash),
			zap.String("result", tx.TransactionResult),
		)
		return nil
	}

	s.logger.Info("incoming payment detected",
		zap.String("tx_hash", tx.Hash),
		zap.String("from", tx.Account),
		zap.Int64("amount", tx.Amount),
		zap.Any("destination_tag", tx.DestinationTag),
	)

	// Find matching payment monitor
	s.mu.RLock()
	var matchingMonitor *paymentMonitor
	for _, monitor := range s.monitors {
		if s.isMatchingPayment(tx, monitor) {
			matchingMonitor = monitor
			break
		}
	}
	s.mu.RUnlock()

	if matchingMonitor == nil {
		s.logger.Warn("no matching payment found",
			zap.String("tx_hash", tx.Hash),
			zap.Any("destination_tag", tx.DestinationTag),
		)
		return nil
	}

	// Process matched payment
	return s.processMatchedPayment(tx, matchingMonitor)
}

// isMatchingPayment checks if transaction matches a payment monitor.
func (s *MonitorService) isMatchingPayment(tx *Transaction, monitor *paymentMonitor) bool {
	// Check destination tag
	if tx.DestinationTag == nil || *tx.DestinationTag != monitor.destinationTag {
		return false
	}

	// Check amount (1% tolerance)
	tolerance := monitor.expectedAmount / 100
	if tx.Amount < monitor.expectedAmount-tolerance || tx.Amount > monitor.expectedAmount+tolerance {
		s.logger.Warn("amount mismatch",
			zap.String("payment_id", monitor.paymentID),
			zap.Int64("expected", monitor.expectedAmount),
			zap.Int64("received", tx.Amount),
		)
		return false
	}

	return true
}

// processMatchedPayment handles a matched payment.
func (s *MonitorService) processMatchedPayment(tx *Transaction, monitor *paymentMonitor) error {
	s.logger.Info("payment matched",
		zap.String("payment_id", monitor.paymentID),
		zap.String("tx_hash", tx.Hash),
	)

	result := &PaymentResult{
		PaymentID:    monitor.paymentID,
		Status:       PaymentStatusConfirmed,
		TxHash:       tx.Hash,
		LedgerIndex:  tx.LedgerIndex,
		ActualAmount: tx.Amount,
		ConfirmedAt:  tx.Date,
	}

	// Call callback
	if monitor.callback != nil {
		if err := monitor.callback(monitor.ctx, result); err != nil {
			s.logger.Error("payment callback failed",
				zap.String("payment_id", monitor.paymentID),
				zap.Error(err),
			)
			return err
		}
	}

	// Stop monitoring this payment
	s.StopMonitoring(monitor.paymentID)

	return nil
}

// checkExpirations periodically checks for expired payments.
func (s *MonitorService) checkExpirations(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processExpiredPayments()
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// processExpiredPayments handles expired payments.
func (s *MonitorService) processExpiredPayments() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for paymentID, monitor := range s.monitors {
		if now.After(monitor.expiresAt) {
			s.logger.Info("payment expired",
				zap.String("payment_id", paymentID),
				zap.Time("expired_at", monitor.expiresAt),
			)

			// Call callback with expired status
			if monitor.callback != nil {
				result := &PaymentResult{
					PaymentID: paymentID,
					Status:    PaymentStatusExpired,
					Error:     fmt.Errorf("payment expired"),
				}
				_ = monitor.callback(monitor.ctx, result)
			}

			// Clean up
			monitor.cancel()
			delete(s.monitors, paymentID)
		}
	}
}

// GetMonitoringStatus returns the current monitoring status.
func (s *MonitorService) GetMonitoringStatus() map[string]MonitorStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]MonitorStatus)
	for paymentID, monitor := range s.monitors {
		status[paymentID] = MonitorStatus{
			PaymentID:      paymentID,
			DestinationTag: monitor.destinationTag,
			ExpiresAt:      monitor.expiresAt,
			TimeRemaining:  time.Until(monitor.expiresAt),
		}
	}

	return status
}

// MonitorStatus represents the status of a monitored payment.
type MonitorStatus struct {
	PaymentID      string
	DestinationTag uint32
	ExpiresAt      time.Time
	TimeRemaining  time.Duration
}
