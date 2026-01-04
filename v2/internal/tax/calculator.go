// Package tax provides excise tax calculation functionality.
package tax

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Calculator provides tax calculation operations.
type Calculator struct {
	rates  []TaxRate
	logger *slog.Logger
}

// Option configures a Calculator.
type Option func(*Calculator)

// WithLogger sets the logger for the calculator.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Calculator) {
		c.logger = logger
	}
}

// NewCalculator creates a new tax calculator with the given tax rates.
func NewCalculator(rates []TaxRate, opts ...Option) *Calculator {
	c := &Calculator{
		rates:  rates,
		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Calculate performs tax calculation for the given production data.
// This is a pure function that takes inputs and produces outputs without side effects.
func (c *Calculator) Calculate(ctx context.Context, data *ProductionData, jurisdiction string, effectiveDate time.Time) (*TaxCalculation, error) {
	if data == nil || len(data.Items) == 0 {
		return nil, ErrNoProductionItems
	}

	result := &TaxCalculation{
		BreakdownByProduct: make([]TaxCalculationProduct, 0, len(data.Items)),
		CalculatedAt:       time.Now(),
		Jurisdiction:       jurisdiction,
	}

	totalTax := 0.0

	for _, item := range data.Items {
		// Validate item
		if err := c.validateProductionItem(&item); err != nil {
			return nil, fmt.Errorf("invalid production item %q: %w", item.ProductName, err)
		}

		// Find applicable tax rate
		rate, err := c.findApplicableRate(item.ProductType, item.UnitType, jurisdiction, effectiveDate)
		if err != nil {
			return nil, fmt.Errorf("no tax rate for %s (%s): %w", item.ProductName, item.ProductType, err)
		}

		// Convert quantity to the rate's unit type if needed
		quantity := c.convertQuantity(item.Quantity, item.UnitType, rate.UnitType)

		// Calculate tax for this item
		taxAmount := quantity * rate.RatePerUnit

		// Add to breakdown
		result.BreakdownByProduct = append(result.BreakdownByProduct, TaxCalculationProduct{
			ProductType: item.ProductType,
			ProductName: item.ProductName,
			Quantity:    quantity,
			UnitType:    rate.UnitType,
			RatePerUnit: rate.RatePerUnit,
			TaxAmount:   taxAmount,
		})

		totalTax += taxAmount
	}

	result.TotalTaxAmount = roundToTwoDecimals(totalTax)

	return result, nil
}

// CalculateForItem calculates tax for a single production item.
// Convenience function for calculating tax on a single item.
func (c *Calculator) CalculateForItem(ctx context.Context, item ProductionItem, jurisdiction string, effectiveDate time.Time) (float64, error) {
	data := &ProductionData{Items: []ProductionItem{item}}
	result, err := c.Calculate(ctx, data, jurisdiction, effectiveDate)
	if err != nil {
		return 0, err
	}
	return result.TotalTaxAmount, nil
}

// GetRate retrieves a tax rate for the given criteria.
func (c *Calculator) GetRate(productType ProductType, unitType UnitType, jurisdiction string, effectiveDate time.Time) (*TaxRate, error) {
	return c.findApplicableRate(productType, unitType, jurisdiction, effectiveDate)
}

// GetAllRates returns all loaded tax rates.
func (c *Calculator) GetAllRates() []TaxRate {
	// Return a copy to prevent modification
	rates := make([]TaxRate, len(c.rates))
	copy(rates, c.rates)
	return rates
}

// findApplicableRate finds the applicable tax rate for given criteria.
func (c *Calculator) findApplicableRate(productType ProductType, unitType UnitType, jurisdiction string, effectiveDate time.Time) (*TaxRate, error) {
	for _, rate := range c.rates {
		// Match product type
		if rate.ProductType != productType {
			continue
		}

		// Match unit type
		if rate.UnitType != unitType {
			continue
		}

		// Match jurisdiction (case-insensitive)
		if rate.Jurisdiction != jurisdiction {
			continue
		}

		// Check if rate is effective on the given date
		if effectiveDate.Before(rate.EffectiveDate) {
			continue
		}

		// Check if rate has expired
		if rate.ExpirationDate != nil && effectiveDate.After(*rate.ExpirationDate) {
			continue
		}

		// Found a matching rate
		return &rate, nil
	}

	return nil, ErrNoTaxRate
}

// validateProductionItem validates a production item.
func (c *Calculator) validateProductionItem(item *ProductionItem) error {
	// Validate product type
	if !isValidProductType(item.ProductType) {
		return &ValidationErr{
			Field:   "product_type",
			Message: fmt.Sprintf("unsupported product type: %s", item.ProductType),
			Err:     ErrInvalidProductType,
		}
	}

	// Validate unit type
	if !isValidUnitType(item.UnitType) {
		return &ValidationErr{
			Field:   "unit_type",
			Message: fmt.Sprintf("unsupported unit type: %s", item.UnitType),
			Err:     ErrInvalidUnitType,
		}
	}

	// Validate quantity
	if item.Quantity <= 0 {
		return &ValidationErr{
			Field:   "quantity",
			Message: "quantity must be greater than zero",
			Err:     ErrInvalidQuantity,
		}
	}

	// Validate ABV if provided
	if item.ABV < 0 || item.ABV > 100 {
		return &ValidationErr{
			Field:   "abv",
			Message: "ABV must be between 0 and 100",
			Err:     ErrInvalidABV,
		}
	}

	return nil
}

// convertQuantity converts quantity from one unit type to another.
// For now, we'll require exact unit matches. In the future, this could handle conversions.
func (c *Calculator) convertQuantity(quantity float64, fromUnit, toUnit UnitType) float64 {
	if fromUnit == toUnit {
		return quantity
	}

	// Future: implement unit conversions
	// For now, log a warning and return as-is
	c.logger.Warn("unit conversion not implemented",
		"from", fromUnit,
		"to", toUnit,
		"quantity", quantity)

	return quantity
}

// isValidProductType checks if a product type is valid.
func isValidProductType(pt ProductType) bool {
	switch pt {
	case ProductTypeBeer, ProductTypeWine, ProductTypeSpirits, ProductTypeOther:
		return true
	default:
		return false
	}
}

// isValidUnitType checks if a unit type is valid.
func isValidUnitType(ut UnitType) bool {
	switch ut {
	case UnitTypeGallon, UnitTypeBarrel, UnitTypeCase, UnitTypeLiter:
		return true
	default:
		return false
	}
}

// roundToTwoDecimals rounds a float to 2 decimal places.
func roundToTwoDecimals(val float64) float64 {
	return float64(int(val*100+0.5)) / 100
}

// GetRates returns tax rates, optionally filtered by jurisdiction
func (c *Calculator) GetRates(jurisdiction string) ([]TaxRate, error) {
	if jurisdiction == "" {
		// Return all rates
		return c.rates, nil
	}

	// Filter by jurisdiction
	var filteredRates []TaxRate
	for _, rate := range c.rates {
		if rate.Jurisdiction == jurisdiction {
			filteredRates = append(filteredRates, rate)
		}
	}

	return filteredRates, nil
}
