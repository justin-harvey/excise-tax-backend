// Package tax provides excise tax calculation functionality for alcoholic beverages.
package tax

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Calculator provides excise tax calculation operations.
// It is safe for concurrent use.
type Calculator struct {
	rates  []TaxRate
	logger *slog.Logger
}

// Option configures a Calculator.
type Option func(*Calculator)

// WithRates sets custom tax rates for the calculator.
func WithRates(rates []TaxRate) Option {
	return func(c *Calculator) {
		c.rates = make([]TaxRate, len(rates))
		copy(c.rates, rates)
	}
}

// WithLogger sets a custom logger for the calculator.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Calculator) {
		c.logger = logger
	}
}

// WithRatesFromFile loads tax rates from a JSON file.
func WithRatesFromFile(filename string) Option {
	return func(c *Calculator) {
		rates, err := LoadRatesFromJSON(filename)
		if err != nil {
			// Use default rates and log error
			if c.logger != nil {
				c.logger.Error("failed to load rates from file, using defaults",
					"file", filename, "error", err)
			}
			return
		}
		c.rates = rates
	}
}

// New creates a new tax calculator with optional configuration.
// By default, it uses federal excise tax rates and a default logger.
func New(opts ...Option) (*Calculator, error) {
	c := &Calculator{
		rates:  DefaultFederalRates(),
		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
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
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Validate item
		if err := c.validateProductionItem(&item); err != nil {
			return nil, NewCalculationError("invalid production item", item.ProductName, err)
		}

		// Find applicable tax rate
		rate, err := c.findApplicableRate(item.ProductType, item.UnitType, jurisdiction, effectiveDate)
		if err != nil {
			return nil, NewCalculationError("no tax rate found", item.ProductName, err)
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
func (c *Calculator) CalculateForItem(ctx context.Context, item ProductionItem, jurisdiction string, effectiveDate time.Time) (float64, error) {
	data := &ProductionData{Items: []ProductionItem{item}}
	result, err := c.Calculate(ctx, data, jurisdiction, effectiveDate)
	if err != nil {
		return 0, err
	}
	return result.TotalTaxAmount, nil
}

// GetRate retrieves the applicable tax rate for given criteria.
func (c *Calculator) GetRate(productType ProductType, unitType UnitType, jurisdiction string, effectiveDate time.Time) (*TaxRate, error) {
	return c.findApplicableRate(productType, unitType, jurisdiction, effectiveDate)
}

// GetAllRates returns all tax rates currently loaded in the calculator.
func (c *Calculator) GetAllRates() []TaxRate {
	rates := make([]TaxRate, len(c.rates))
	copy(rates, c.rates)
	return rates
}

// ValidateProductionData validates production data without performing calculations.
func (c *Calculator) ValidateProductionData(data *ProductionData) []ValidationError {
	var errors []ValidationError

	if data == nil {
		errors = append(errors, ValidationError{
			Field:   "data",
			Message: "production data is required",
		})
		return errors
	}

	if len(data.Items) == 0 {
		errors = append(errors, ValidationError{
			Field:   "items",
			Message: "at least one production item is required",
		})
		return errors
	}

	for i, item := range data.Items {
		if err := c.validateProductionItem(&item); err != nil {
			if validationErr, ok := err.(*ValidationErr); ok {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("items[%d].%s", i, validationErr.Field),
					Message: validationErr.Message,
				})
			} else {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("items[%d]", i),
					Message: err.Error(),
				})
			}
		}
	}

	return errors
}

// Private helper methods

func (c *Calculator) findApplicableRate(productType ProductType, unitType UnitType, jurisdiction string, effectiveDate time.Time) (*TaxRate, error) {
	for _, rate := range c.rates {
		if rate.ProductType != productType {
			continue
		}
		if rate.UnitType != unitType {
			continue
		}
		if rate.Jurisdiction != jurisdiction {
			continue
		}
		if effectiveDate.Before(rate.EffectiveDate) {
			continue
		}
		if rate.ExpirationDate != nil && effectiveDate.After(*rate.ExpirationDate) {
			continue
		}
		return &rate, nil
	}
	return nil, ErrNoTaxRate
}

func (c *Calculator) validateProductionItem(item *ProductionItem) error {
	if !isValidProductType(item.ProductType) {
		return &ValidationErr{
			Field:   "product_type",
			Message: "unsupported product type: " + string(item.ProductType),
			Err:     ErrInvalidProductType,
		}
	}

	if !isValidUnitType(item.UnitType) {
		return &ValidationErr{
			Field:   "unit_type",
			Message: "unsupported unit type: " + string(item.UnitType),
			Err:     ErrInvalidUnitType,
		}
	}

	if item.Quantity <= 0 {
		return &ValidationErr{
			Field:   "quantity",
			Message: "quantity must be greater than zero",
			Err:     ErrInvalidQuantity,
		}
	}

	if item.ABV < 0 || item.ABV > 100 {
		return &ValidationErr{
			Field:   "abv",
			Message: "ABV must be between 0 and 100",
			Err:     ErrInvalidABV,
		}
	}

	return nil
}

func (c *Calculator) convertQuantity(quantity float64, fromUnit, toUnit UnitType) float64 {
	if fromUnit == toUnit {
		return quantity
	}

	// Future: implement unit conversions
	if c.logger != nil {
		c.logger.Warn("unit conversion not implemented",
			"from", fromUnit,
			"to", toUnit,
			"quantity", quantity)
	}

	return quantity
}

func roundToTwoDecimals(val float64) float64 {
	return float64(int(val*100+0.5)) / 100
}

func isValidProductType(pt ProductType) bool {
	switch pt {
	case ProductTypeBeer, ProductTypeWine, ProductTypeSpirits, ProductTypeOther:
		return true
	default:
		return false
	}
}

func isValidUnitType(ut UnitType) bool {
	switch ut {
	case UnitTypeGallon, UnitTypeBarrel, UnitTypeCase, UnitTypeLiter:
		return true
	default:
		return false
	}
}

// NewCalculationError creates a calculation error with context.
func NewCalculationError(message, itemName string, cause error) error {
	return &CalculationError{
		Message: message + " for item: " + itemName,
		Err:     cause,
	}
}
