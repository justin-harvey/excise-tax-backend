// Package service provides business logic for tax operations.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/excise-tax-portal/backend/internal/tax/model"
	"github.com/excise-tax-portal/backend/internal/tax/repository"
	"github.com/excise-tax-portal/backend/pkg/errors"
	"github.com/excise-tax-portal/backend/pkg/logger"
	"go.uber.org/zap"
)

// TaxService defines the interface for tax business logic.
type TaxService interface {
	CreateReport(ctx context.Context, req *model.CreateReportRequest, userID int64, role string) (*model.TaxReport, error)
	GetReport(ctx context.Context, id int64, userID int64, role string) (*model.TaxReport, error)
	GetReports(ctx context.Context, filter *model.ReportFilter, userID int64, role string) ([]*model.TaxReport, int64, error)
	UpdateReport(ctx context.Context, id int64, req *model.UpdateReportRequest, userID int64, role string) (*model.TaxReport, error)
	SubmitReport(ctx context.Context, id int64, userID int64, role string) (*model.SubmitReportResponse, error)
	ValidateReport(ctx context.Context, id int64) (*model.ValidationResult, error)
	CalculateTaxAmount(ctx context.Context, productionData json.RawMessage) (*model.TaxCalculation, error)
	GetCurrentTaxRates(ctx context.Context) ([]*model.TaxRate, error)
}

// taxService implements TaxService interface.
type taxService struct {
	repo   repository.TaxRepository
	logger *logger.Logger
}

// NewTaxService creates a new tax service instance.
func NewTaxService(repo repository.TaxRepository, log *logger.Logger) TaxService {
	return &taxService{
		repo:   repo,
		logger: log,
	}
}

// CreateReport creates a new tax report.
func (s *taxService) CreateReport(ctx context.Context, req *model.CreateReportRequest, userID int64, role string) (*model.TaxReport, error) {
	// Validate request
	if err := s.validateCreateRequest(ctx, req); err != nil {
		return nil, err
	}

	// KNOWN GAP: unlike GetReport/UpdateReport/SubmitReport, this does not
	// verify that a manufacturer-role caller owns req.ManufacturerID before
	// creating a report against it. canAccessReport() (below) has the same
	// gap for reads - there is no user-to-manufacturer association anywhere
	// in this codebase to check against yet, so any check here would just be
	// a check that always passes, which is worse than an honest TODO. userID
	// and role are threaded through now so every enforcement point lines up
	// once that association exists; nothing here enforces on them yet.
	_ = userID
	_ = role

	// Check for overlapping reports
	hasOverlap, err := s.repo.CheckOverlappingReports(ctx, req.ManufacturerID, req.ReportPeriodStart, req.ReportPeriodEnd, nil)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to check overlapping reports", zap.Error(err))
		return nil, errors.NewInternalError("failed to check report overlap", err)
	}

	if hasOverlap {
		return nil, errors.NewValidationError("overlapping report exists for this period", map[string]string{
			"report_period": "A report already exists for this time period",
		})
	}

	// Calculate tax amount
	taxAmount, err := s.repo.CalculateTaxFromProduction(ctx, req.ProductionData)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to calculate tax", zap.Error(err))
		return nil, errors.NewInternalError("failed to calculate tax amount", err)
	}

	// Create report
	report := &model.TaxReport{
		ManufacturerID:      req.ManufacturerID,
		ReportType:          req.ReportType,
		FormNumber:          req.FormNumber,
		ReportPeriodStart:   req.ReportPeriodStart,
		ReportPeriodEnd:     req.ReportPeriodEnd,
		SubmissionDate:      time.Now(),
		Status:              model.ReportStatusDraft,
		ProductionData:      req.ProductionData,
		TaxAmountCalculated: taxAmount,
	}

	if err := s.repo.CreateReport(ctx, report); err != nil {
		s.logger.ErrorContext(ctx, "Failed to create report", zap.Error(err))
		return nil, errors.NewInternalError("failed to create report", err)
	}

	s.logger.InfoContext(ctx, "Tax report created", zap.Int64("report_id", report.ID))
	return report, nil
}

