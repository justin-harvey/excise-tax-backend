// Package model defines domain models for the tax service.
package model

import (
	"encoding/json"
	"time"
)

// TaxReport represents a tax report submitted by a manufacturer.
// Maps to the tax_reports table in the database.
type TaxReport struct {
	ID                  int64           `json:"id" db:"id"`
	ManufacturerID      int64           `json:"manufacturer_id" db:"manufacturer_id"`
	ReportType          string          `json:"report_type" db:"report_type"` // monthly, quarterly, annual
	FormNumber          string          `json:"form_number,omitempty" db:"form_number"`
	ReportPeriodStart   time.Time       `json:"report_period_start" db:"report_period_start"`
	ReportPeriodEnd     time.Time       `json:"report_period_end" db:"report_period_end"`
	SubmissionDate      time.Time       `json:"submission_date" db:"submission_date"`
	Status              string          `json:"status" db:"status"` // draft, pending, under_review, approved, needs_correction, rejected
	ReviewerID          *int64          `json:"reviewer_id,omitempty" db:"reviewer_id"`
	ReviewedAt          *time.Time      `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewNotes         string          `json:"review_notes,omitempty" db:"review_notes"`
	ProductionData      json.RawMessage `json:"production_data,omitempty" db:"production_data"`
	TaxAmountCalculated float64         `json:"tax_amount_calculated" db:"tax_amount_calculated"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// TaxRate represents a tax rate for a specific product type.
// Maps to the tax_rates table in the database.
type TaxRate struct {
	ID             int64      `json:"id" db:"id"`
	ProductType    string     `json:"product_type" db:"product_type"` // beer, wine, spirits, other
	RatePerUnit    float64    `json:"rate_per_unit" db:"rate_per_unit"`
	UnitType       string     `json:"unit_type" db:"unit_type"` // gallon, barrel, case, liter
	EffectiveDate  time.Time  `json:"effective_date" db:"effective_date"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty" db:"expiration_date"`
	Description    string     `json:"description,omitempty" db:"description"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// ProductionData represents the production data structure stored in JSONB.
type ProductionData struct {
	Items []ProductionItem `json:"items"`
}

// ProductionItem represents a single production item.
type ProductionItem struct {
	ProductType string  `json:"product_type"` // beer, wine, spirits
	ProductName string  `json:"product_name"`
	UnitType    string  `json:"unit_type"` // gallon, barrel, case, liter
	Quantity    float64 `json:"quantity"`
	ABV         float64 `json:"abv,omitempty"` // Alcohol by volume (for spirits proof calculation)
	Notes       string  `json:"notes,omitempty"`
}

// TaxCalculation represents the result of a tax calculation.
type TaxCalculation struct {
	ReportID           int64                   `json:"report_id"`
	TotalTaxAmount     float64                 `json:"total_tax_amount"`
	BreakdownByProduct []TaxCalculationProduct `json:"breakdown_by_product"`
	CalculatedAt       time.Time               `json:"calculated_at"`
}

// TaxCalculationProduct represents tax calculation for a single product.
type TaxCalculationProduct struct {
	ProductType string  `json:"product_type"`
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	UnitType    string  `json:"unit_type"`
	RatePerUnit float64 `json:"rate_per_unit"`
	TaxAmount   float64 `json:"tax_amount"`
}

// ValidationError represents a validation error for a field.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ValidationResult represents the result of report validation.
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// ReportFilter represents filters for querying tax reports.
type ReportFilter struct {
	ManufacturerID *int64     `json:"manufacturer_id,omitempty"`
	Status         *string    `json:"status,omitempty"`
	ReportType     *string    `json:"report_type,omitempty"`
	FromDate       *time.Time `json:"from_date,omitempty"`
	ToDate         *time.Time `json:"to_date,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
}

// CreateReportRequest represents a request to create a tax report.
type CreateReportRequest struct {
	ManufacturerID    int64           `json:"manufacturer_id" binding:"required"`
	ReportType        string          `json:"report_type" binding:"required,oneof=monthly quarterly annual"`
	FormNumber        string          `json:"form_number,omitempty"`
	ReportPeriodStart time.Time       `json:"report_period_start" binding:"required"`
	ReportPeriodEnd   time.Time       `json:"report_period_end" binding:"required"`
	ProductionData    json.RawMessage `json:"production_data" binding:"required"`
}

// UpdateReportRequest represents a request to update a tax report.
type UpdateReportRequest struct {
	ReportType        *string         `json:"report_type,omitempty" binding:"omitempty,oneof=monthly quarterly annual"`
	FormNumber        *string         `json:"form_number,omitempty"`
	ReportPeriodStart *time.Time      `json:"report_period_start,omitempty"`
	ReportPeriodEnd   *time.Time      `json:"report_period_end,omitempty"`
	ProductionData    json.RawMessage `json:"production_data,omitempty"`
}

// SubmitReportResponse represents the response after submitting a report.
type SubmitReportResponse struct {
	ReportID           int64          `json:"report_id"`
	Status             string         `json:"status"`
	TaxAmountCalculated float64       `json:"tax_amount_calculated"`
	SubmissionDate     time.Time      `json:"submission_date"`
	ValidationResult   ValidationResult `json:"validation_result"`
}

// Report status constants
const (
	ReportStatusDraft            = "draft"
	ReportStatusPending          = "pending"
	ReportStatusUnderReview      = "under_review"
	ReportStatusApproved         = "approved"
	ReportStatusNeedsCorrection  = "needs_correction"
	ReportStatusRejected         = "rejected"
)

// Report type constants
const (
	ReportTypeMonthly    = "monthly"
	ReportTypeQuarterly  = "quarterly"
	ReportTypeAnnual     = "annual"
)

// Product type constants
const (
	ProductTypeBeer    = "beer"
	ProductTypeWine    = "wine"
	ProductTypeSpirits = "spirits"
	ProductTypeOther   = "other"
)

// Unit type constants
const (
	UnitTypeGallon = "gallon"
	UnitTypeBarrel = "barrel"
	UnitTypeCase   = "case"
	UnitTypeLiter  = "liter"
)

// IsValidStatus checks if a report status is valid.
func IsValidStatus(status string) bool {
	validStatuses := []string{
		ReportStatusDraft,
		ReportStatusPending,
		ReportStatusUnderReview,
		ReportStatusApproved,
		ReportStatusNeedsCorrection,
		ReportStatusRejected,
	}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsValidReportType checks if a report type is valid.
func IsValidReportType(reportType string) bool {
	validTypes := []string{
		ReportTypeMonthly,
		ReportTypeQuarterly,
		ReportTypeAnnual,
	}
	for _, t := range validTypes {
		if t == reportType {
			return true
		}
	}
	return false
}

// IsValidProductType checks if a product type is valid.
func IsValidProductType(productType string) bool {
	validTypes := []string{
		ProductTypeBeer,
		ProductTypeWine,
		ProductTypeSpirits,
		ProductTypeOther,
	}
	for _, t := range validTypes {
		if t == productType {
			return true
		}
	}
	return false
}

// IsValidUnitType checks if a unit type is valid.
func IsValidUnitType(unitType string) bool {
	validTypes := []string{
		UnitTypeGallon,
		UnitTypeBarrel,
		UnitTypeCase,
		UnitTypeLiter,
	}
	for _, t := range validTypes {
		if t == unitType {
			return true
		}
	}
	return false
}

// CanEdit checks if a report can be edited based on its status.
func (r *TaxReport) CanEdit() bool {
	return r.Status == ReportStatusDraft || r.Status == ReportStatusNeedsCorrection
}

// CanSubmit checks if a report can be submitted based on its status.
func (r *TaxReport) CanSubmit() bool {
	return r.Status == ReportStatusDraft || r.Status == ReportStatusNeedsCorrection
}
