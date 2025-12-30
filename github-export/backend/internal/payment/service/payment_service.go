package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"excise-tax-portal/backend/internal/payment/model"
	"excise-tax-portal/backend/internal/payment/repository"
	"excise-tax-portal/backend/internal/payment/xrpl"
	"excise-tax-portal/backend/pkg/cache"

	"go.uber.org/zap"
)

// PaymentService handles payment business logic.
type PaymentService struct {
	repo       *repository.PaymentRepository
	cache      *cache.RedisClient
	xrplClient *xrpl.Client
	oracle     *xrpl.PriceOracle
	monitor    *xrpl.MonitorService
	processor  *xrpl.PaymentProcessor
	logger     *zap.Logger
}

// NewPaymentService creates a new payment service.
func NewPaymentService(
	repo *repository.PaymentRepository,
	cache *cache.RedisClient,
	xrplClient *xrpl.Client,
	oracle *xrpl.PriceOracle,
	monitor *xrpl.MonitorService,
	processor *xrpl.PaymentProcessor,
	logger *zap.Logger,
) *PaymentService {
	return &PaymentService{
		repo:       repo,
		cache:      cache,
		xrplClient: xrplClient,
		oracle:     oracle,
		monitor:    monitor,
		processor:  processor,
		logger:     logger.With(zap.String("component", "payment-service")),
	}
}

