// Package handler provides HTTP handlers for tax operations.
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/excise-tax-portal/backend/internal/tax/model"
	"github.com/excise-tax-portal/backend/internal/tax/service"
	"github.com/excise-tax-portal/backend/pkg/errors"
	"github.com/excise-tax-portal/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TaxHandler handles HTTP requests for tax operations.
type TaxHandler struct {
	service service.TaxService
	logger  *logger.Logger
}

// NewTaxHandler creates a new tax handler instance.
func NewTaxHandler(service service.TaxService, log *logger.Logger) *TaxHandler {
	return &TaxHandler{
		service: service,
		logger:  log,
	}
}

// CreateReport godoc
// @Summary Create a new tax report
// @Description Create a draft tax report
// @Tags tax-reports
// @Accept json
// @Produce json
// @Param request body model.CreateReportRequest true "Create Report Request"
// @Success 201 {object} model.TaxReport
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /reports [post]
func (h *TaxHandler) CreateReport(c *gin.Context) {
	ctx := c.Request.Context()

	var req model.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "Invalid request body", zap.Error(err))
		h.respondWithError(c, errors.NewBadRequestError("invalid request body"))
		return
	}

	// Get user info from context (set by auth middleware)
	userID := h.getUserID(c)
	role := h.getUserRole(c)

	// For manufacturers, ensure they can only create reports for their own manufacturer_id
	// This would require getting manufacturer_id from user_id via a repository
	// For now, we'll trust the manufacturer_id in the request

	report, err := h.service.CreateReport(ctx, &req, userID, role)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to create report", zap.Error(err))
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, report)
}

