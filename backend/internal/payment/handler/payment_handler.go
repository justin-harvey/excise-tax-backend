package handler

import (
	"net/http"
	"strconv"

	"excise-tax-portal/backend/internal/payment/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PaymentHandler handles payment HTTP requests.
type PaymentHandler struct {
	service *service.PaymentService
	logger  *zap.Logger
}

// NewPaymentHandler creates a new payment handler.
func NewPaymentHandler(service *service.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		service: service,
		logger:  logger.With(zap.String("component", "payment-handler")),
	}
}

// RegisterRoutes registers payment routes.
func (h *PaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		// XRPL payment routes
		xrpl := payments.Group("/xrpl")
		{
			xrpl.POST("/create", h.CreateXRPLPayment)
			xrpl.GET("/:paymentId/status", h.GetPaymentStatus)
			xrpl.POST("/:paymentId/verify", h.VerifyPayment)
		}

		// Exchange rate routes
		payments.GET("/exchange-rate", h.GetExchangeRate)
		payments.POST("/convert", h.ConvertCurrency)

		// Payment list
		payments.GET("", h.ListPayments)
		payments.GET("/:paymentId", h.GetPayment)
	}
}

// CreateXRPLPayment handles POST /payments/xrpl/create
func (h *PaymentHandler) CreateXRPLPayment(c *gin.Context) {
	var req service.CreateXRPLPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Validate request
	if req.ManufacturerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "manufacturer_id is required",
		})
		return
	}

	if req.AmountUSD <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "amount_usd must be greater than 0",
		})
		return
	}

	// Create payment
	payment, err := h.service.CreateXRPLPayment(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create payment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create payment",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    payment,
	})
}

// GetPaymentStatus handles GET /payments/xrpl/:paymentId/status
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	paymentID, err := strconv.ParseInt(c.Param("paymentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment ID",
		})
		return
	}

	payment, err := h.service.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.Error("failed to get payment", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Payment not found",
		})
		return
	}

	// Build status response
	status := gin.H{
		"payment_id":     payment.ID,
		"status":         payment.Status,
		"amount_usd":     payment.AmountUSD,
		"payment_method": payment.PaymentMethod,
		"created_at":     payment.CreatedAt,
		"updated_at":     payment.UpdatedAt,
	}

	if payment.XRPLPayment != nil {
		status["xrpl"] = gin.H{
			"xrp_amount":          payment.XRPLPayment.XRPAmount,
			"exchange_rate":       payment.XRPLPayment.ExchangeRate,
			"destination_address": payment.XRPLPayment.DestinationAddress,
			"destination_tag":     payment.XRPLPayment.DestinationTag.Int32,
			"status":              payment.XRPLPayment.Status,
			"tx_hash":             payment.XRPLPayment.TxHash.String,
			"expires_at":          payment.XRPLPayment.ExpiresAt.Time,
			"confirmed_at":        payment.XRPLPayment.ConfirmedAt.Time,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// VerifyPayment handles POST /payments/xrpl/:paymentId/verify
func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	paymentID, err := strconv.ParseInt(c.Param("paymentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment ID",
		})
		return
	}

	var req struct {
		TxHash string `json:"tx_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	result, err := h.service.VerifyPayment(c.Request.Context(), paymentID, req.TxHash)
	if err != nil {
		h.logger.Error("failed to verify payment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify payment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetExchangeRate handles GET /payments/exchange-rate
func (h *PaymentHandler) GetExchangeRate(c *gin.Context) {
	rate, err := h.service.GetExchangeRate(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get exchange rate", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get exchange rate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rate,
	})
}

// ConvertCurrency handles POST /payments/convert
func (h *PaymentHandler) ConvertCurrency(c *gin.Context) {
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
		From   string  `json:"from" binding:"required"`
		To     string  `json:"to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	result, err := h.service.ConvertCurrency(c.Request.Context(), req.Amount, req.From, req.To)
	if err != nil {
		h.logger.Error("failed to convert currency", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to convert currency",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ListPayments handles GET /payments
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	// Get query parameters
	manufacturerID, _ := strconv.ParseInt(c.Query("manufacturer_id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if manufacturerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "manufacturer_id is required",
		})
		return
	}

	payments, err := h.service.GetPaymentsByManufacturer(c.Request.Context(), manufacturerID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list payments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list payments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"payments": payments,
			"limit":    limit,
			"offset":   offset,
			"count":    len(payments),
		},
	})
}

// GetPayment handles GET /payments/:paymentId
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID, err := strconv.ParseInt(c.Param("paymentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment ID",
		})
		return
	}

	payment, err := h.service.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.Error("failed to get payment", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Payment not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    payment,
	})
}
