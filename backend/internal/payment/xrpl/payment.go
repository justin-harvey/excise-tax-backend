package xrpl

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
)

// PaymentProcessor handles XRPL payment operations.
type PaymentProcessor struct {
	client  *Client
	oracle  *PriceOracle
	logger  *zap.Logger
	monitor *MonitorService
}

// NewPaymentProcessor creates a new payment processor.
func NewPaymentProcessor(client *Client, oracle *PriceOracle, monitor *MonitorService, logger *zap.Logger) *PaymentProcessor {
	return &PaymentProcessor{
		client:  client,
		oracle:  oracle,
		monitor: monitor,
		logger:  logger.With(zap.String("component", "payment-processor")),
	}
}

// GenerateDestinationTag generates a unique destination tag from manufacturer ID.
// Range: 1,000,000 - 9,999,999
func GenerateDestinationTag(manufacturerID int64) (uint32, error) {
	const minTag uint32 = 1000000
	const maxTag uint32 = 9999999

	tag := minTag + uint32(manufacturerID)

	if tag > maxTag {
		return 0, fmt.Errorf("manufacturer ID %d exceeds destination tag range", manufacturerID)
	}

	return tag, nil
}

// CreatePaymentRequest creates a new XRPL payment request.
func (p *PaymentProcessor) CreatePaymentRequest(ctx context.Context, req *CreatePaymentRequest) (*PaymentResponse, error) {
	p.logger.Info("creating payment request",
		zap.Int64("manufacturer_id", req.ManufacturerID),
		zap.Float64("amount_usd", req.AmountUSD),
		zap.String("report_id", req.ReportID),
	)

	// Get current exchange rate
	rate, err := p.oracle.GetExchangeRate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Convert USD to XRP with volatility buffer (2%)
	xrpAmount, err := p.oracle.ConvertUSDtoXRP(ctx, req.AmountUSD, true)
	if err != nil {
		return nil, fmt.Errorf("failed to convert USD to XRP: %w", err)
	}

	// Generate destination tag
	destinationTag, err := GenerateDestinationTag(req.ManufacturerID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate destination tag: %w", err)
	}

	// Generate payment ID
	paymentID := GeneratePaymentID()

	// Calculate expiration time (30 minutes)
	expiresAt := time.Now().Add(30 * time.Minute)

	// Create payment response
	response := &PaymentResponse{
		PaymentID:          paymentID,
		ManufacturerID:     req.ManufacturerID,
		ReportID:           req.ReportID,
		AmountUSD:          req.AmountUSD,
		XRPAmount:          xrpAmount,
		ExchangeRate:       rate.XRPUSD,
		DestinationAddress: p.client.stateAddress,
		DestinationTag:     destinationTag,
		Status:             string(PaymentStatusPending),
		CreatedAt:          time.Now(),
		ExpiresAt:          expiresAt,
		Instructions: PaymentInstructions{
			Destination:    p.client.stateAddress,
			Amount:         xrpAmount,
			DestinationTag: destinationTag,
			Currency:       "XRP",
			Network:        p.client.network,
		},
	}

	// Generate QR code
	qrCode, err := p.GenerateQRCode(response)
	if err != nil {
		p.logger.Warn("failed to generate QR code", zap.Error(err))
		// Don't fail the request if QR code generation fails
	} else {
		response.QRCode = qrCode
	}

	p.logger.Info("payment request created",
		zap.String("payment_id", paymentID),
		zap.Float64("xrp_amount", xrpAmount),
		zap.Uint32("destination_tag", destinationTag),
	)

	return response, nil
}

// GenerateQRCode generates a QR code for the payment.
// Format: xrpl:{address}?dt={tag}&amount={xrp}
func (p *PaymentProcessor) GenerateQRCode(payment *PaymentResponse) (string, error) {
	// Create payment URI
	uri := fmt.Sprintf("xrpl:%s?dt=%d&amount=%.6f",
		payment.DestinationAddress,
		payment.DestinationTag,
		payment.XRPAmount,
	)

	// Generate QR code
	qr, err := qrcode.New(uri, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to create QR code: %w", err)
	}

	// Encode as PNG
	png, err := qr.PNG(256)
	if err != nil {
		return "", fmt.Errorf("failed to encode QR code: %w", err)
	}

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(png)

	return encoded, nil
}

// VerifyTransaction verifies a transaction on the blockchain.
func (p *PaymentProcessor) VerifyTransaction(ctx context.Context, txHash string) (*TransactionVerification, error) {
	p.logger.Info("verifying transaction",
		zap.String("tx_hash", txHash),
	)

	verification, err := p.client.VerifyTransaction(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify transaction: %w", err)
	}

	p.logger.Info("transaction verified",
		zap.String("tx_hash", txHash),
		zap.Bool("validated", verification.Validated),
		zap.Bool("success", verification.Success),
	)

	return verification, nil
}

// CreatePaymentRequest represents a request to create a payment.
type CreatePaymentRequest struct {
	ManufacturerID    int64
	ReportID          string
	AmountUSD         float64
	Description       string
	ManufacturerEmail string
}

// PaymentResponse represents a payment request response.
type PaymentResponse struct {
	PaymentID          string
	ManufacturerID     int64
	ReportID           string
	AmountUSD          float64
	XRPAmount          float64
	ExchangeRate       float64
	DestinationAddress string
	DestinationTag     uint32
	Status             string
	QRCode             string
	CreatedAt          time.Time
	ExpiresAt          time.Time
	Instructions       PaymentInstructions
}

// PaymentInstructions contains payment instructions for the manufacturer.
type PaymentInstructions struct {
	Destination    string
	Amount         float64
	DestinationTag uint32
	Currency       string
	Network        string
}

// VerifyPaymentRequest represents a request to verify a payment.
type VerifyPaymentRequest struct {
	PaymentID string
	TxHash    string
}

// VerifyPaymentResponse represents a payment verification response.
type VerifyPaymentResponse struct {
	PaymentID   string
	TxHash      string
	Validated   bool
	Success     bool
	Amount      float64
	Status      string
	ConfirmedAt time.Time
}