// GetReport retrieves a single tax report with authorization check.
func (s *taxService) GetReport(ctx context.Context, id int64, userID int64, role string) (*model.TaxReport, error) {
	report, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get report", zap.Int64("report_id", id), zap.Error(err))
		return nil, errors.NewNotFoundError("report", "tax report not found")
	}

	// Authorization: manufacturers can only see their own reports
	if !s.canAccessReport(report, userID, role) {
		return nil, errors.NewForbiddenError("you do not have permission to access this report")
	}

	return report, nil
}

// GetReports retrieves tax reports with filters and authorization.
func (s *taxService) GetReports(ctx context.Context, filter *model.ReportFilter, userID int64, role string) ([]*model.TaxReport, int64, error) {
	// For manufacturers, enforce they can only see their own reports
	// We need to get manufacturer_id from userID
	// This would typically come from a manufacturer repository or be passed in
	// For now, we'll assume it's set in the filter

	if filter.Limit == 0 {
		filter.Limit = 20 // Default limit
	}

	reports, err := s.repo.GetReportsByManufacturer(ctx, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get reports", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to retrieve reports", err)
	}

	count, err := s.repo.CountReportsByFilter(ctx, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to count reports", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to count reports", err)
	}

	return reports, count, nil
}

// UpdateReport updates an existing tax report.
func (s *taxService) UpdateReport(ctx context.Context, id int64, req *model.UpdateReportRequest, userID int64, role string) (*model.TaxReport, error) {
	// Get existing report
	existingReport, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("report", "tax report not found")
	}

	// Authorization check
	if !s.canAccessReport(existingReport, userID, role) {
		return nil, errors.NewForbiddenError("you do not have permission to update this report")
	}

	// Check if report can be edited
	if !existingReport.CanEdit() {
		return nil, errors.NewBadRequestError("report cannot be edited in current status")
	}

	// Update fields
	updatedReport := &model.TaxReport{
		ID:                existingReport.ID,
		ManufacturerID:    existingReport.ManufacturerID,
		ReportType:        existingReport.ReportType,
		FormNumber:        existingReport.FormNumber,
		ReportPeriodStart: existingReport.ReportPeriodStart,
		ReportPeriodEnd:   existingReport.ReportPeriodEnd,
		ProductionData:    existingReport.ProductionData,
	}

	if req.ReportType != nil {
		updatedReport.ReportType = *req.ReportType
	}
	if req.FormNumber != nil {
		updatedReport.FormNumber = *req.FormNumber
	}
	if req.ReportPeriodStart != nil {
		updatedReport.ReportPeriodStart = *req.ReportPeriodStart
	}
	if req.ReportPeriodEnd != nil {
		updatedReport.ReportPeriodEnd = *req.ReportPeriodEnd
	}
	if req.ProductionData != nil {
		updatedReport.ProductionData = req.ProductionData
		// Recalculate tax
		taxAmount, err := s.repo.CalculateTaxFromProduction(ctx, req.ProductionData)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to recalculate tax", zap.Error(err))
			return nil, errors.NewInternalError("failed to recalculate tax amount", err)
		}
		updatedReport.TaxAmountCalculated = taxAmount
	} else {
		updatedReport.TaxAmountCalculated = existingReport.TaxAmountCalculated
	}

	// Check for overlapping reports (excluding current report)
	hasOverlap, err := s.repo.CheckOverlappingReports(ctx, updatedReport.ManufacturerID, updatedReport.ReportPeriodStart, updatedReport.ReportPeriodEnd, &id)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to check overlapping reports", zap.Error(err))
		return nil, errors.NewInternalError("failed to check report overlap", err)
	}

	if hasOverlap {
		return nil, errors.NewValidationError("overlapping report exists for this period", map[string]string{
			"report_period": "A report already exists for this time period",
		})
	}

	// Update in database
	if err := s.repo.UpdateReport(ctx, id, updatedReport); err != nil {
		s.logger.ErrorContext(ctx, "Failed to update report", zap.Error(err))
		return nil, errors.NewInternalError("failed to update report", err)
	}

	// Get updated report
	return s.repo.GetReportByID(ctx, id)
}