// GetReports godoc
// @Summary List tax reports
// @Description Get tax reports with optional filters
// @Tags tax-reports
// @Accept json
// @Produce json
// @Param manufacturer_id query int false "Manufacturer ID"
// @Param status query string false "Report Status"
// @Param report_type query string false "Report Type"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /reports [get]
func (h *TaxHandler) GetReports(c *gin.Context) {
	ctx := c.Request.Context()

	filter := &model.ReportFilter{}

	if manufacturerID := c.Query("manufacturer_id"); manufacturerID != "" {
		id, err := strconv.ParseInt(manufacturerID, 10, 64)
		if err == nil {
			filter.ManufacturerID = &id
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	if reportType := c.Query("report_type"); reportType != "" {
		filter.ReportType = &reportType
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	filter.Limit = limit

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	filter.Offset = offset

	userID := h.getUserID(c)
	role := h.getUserRole(c)

	reports, total, err := h.service.GetReports(ctx, filter, userID, role)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get reports", zap.Error(err))
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   reports,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetReport godoc
// @Summary Get a tax report by ID
// @Description Retrieve a single tax report
// @Tags tax-reports
// @Accept json
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {object} model.TaxReport
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /reports/{id} [get]
func (h *TaxHandler) GetReport(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondWithError(c, errors.NewBadRequestError("invalid report ID"))
		return
	}

	userID := h.getUserID(c)
	role := h.getUserRole(c)

	report, err := h.service.GetReport(ctx, id, userID, role)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get report", zap.Int64("report_id", id), zap.Error(err))
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

// UpdateReport godoc
// @Summary Update a tax report
// @Description Update a draft or needs_correction tax report
// @Tags tax-reports
// @Accept json
// @Produce json
// @Param id path int true "Report ID"
// @Param request body model.UpdateReportRequest true "Update Report Request"
// @Success 200 {object} model.TaxReport
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /reports/{id} [put]
func (h *TaxHandler) UpdateReport(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondWithError(c, errors.NewBadRequestError("invalid report ID"))
		return
	}

	var req model.UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "Invalid request body", zap.Error(err))
		h.respondWithError(c, errors.NewBadRequestError("invalid request body"))
		return
	}

	userID := h.getUserID(c)
	role := h.getUserRole(c)

	report, err := h.service.UpdateReport(ctx, id, &req, userID, role)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to update report", zap.Int64("report_id", id), zap.Error(err))
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

// SubmitReport godoc
// @Summary Submit a tax report
// @Description Submit a draft tax report for review
// @Tags tax-reports
// @Accept json
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {object} model.SubmitReportResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /reports/{id}/submit [post]
func (h *TaxHandler) SubmitReport(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondWithError(c, errors.NewBadRequestError("invalid report ID"))
		return
	}

	userID := h.getUserID(c)
	role := h.getUserRole(c)

	response, err := h.service.SubmitReport(ctx, id, userID, role)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to submit report", zap.Int64("report_id", id), zap.Error(err))
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ValidateReport godoc
// @Summary Validate a tax report
// @Description Validate a tax report against business rules
// @Tags tax-reports
// @Accept json
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {object} model.ValidationResult
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /reports/{id}/validate [post]
func (h *TaxHandler) ValidateReport(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondWithError(c, errors.NewBadRequestError("invalid report ID"))
		return
	}

	result, err := h.service.ValidateReport(ctx, id)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to validate report", zap.Int64("report_id", id), zap.Error(err))
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CalculateTax godoc
// @Summary Calculate tax amount
// @Description Calculate tax amount based on production data
// @Tags tax-reports
// @Accept json
// @Produce json
// @Param production_data body model.ProductionData true "Production Data"
// @Success 200 {object} model.TaxCalculation
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /reports/calculate [post]
func (h *TaxHandler) CalculateTax(c *gin.Context) {
	ctx := c.Request.Context()

	var productionData model.ProductionData
	if err := c.ShouldBindJSON(&productionData); err != nil {
		h.logger.WarnContext(ctx, "Invalid request body", zap.Error(err))
		h.respondWithError(c, errors.NewBadRequestError("invalid production data"))
		return
	}

	// Marshal the whole payload - CalculateTaxAmount/CalculateTaxFromProduction
	// iterate every item in Items, so encoding only Items[0] would silently
	// drop every other line item from the calculation.
	jsonData, err := json.Marshal(productionData)
	if err != nil {
		h.respondWithError(c, errors.NewBadRequestError("failed to process production data"))
		return
	}

	calculation, err := h.service.CalculateTaxAmount(ctx, jsonData)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to calculate tax", zap.Error(err))
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, calculation)
}

// GetTaxRates godoc
// @Summary Get current tax rates
// @Description Retrieve all current tax rates
// @Tags tax-rates
// @Accept json
// @Produce json
// @Success 200 {array} model.TaxRate
// @Failure 500 {object} errors.ErrorResponse
// @Router /tax-rates [get]
func (h *TaxHandler) GetTaxRates(c *gin.Context) {
	ctx := c.Request.Context()

	rates, err := h.service.GetCurrentTaxRates(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get tax rates", zap.Error(err))
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, rates)
}

// Helper methods

// getUserID extracts user ID from context (set by auth middleware).
func (h *TaxHandler) getUserID(c *gin.Context) int64 {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int64); ok {
			return id
		}
	}
	return 0
}

// getUserRole extracts user role from context (set by auth middleware).
func (h *TaxHandler) getUserRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return ""
}

// respondWithError responds with an appropriate error response.
func (h *TaxHandler) respondWithError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.HTTPStatus, errors.NewErrorResponse(appErr))
		return
	}

	// Default to internal server error
	appErr := errors.NewInternalError("an unexpected error occurred", err)
	c.JSON(appErr.HTTPStatus, errors.NewErrorResponse(appErr))
}

// RegisterRoutes registers all tax routes.
func (h *TaxHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Tax reports endpoints
	reports := r.Group("/reports")
	{
		reports.POST("", h.CreateReport)
		reports.GET("", h.GetReports)
		reports.POST("/calculate", h.CalculateTax)
		reports.GET("/:id", h.GetReport)
		reports.PUT("/:id", h.UpdateReport)
		reports.POST("/:id/submit", h.SubmitReport)
		reports.POST("/:id/validate", h.ValidateReport)
		// Future: reports.GET("/:id/pdf", h.GeneratePDF)
	}

	// Tax rates endpoints
	r.GET("/tax-rates", h.GetTaxRates)
}