// CreateXRPLPayment creates a new XRPL payment request.
func (s *PaymentService) CreateXRPLPayment(ctx context.Context, req *CreateXRPLPaymentRequest) (*XRPLPaymentResponse, error) {
	s.logger.Info("creating XRPL payment",
		zap.Int64("manufacturer_id", req.ManufacturerID),
		zap.Float64("amount_usd", req.AmountUSD),
	)

	// Create payment request
	paymentReq := &xrpl.CreatePaymentRequest{
		ManufacturerID:    req.ManufacturerID,
		ReportID:          req.ReportID,
		AmountUSD:         req.AmountUSD,
		Description:       req.Description,
		ManufacturerEmail: req.ManufacturerEmail,
	}

	paymentResp, err := s.processor.CreatePaymentRequest(ctx, paymentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment request: %w", err)
	}

	// Save to database
	payment := &model.Payment{
		ManufacturerID: req.ManufacturerID,
		ReportID:       sql.NullInt64{Valid: req.ReportID != ""},
		PaymentMethod:  model.PaymentMethodXRPL,
		AmountUSD:      req.AmountUSD,
		Status:         model.PaymentStatusPending,
	}

	// Set report ID if provided
	if req.ReportID != "" {
		var reportID int64
		if _, err := fmt.Sscanf(req.ReportID, "%d", &reportID); err == nil {
			payment.ReportID = sql.NullInt64{Int64: reportID, Valid: true}
		}
	}

	xrplPayment := &model.XRPLPayment{
		XRPAmount:          paymentResp.XRPAmount,
		ExchangeRate:       paymentResp.ExchangeRate,
		DestinationAddress: paymentResp.DestinationAddress,
		DestinationTag:     sql.NullInt32{Int32: int32(paymentResp.DestinationTag), Valid: true},
		Status:             model.XRPLPaymentStatusPending,
		QRCode:             sql.NullString{String: paymentResp.QRCode, Valid: paymentResp.QRCode != ""},
		ExpiresAt:          sql.NullTime{Time: paymentResp.ExpiresAt, Valid: true},
	}

	if err := s.repo.CreatePayment(ctx, payment, xrplPayment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Start monitoring payment
	expectedAmount := xrpl.XRPToDrops(paymentResp.XRPAmount)
	err = s.monitor.MonitorPayment(
		ctx,
		paymentResp.PaymentID,
		paymentResp.DestinationTag,
		expectedAmount,
		paymentResp.ExpiresAt,
		s.createPaymentCallback(payment.ID),
	)
	if err != nil {
		s.logger.Error("failed to start payment monitoring", zap.Error(err))
		// Don't fail the request, monitoring can be retried
	}

	// Save exchange rate
	s.saveExchangeRate(ctx, paymentResp.ExchangeRate)

	response := &XRPLPaymentResponse{
		PaymentID:          payment.ID,
		XRPLPaymentID:      xrplPayment.ID.Int64,
		ManufacturerID:     payment.ManufacturerID,
		AmountUSD:          payment.AmountUSD,
		XRPAmount:          xrplPayment.XRPAmount,
		ExchangeRate:       xrplPayment.ExchangeRate,
		DestinationAddress: xrplPayment.DestinationAddress,
		DestinationTag:     uint32(xrplPayment.DestinationTag.Int32),
		QRCode:             xrplPayment.QRCode.String,
		Status:             xrplPayment.Status,
		ExpiresAt:          xrplPayment.ExpiresAt.Time,
		CreatedAt:          payment.CreatedAt,
		Instructions:       paymentResp.Instructions,
	}

	s.logger.Info("XRPL payment created",
		zap.Int64("payment_id", payment.ID),
		zap.Uint32("destination_tag", uint32(xrplPayment.DestinationTag.Int32)),
	)

	return response, nil
}

// GetPayment retrieves a payment by ID.
func (s *PaymentService) GetPayment(ctx context.Context, id int64) (*model.PaymentWithXRPL, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("payment:%d", id)
	var cached model.PaymentWithXRPL
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	// Get from database
	payment, err := s.repo.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	_ = s.cache.Set(ctx, cacheKey, payment, 5*time.Minute)

	return payment, nil
}

// VerifyPayment verifies a payment transaction on the blockchain.
func (s *PaymentService) VerifyPayment(ctx context.Context, paymentID int64, txHash string) (*VerifyPaymentResponse, error) {
	s.logger.Info("verifying payment",
		zap.Int64("payment_id", paymentID),
		zap.String("tx_hash", txHash),
	)

	// Get payment
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.PaymentMethod != model.PaymentMethodXRPL {
		return nil, fmt.Errorf("payment is not an XRPL payment")
	}

	// Verify transaction on blockchain
	verification, err := s.processor.VerifyTransaction(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify transaction: %w", err)
	}

	if !verification.Validated || !verification.Success {
		return &VerifyPaymentResponse{
			PaymentID: paymentID,
			TxHash:    txHash,
			Verified:  false,
			Message:   "Transaction not validated or failed",
		}, nil
	}

	// Check if transaction matches payment
	if verification.DestinationTag == nil || uint32(*verification.DestinationTag) != uint32(payment.XRPLPayment.DestinationTag.Int32) {
		return &VerifyPaymentResponse{
			PaymentID: paymentID,
			TxHash:    txHash,
			Verified:  false,
			Message:   "Destination tag mismatch",
		}, nil
	}

	// Update payment status
	confirmedAt := time.Now()
	err = s.repo.UpdateXRPLPaymentStatus(ctx, paymentID, model.XRPLPaymentStatusConfirmed, &txHash, &verification.LedgerIndex, &confirmedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	err = s.repo.UpdatePaymentStatus(ctx, paymentID, model.PaymentStatusCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("payment:%d", paymentID)
	_ = s.cache.Delete(ctx, cacheKey)

	// Log transaction
	s.logTransaction(ctx, payment, verification)

	s.logger.Info("payment verified",
		zap.Int64("payment_id", paymentID),
		zap.String("tx_hash", txHash),
	)

	return &VerifyPaymentResponse{
		PaymentID:   paymentID,
		TxHash:      txHash,
		Verified:    true,
		ConfirmedAt: confirmedAt,
		Message:     "Payment verified successfully",
	}, nil
}

// GetExchangeRate retrieves the current exchange rate.
func (s *PaymentService) GetExchangeRate(ctx context.Context) (*ExchangeRateResponse, error) {
	rate, err := s.oracle.GetExchangeRate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	sources := make([]ExchangeSourceInfo, len(rate.Sources))
	for i, src := range rate.Sources {
		sources[i] = ExchangeSourceInfo{
			Name:      src.Source,
			Rate:      src.Rate,
			Timestamp: src.LastUpdated,
		}
	}

	return &ExchangeRateResponse{
		XRPUSD:     rate.XRPUSD,
		USDXRP:     rate.USDXRP,
		Sources:    sources,
		UpdatedAt:  rate.Updated,
	}, nil
}

// ConvertCurrency converts between USD and XRP.
func (s *PaymentService) ConvertCurrency(ctx context.Context, amount float64, from, to string) (*ConvertCurrencyResponse, error) {
	var result float64
	var err error

	if from == "USD" && to == "XRP" {
		result, err = s.oracle.ConvertUSDtoXRP(ctx, amount, false)
	} else if from == "XRP" && to == "USD" {
		result, err = s.oracle.ConvertXRPtoUSD(ctx, amount)
	} else {
		return nil, fmt.Errorf("unsupported currency conversion: %s to %s", from, to)
	}

	if err != nil {
		return nil, err
	}

	rate, _ := s.oracle.GetExchangeRate(ctx)

	return &ConvertCurrencyResponse{
		Amount:       amount,
		FromCurrency: from,
		ToCurrency:   to,
		Result:       result,
		Rate:         rate.XRPUSD,
		Timestamp:    time.Now(),
	}, nil
}

// GetPaymentsByManufacturer retrieves payments for a manufacturer.
func (s *PaymentService) GetPaymentsByManufacturer(ctx context.Context, manufacturerID int64, limit, offset int) ([]*model.PaymentWithXRPL, error) {
	return s.repo.GetPaymentsByManufacturer(ctx, manufacturerID, limit, offset)
}

// createPaymentCallback creates a callback for payment monitoring.
func (s *PaymentService) createPaymentCallback(paymentID int64) xrpl.PaymentCallback {
	return func(ctx context.Context, result *xrpl.PaymentResult) error {
		s.logger.Info("payment callback triggered",
			zap.Int64("payment_id", paymentID),
			zap.String("status", string(result.Status)),
		)

		switch result.Status {
		case xrpl.PaymentStatusConfirmed:
			confirmedAt := result.ConfirmedAt
			err := s.repo.UpdateXRPLPaymentStatus(ctx, paymentID, model.XRPLPaymentStatusConfirmed, &result.TxHash, &result.LedgerIndex, &confirmedAt)
			if err != nil {
				return fmt.Errorf("failed to update payment status: %w", err)
			}

			err = s.repo.UpdatePaymentStatus(ctx, paymentID, model.PaymentStatusCompleted)
			if err != nil {
				return fmt.Errorf("failed to update payment status: %w", err)
			}

		case xrpl.PaymentStatusExpired:
			err := s.repo.UpdateXRPLPaymentStatus(ctx, paymentID, model.XRPLPaymentStatusExpired, nil, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to update payment status: %w", err)
			}

			err = s.repo.UpdatePaymentStatus(ctx, paymentID, model.PaymentStatusExpired)
			if err != nil {
				return fmt.Errorf("failed to update payment status: %w", err)
			}

		case xrpl.PaymentStatusFailed:
			err := s.repo.UpdateXRPLPaymentStatus(ctx, paymentID, model.XRPLPaymentStatusFailed, nil, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to update payment status: %w", err)
			}

			err = s.repo.UpdatePaymentStatus(ctx, paymentID, model.PaymentStatusFailed)
			if err != nil {
				return fmt.Errorf("failed to update payment status: %w", err)
			}
		}

		// Invalidate cache
		cacheKey := fmt.Sprintf("payment:%d", paymentID)
		_ = s.cache.Delete(ctx, cacheKey)

		return nil
	}
}

// saveExchangeRate saves the exchange rate to the database.
func (s *PaymentService) saveExchangeRate(ctx context.Context, rate float64) {
	exchangeRate := &model.ExchangeRate{
		Source:     model.ExchangeSourceAggregated,
		XRPUSDRate: rate,
		Timestamp:  time.Now(),
	}

	if err := s.repo.SaveExchangeRate(ctx, exchangeRate); err != nil {
		s.logger.Error("failed to save exchange rate", zap.Error(err))
	}
}

// logTransaction logs a transaction to the audit trail.
func (s *PaymentService) logTransaction(ctx context.Context, payment *model.PaymentWithXRPL, verification *xrpl.TransactionVerification) {
	rawTx, _ := json.Marshal(verification)

	log := &model.XRPLTransactionLog{
		PaymentID:       sql.NullInt64{Int64: payment.ID, Valid: true},
		TxHash:          verification.TxHash,
		TxType:          "Payment",
		ToAddress:       sql.NullString{String: verification.Destination, Valid: true},
		AmountDrops:     sql.NullInt64{Int64: verification.Amount, Valid: true},
		DestinationTag:  payment.XRPLPayment.DestinationTag,
		LedgerIndex:     sql.NullInt64{Int64: verification.LedgerIndex, Valid: true},
		RawTransaction:  rawTx,
		Notes:           sql.NullString{String: "Payment confirmed", Valid: true},
	}

	if err := s.repo.LogTransaction(ctx, log); err != nil {
		s.logger.Error("failed to log transaction", zap.Error(err))
	}
}

// Request/Response types

type CreateXRPLPaymentRequest struct {
	ManufacturerID    int64   `json:"manufacturer_id"`
	ReportID          string  `json:"report_id,omitempty"`
	AmountUSD         float64 `json:"amount_usd"`
	Description       string  `json:"description,omitempty"`
	ManufacturerEmail string  `json:"manufacturer_email,omitempty"`
}

type XRPLPaymentResponse struct {
	PaymentID          int64                   `json:"payment_id"`
	XRPLPaymentID      int64                   `json:"xrpl_payment_id"`
	ManufacturerID     int64                   `json:"manufacturer_id"`
	AmountUSD          float64                 `json:"amount_usd"`
	XRPAmount          float64                 `json:"xrp_amount"`
	ExchangeRate       float64                 `json:"exchange_rate"`
	DestinationAddress string                  `json:"destination_address"`
	DestinationTag     uint32                  `json:"destination_tag"`
	QRCode             string                  `json:"qr_code,omitempty"`
	Status             string                  `json:"status"`
	ExpiresAt          time.Time               `json:"expires_at"`
	CreatedAt          time.Time               `json:"created_at"`
	Instructions       xrpl.PaymentInstructions `json:"instructions"`
}

type VerifyPaymentResponse struct {
	PaymentID   int64     `json:"payment_id"`
	TxHash      string    `json:"tx_hash"`
	Verified    bool      `json:"verified"`
	ConfirmedAt time.Time `json:"confirmed_at,omitempty"`
	Message     string    `json:"message"`
}

type ExchangeRateResponse struct {
	XRPUSD    float64              `json:"xrp_usd"`
	USDXRP    float64              `json:"usd_xrp"`
	Sources   []ExchangeSourceInfo `json:"sources"`
	UpdatedAt time.Time            `json:"updated_at"`
}

type ExchangeSourceInfo struct {
	Name      string    `json:"name"`
	Rate      float64   `json:"rate"`
	Timestamp time.Time `json:"timestamp"`
}

type ConvertCurrencyResponse struct {
	Amount       float64   `json:"amount"`
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	Result       float64   `json:"result"`
	Rate         float64   `json:"rate"`
	Timestamp    time.Time `json:"timestamp"`
}