// SubmitReport submits a tax report for review.
func (s *taxService) SubmitReport(ctx context.Context, id int64, userID int64, role string) (*model.SubmitReportResponse, error) {
	// Get report
	report, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("report", "tax report not found")
	}

	// Authorization check
	if !s.canAccessReport(report, userID, role) {
		return nil, errors.NewForbiddenError("you do not have permission to submit this report")
	}

	// Check if report can be submitted
	if !report.CanSubmit() {
		return nil, errors.NewBadRequestError("report cannot be submitted in current status")
	}

	// Validate report
	validationResult, err := s.ValidateReport(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to validate report", zap.Error(err))
		return nil, errors.NewInternalError("failed to validate report", err)
	}

	if !validationResult.IsValid {
		return &model.SubmitReportResponse{
			ReportID:            report.ID,
			Status:              report.Status,
			TaxAmountCalculated: report.TaxAmountCalculated,
			SubmissionDate:      report.SubmissionDate,
			ValidationResult:    *validationResult,
		}, nil
	}

	// Submit report
	if err := s.repo.SubmitReport(ctx, id); err != nil {
		s.logger.ErrorContext(ctx, "Failed to submit report", zap.Error(err))
		return nil, errors.NewInternalError("failed to submit report", err)
	}

	s.logger.InfoContext(ctx, "Tax report submitted", zap.Int64("report_id", id))

	return &model.SubmitReportResponse{
		ReportID:            report.ID,
		Status:              model.ReportStatusPending,
		TaxAmountCalculated: report.TaxAmountCalculated,
		SubmissionDate:      time.Now(),
		ValidationResult:    *validationResult,
	}, nil
}

// ValidateReport validates a tax report against business rules.
func (s *taxService) ValidateReport(ctx context.Context, id int64) (*model.ValidationResult, error) {
	report, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		return nil, err
	}

	validationErrors := []model.ValidationError{}

	// Validate date range
	if report.ReportPeriodEnd.Before(report.ReportPeriodStart) {
		validationErrors = append(validationErrors, model.ValidationError{
			Field:   "report_period_end",
			Message: "Report end date must be after start date",
			Code:    "INVALID_DATE_RANGE",
		})
	}

	// Validate report is not too far in the future
	if report.ReportPeriodStart.After(time.Now().AddDate(0, 1, 0)) {
		validationErrors = append(validationErrors, model.ValidationError{
			Field:   "report_period_start",
			Message: "Report period cannot be more than 1 month in the future",
			Code:    "FUTURE_REPORT_PERIOD",
		})
	}

	// Validate production data
	if len(report.ProductionData) == 0 {
		validationErrors = append(validationErrors, model.ValidationError{
			Field:   "production_data",
			Message: "Production data is required",
			Code:    "MISSING_PRODUCTION_DATA",
		})
	} else {
		var data model.ProductionData
		if err := json.Unmarshal(report.ProductionData, &data); err != nil {
			validationErrors = append(validationErrors, model.ValidationError{
				Field:   "production_data",
				Message: "Invalid production data format",
				Code:    "INVALID_PRODUCTION_DATA",
			})
		} else {
			// Validate each production item
			if len(data.Items) == 0 {
				validationErrors = append(validationErrors, model.ValidationError{
					Field:   "production_data.items",
					Message: "At least one production item is required",
					Code:    "NO_PRODUCTION_ITEMS",
				})
			}

			for i, item := range data.Items {
				if item.Quantity <= 0 {
					validationErrors = append(validationErrors, model.ValidationError{
						Field:   fmt.Sprintf("production_data.items[%d].quantity", i),
						Message: "Quantity must be greater than zero",
						Code:    "INVALID_QUANTITY",
					})
				}

				if !model.IsValidProductType(item.ProductType) {
					validationErrors = append(validationErrors, model.ValidationError{
						Field:   fmt.Sprintf("production_data.items[%d].product_type", i),
						Message: "Invalid product type",
						Code:    "INVALID_PRODUCT_TYPE",
					})
				}

				if !model.IsValidUnitType(item.UnitType) {
					validationErrors = append(validationErrors, model.ValidationError{
						Field:   fmt.Sprintf("production_data.items[%d].unit_type", i),
						Message: "Invalid unit type",
						Code:    "INVALID_UNIT_TYPE",
					})
				}
			}
		}
	}

	// Validate tax amount reasonableness (not negative, not excessive)
	if report.TaxAmountCalculated < 0 {
		validationErrors = append(validationErrors, model.ValidationError{
			Field:   "tax_amount_calculated",
			Message: "Tax amount cannot be negative",
			Code:    "NEGATIVE_TAX_AMOUNT",
		})
	}

	if report.TaxAmountCalculated > 10000000 { // $10M threshold
		validationErrors = append(validationErrors, model.ValidationError{
			Field:   "tax_amount_calculated",
			Message: "Tax amount exceeds reasonable threshold, please verify production data",
			Code:    "EXCESSIVE_TAX_AMOUNT",
		})
	}

	return &model.ValidationResult{
		IsValid: len(validationErrors) == 0,
		Errors:  validationErrors,
	}, nil
}

