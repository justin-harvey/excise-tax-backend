package tax

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// DefaultFederalRates returns the standard US federal excise tax rates
// for alcoholic beverages as of 2024.
//
// These rates are based on the current federal excise tax schedule:
// - Beer: $18.00 per barrel (31 gallons)
// - Wine: $1.07 per gallon (for wines up to 14% ABV)
// - Spirits: $13.50 per gallon (80 proof and above)
//
// Reference: TTB (Alcohol and Tobacco Tax and Trade Bureau) current rates
func DefaultFederalRates() []TaxRate {
	return []TaxRate{
		{
			ID:              "federal-beer-barrel-2024",
			ProductType:     ProductTypeBeer,
			UnitType:        UnitTypeBarrel,
			Jurisdiction:    "federal",
			RatePerUnit:     18.00,
			EffectiveDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate:  nil,
			Description:     "Federal excise tax on beer per barrel (31 gallons)",
			RegulatoryBasis: "26 USC 5051",
		},
		{
			ID:              "federal-beer-gallon-2024",
			ProductType:     ProductTypeBeer,
			UnitType:        UnitTypeGallon,
			Jurisdiction:    "federal",
			RatePerUnit:     0.58, // $18.00 / 31 gallons per barrel
			EffectiveDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate:  nil,
			Description:     "Federal excise tax on beer per gallon",
			RegulatoryBasis: "26 USC 5051",
		},
		{
			ID:              "federal-wine-gallon-2024",
			ProductType:     ProductTypeWine,
			UnitType:        UnitTypeGallon,
			Jurisdiction:    "federal",
			RatePerUnit:     1.07,
			EffectiveDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate:  nil,
			Description:     "Federal excise tax on wine per gallon (14% ABV or less)",
			RegulatoryBasis: "26 USC 5041",
		},
		{
			ID:              "federal-spirits-gallon-2024",
			ProductType:     ProductTypeSpirits,
			UnitType:        UnitTypeGallon,
			Jurisdiction:    "federal",
			RatePerUnit:     13.50,
			EffectiveDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate:  nil,
			Description:     "Federal excise tax on distilled spirits per gallon (80+ proof)",
			RegulatoryBasis: "26 USC 5001",
		},
	}
}

// LoadRatesFromJSON loads tax rates from a JSON file.
// The file should contain an array of TaxRate objects.
//
// Example JSON format:
//
//	[
//	  {
//	    "id": "federal-beer-barrel-2024",
//	    "product_type": "beer",
//	    "unit_type": "barrel",
//	    "jurisdiction": "federal",
//	    "rate_per_unit": 18.00,
//	    "effective_date": "2024-01-01T00:00:00Z",
//	    "description": "Federal excise tax on beer per barrel",
//	    "regulatory_basis": "26 USC 5051"
//	  }
//	]
func LoadRatesFromJSON(filename string) ([]TaxRate, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read rates file %s: %w", filename, err)
	}

	var rates []TaxRate
	if err := json.Unmarshal(data, &rates); err != nil {
		return nil, fmt.Errorf("failed to parse rates JSON from %s: %w", filename, err)
	}

	// Validate loaded rates
	for i, rate := range rates {
		if err := validateTaxRate(&rate); err != nil {
			return nil, fmt.Errorf("invalid rate at index %d: %w", i, err)
		}
	}

	return rates, nil
}

// SaveRatesToJSON saves tax rates to a JSON file.
// The file will contain a formatted JSON array of TaxRate objects.
func SaveRatesToJSON(rates []TaxRate, filename string) error {
	data, err := json.MarshalIndent(rates, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rates to JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write rates file %s: %w", filename, err)
	}

	return nil
}

// validateTaxRate validates a tax rate for correctness
func validateTaxRate(rate *TaxRate) error {
	if rate.ID == "" {
		return &ValidationErr{
			Field:   "id",
			Message: "rate ID is required",
			Err:     ErrInvalidProductType,
		}
	}

	if !isValidProductType(rate.ProductType) {
		return &ValidationErr{
			Field:   "product_type",
			Message: "invalid product type: " + string(rate.ProductType),
			Err:     ErrInvalidProductType,
		}
	}

	if !isValidUnitType(rate.UnitType) {
		return &ValidationErr{
			Field:   "unit_type",
			Message: "invalid unit type: " + string(rate.UnitType),
			Err:     ErrInvalidUnitType,
		}
	}

	if rate.Jurisdiction == "" {
		return &ValidationErr{
			Field:   "jurisdiction",
			Message: "jurisdiction is required",
			Err:     ErrInvalidJurisdiction,
		}
	}

	if rate.RatePerUnit < 0 {
		return &ValidationErr{
			Field:   "rate_per_unit",
			Message: "rate per unit cannot be negative",
			Err:     ErrInvalidQuantity,
		}
	}

	// Check expiration date is after effective date
	if rate.ExpirationDate != nil && rate.ExpirationDate.Before(rate.EffectiveDate) {
		return &ValidationErr{
			Field:   "expiration_date",
			Message: "expiration date must be after effective date",
			Err:     ErrInvalidQuantity,
		}
	}

	return nil
}
