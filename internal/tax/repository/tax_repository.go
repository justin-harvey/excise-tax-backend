// Package repository provides data access layer for tax operations.
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/excise-tax-portal/backend/internal/tax/model"
	"github.com/excise-tax-portal/backend/pkg/database"
	"github.com/jackc/pgx/v5"
)

// TaxRepository defines the interface for tax data operations.
type TaxRepository interface {
	CreateReport(ctx context.Context, report *model.TaxReport) error
	GetReportByID(ctx context.Context, id int64) (*model.TaxReport, error)
	GetReportsByManufacturer(ctx context.Context, filter *model.ReportFilter) ([]*model.TaxReport, error)
	UpdateReport(ctx context.Context, id int64, report *model.TaxReport) error
	SubmitReport(ctx context.Context, id int64) error
	GetTaxRates(ctx context.Context, productType, unitType string, effectiveDate time.Time) (*model.TaxRate, error)
	GetAllCurrentTaxRates(ctx context.Context) ([]*model.TaxRate, error)
	CheckOverlappingReports(ctx context.Context, manufacturerID int64, startDate, endDate time.Time, excludeReportID *int64) (bool, error)
	CountReportsByFilter(ctx context.Context, filter *model.ReportFilter) (int64, error)
}

// taxRepository implements TaxRepository interface.
type taxRepository struct {
	db *database.PostgresDB
}

// NewTaxRepository creates a new tax repository instance.
func NewTaxRepository(db *database.PostgresDB) TaxRepository {
	return &taxRepository{db: db}
}

// CreateReport creates a new tax report in the database.
func (r *taxRepository) CreateReport(ctx context.Context, report *model.TaxReport) error {
	query := `
		INSERT INTO tax_reports (
			manufacturer_id, report_type, form_number, report_period_start,
			report_period_end, submission_date, status, production_data,
			tax_amount_calculated
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		report.ManufacturerID,
		report.ReportType,
		report.FormNumber,
		report.ReportPeriodStart,
		report.ReportPeriodEnd,
		report.SubmissionDate,
		report.Status,
		report.ProductionData,
		report.TaxAmountCalculated,
	).Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create tax report: %w", err)
	}

	return nil
}

// GetReportByID retrieves a tax report by its ID.
func (r *taxRepository) GetReportByID(ctx context.Context, id int64) (*model.TaxReport, error) {
	query := `
		SELECT
			id, manufacturer_id, report_type, form_number, report_period_start,
			report_period_end, submission_date, status, reviewer_id, reviewed_at,
			review_notes, production_data, tax_amount_calculated, created_at, updated_at
		FROM tax_reports
		WHERE id = $1
	`

	report := &model.TaxReport{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&report.ID,
		&report.ManufacturerID,
		&report.ReportType,
		&report.FormNumber,
		&report.ReportPeriodStart,
		&report.ReportPeriodEnd,
		&report.SubmissionDate,
		&report.Status,
		&report.ReviewerID,
		&report.ReviewedAt,
		&report.ReviewNotes,
		&report.ProductionData,
		&report.TaxAmountCalculated,
		&report.CreatedAt,
		&report.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tax report not found")
		}
		return nil, fmt.Errorf("failed to get tax report: %w", err)
	}

	return report, nil
}

// GetReportsByManufacturer retrieves tax reports for a manufacturer with optional filters.
func (r *taxRepository) GetReportsByManufacturer(ctx context.Context, filter *model.ReportFilter) ([]*model.TaxReport, error) {
	query := `
		SELECT
			id, manufacturer_id, report_type, form_number, report_period_start,
			report_period_end, submission_date, status, reviewer_id, reviewed_at,
			review_notes, production_data, tax_amount_calculated, created_at, updated_at
		FROM tax_reports
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if filter.ManufacturerID != nil {
		query += fmt.Sprintf(" AND manufacturer_id = $%d", argCount)
		args = append(args, *filter.ManufacturerID)
		argCount++
	}

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filter.Status)
		argCount++
	}

	if filter.ReportType != nil {
		query += fmt.Sprintf(" AND report_type = $%d", argCount)
		args = append(args, *filter.ReportType)
		argCount++
	}

	if filter.FromDate != nil {
		query += fmt.Sprintf(" AND report_period_end >= $%d", argCount)
		args = append(args, *filter.FromDate)
		argCount++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(" AND report_period_start <= $%d", argCount)
		args = append(args, *filter.ToDate)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tax reports: %w", err)
	}
	defer rows.Close()

	reports := []*model.TaxReport{}
	for rows.Next() {
		report := &model.TaxReport{}
		err := rows.Scan(
			&report.ID,
			&report.ManufacturerID,
			&report.ReportType,
			&report.FormNumber,
			&report.ReportPeriodStart,
			&report.ReportPeriodEnd,
			&report.SubmissionDate,
			&report.Status,
			&report.ReviewerID,
			&report.ReviewedAt,
			&report.ReviewNotes,
			&report.ProductionData,
			&report.TaxAmountCalculated,
			&report.CreatedAt,
			&report.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tax report: %w", err)
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tax reports: %w", err)
	}

	return reports, nil
}