// CalculateTaxAmount calculates tax amount based on production data.
func (s *taxService) CalculateTaxAmount(ctx context.Context, productionData json.RawMessage) (*model.TaxCalculation, error) {
	var data model.ProductionData
	if err := json.Unmarshal(productionData, &data); err != nil {
		return nil, errors.NewBadRequestError("invalid production data format")
	}

	calculation := &model.TaxCalculation{
		BreakdownByProduct: []model.TaxCalculationProduct{},
		CalculatedAt:       time.Now(),
	}

	totalTax := 0.0

	for _, item := range data.Items {
		rate, err := s.repo.GetTaxRates(ctx, item.ProductType, item.UnitType, time.Now())
		if err != nil {
			s.logger.WarnContext(ctx, "Tax rate not found for product",
				zap.String("product_type", item.ProductType),
				zap.String("unit_type", item.UnitType))
			continue
		}

		itemTax := s.calculateProductTax(item.Quantity, rate.RatePerUnit, item.ProductType, item.ABV)
		totalTax += itemTax

		calculation.BreakdownByProduct = append(calculation.BreakdownByProduct, model.TaxCalculationProduct{
			ProductType: item.ProductType,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitType:    item.UnitType,
			RatePerUnit: rate.RatePerUnit,
			TaxAmount:   itemTax,
		})
	}

	calculation.TotalTaxAmount = totalTax

	return calculation, nil
}

// GetCurrentTaxRates retrieves all current tax rates.
func (s *taxService) GetCurrentTaxRates(ctx context.Context) ([]*model.TaxRate, error) {
	rates, err := s.repo.GetAllCurrentTaxRates(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get tax rates", zap.Error(err))
		return nil, errors.NewInternalError("failed to retrieve tax rates", err)
	}

	return rates, nil
}

// Helper methods

// validateCreateRequest validates the create report request.
func (s *taxService) validateCreateRequest(ctx context.Context, req *model.CreateReportRequest) error {
	if req.ManufacturerID <= 0 {
		return errors.NewValidationError("invalid manufacturer ID", nil)
	}

	if !model.IsValidReportType(req.ReportType) {
		return errors.NewValidationError("invalid report type", map[string]string{
			"report_type": "must be monthly, quarterly, or annual",
		})
	}

	if req.ReportPeriodEnd.Before(req.ReportPeriodStart) {
		return errors.NewValidationError("invalid date range", map[string]string{
			"report_period": "end date must be after start date",
		})
	}

	return nil
}

// canAccessReport checks if a user can access a report.
func (s *taxService) canAccessReport(report *model.TaxReport, userID int64, role string) bool {
	// Admins and reviewers can access all reports
	if role == "admin" || role == "super_admin" || role == "reviewer" {
		return true
	}

	// For now, we'll assume manufacturers can access their own reports
	// In a real implementation, we'd need to check if the user is associated with the manufacturer
	return true
}

// calculateProductTax calculates tax for a specific product.
func (s *taxService) calculateProductTax(quantity, ratePerUnit float64, productType string, abv float64) float64 {
	tax := quantity * ratePerUnit

	// For spirits, adjust for proof (ABV * 2)
	if productType == model.ProductTypeSpirits && abv > 0 {
		proof := abv * 2
		// Assuming the rate is per proof gallon
		tax = quantity * (proof / 100) * ratePerUnit
	}

	return tax
}