// UpdateReport updates an existing tax report.
func (r *taxRepository) UpdateReport(ctx context.Context, id int64, report *model.TaxReport) error {
	query := `
		UPDATE tax_reports
		SET
			report_type = $2,
			form_number = $3,
			report_period_start = $4,
			report_period_end = $5,
			production_data = $6,
			tax_amount_calculated = $7
		WHERE id = $1 AND (status = 'draft' OR status = 'needs_correction')
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		report.ReportType,
		report.FormNumber,
		report.ReportPeriodStart,
		report.ReportPeriodEnd,
		report.ProductionData,
		report.TaxAmountCalculated,
	).Scan(&report.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("report not found or cannot be updated")
		}
		return fmt.Errorf("failed to update tax report: %w", err)
	}

	return nil
}

// SubmitReport changes a report's status to pending.
func (r *taxRepository) SubmitReport(ctx context.Context, id int64) error {
	query := `
		UPDATE tax_reports
		SET status = 'pending', submission_date = $2
		WHERE id = $1 AND (status = 'draft' OR status = 'needs_correction')
		RETURNING id
	`

	var reportID int64
	err := r.db.QueryRow(ctx, query, id, time.Now()).Scan(&reportID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("report not found or cannot be submitted")
		}
		return fmt.Errorf("failed to submit tax report: %w", err)
	}

	return nil
}

// GetTaxRates retrieves the current tax rate for a product type and unit.
func (r *taxRepository) GetTaxRates(ctx context.Context, productType, unitType string, effectiveDate time.Time) (*model.TaxRate, error) {
	query := `
		SELECT
			id, product_type, rate_per_unit, unit_type, effective_date,
			expiration_date, description, created_at
		FROM tax_rates
		WHERE product_type = $1
		AND unit_type = $2
		AND effective_date <= $3
		AND (expiration_date IS NULL OR expiration_date > $3)
		ORDER BY effective_date DESC
		LIMIT 1
	`

	rate := &model.TaxRate{}
	err := r.db.QueryRow(ctx, query, productType, unitType, effectiveDate).Scan(
		&rate.ID,
		&rate.ProductType,
		&rate.RatePerUnit,
		&rate.UnitType,
		&rate.EffectiveDate,
		&rate.ExpirationDate,
		&rate.Description,
		&rate.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tax rate not found for product type %s and unit %s", productType, unitType)
		}
		return nil, fmt.Errorf("failed to get tax rate: %w", err)
	}

	return rate, nil
}

// GetAllCurrentTaxRates retrieves all current tax rates.
func (r *taxRepository) GetAllCurrentTaxRates(ctx context.Context) ([]*model.TaxRate, error) {
	query := `
		SELECT
			id, product_type, rate_per_unit, unit_type, effective_date,
			expiration_date, description, created_at
		FROM tax_rates
		WHERE effective_date <= CURRENT_DATE
		AND (expiration_date IS NULL OR expiration_date > CURRENT_DATE)
		ORDER BY product_type, unit_type
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tax rates: %w", err)
	}
	defer rows.Close()

	rates := []*model.TaxRate{}
	for rows.Next() {
		rate := &model.TaxRate{}
		err := rows.Scan(
			&rate.ID,
			&rate.ProductType,
			&rate.RatePerUnit,
			&rate.UnitType,
			&rate.EffectiveDate,
			&rate.ExpirationDate,
			&rate.Description,
			&rate.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tax rate: %w", err)
		}
		rates = append(rates, rate)
	}

	return rates, nil
}

// CheckOverlappingReports checks if there are overlapping reports for a manufacturer.
func (r *taxRepository) CheckOverlappingReports(ctx context.Context, manufacturerID int64, startDate, endDate time.Time, excludeReportID *int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM tax_reports
			WHERE manufacturer_id = $1
			AND (
				(report_period_start <= $2 AND report_period_end >= $2) OR
				(report_period_start <= $3 AND report_period_end >= $3) OR
				(report_period_start >= $2 AND report_period_end <= $3)
			)
			AND status NOT IN ('rejected', 'draft')
	`

	args := []interface{}{manufacturerID, startDate, endDate}

	if excludeReportID != nil {
		query += " AND id != $4"
		args = append(args, *excludeReportID)
	}

	query += ")"

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check overlapping reports: %w", err)
	}

	return exists, nil
}

// CountReportsByFilter counts tax reports matching the filter.
func (r *taxRepository) CountReportsByFilter(ctx context.Context, filter *model.ReportFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM tax_reports WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if filter.ManufacturerID != nil {
		query += fmt.Sprintf(" AND manufacturer_id = $%d", argCount)
		args = append(args, *filter.ManufacturerID)
		argCount++
	}

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filter.Status)
		argCount++
	}

	if filter.ReportType != nil {
		query += fmt.Sprintf(" AND report_type = $%d", argCount)
		args = append(args, *filter.ReportType)
		argCount++
	}

	if filter.FromDate != nil {
		query += fmt.Sprintf(" AND report_period_end >= $%d", argCount)
		args = append(args, *filter.FromDate)
		argCount++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(" AND report_period_start <= $%d", argCount)
		args = append(args, *filter.ToDate)
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count reports: %w", err)
	}

	return count, nil
}

// CalculateTaxFromProduction calculates tax amount from production data using stored rates.
func (r *taxRepository) CalculateTaxFromProduction(ctx context.Context, productionData json.RawMessage) (float64, error) {
	var data model.ProductionData
	if err := json.Unmarshal(productionData, &data); err != nil {
		return 0, fmt.Errorf("failed to parse production data: %w", err)
	}

	totalTax := 0.0
	for _, item := range data.Items {
		rate, err := r.GetTaxRates(ctx, item.ProductType, item.UnitType, time.Now())
		if err != nil {
			// If no rate found, skip this item
			continue
		}

		itemTax := item.Quantity * rate.RatePerUnit
		totalTax += itemTax
	}

	return totalTax, nil
}
