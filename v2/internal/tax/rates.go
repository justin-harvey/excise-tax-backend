package tax

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LoadRatesFromJSON loads tax rates from a JSON file.
func LoadRatesFromJSON(filename string) ([]TaxRate, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read rates file: %w", err)
	}

	var rates []TaxRate
	if err := json.Unmarshal(data, &rates); err != nil {
		return nil, fmt.Errorf("failed to parse rates JSON: %w", err)
	}

	// Validate rates
	for i, rate := range rates {
		if err := validateRate(&rate); err != nil {
			return nil, fmt.Errorf("invalid rate at index %d: %w", i, err)
		}
	}

	return rates, nil
}

// DefaultFederalRates returns default federal excise tax rates.
// These are example rates and should be updated to match actual federal rates.
func DefaultFederalRates() []TaxRate {
	return []TaxRate{
		{
			ProductType:    ProductTypeBeer,
			RatePerUnit:    18.00, // $18 per barrel (31 gallons)
			UnitType:       UnitTypeBarrel,
			EffectiveDate:  time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate: nil,
			Description:    "Federal excise tax on beer (standard rate for breweries producing over 2M barrels/year)",
			Jurisdiction:   "federal",
		},
		{
			ProductType:    ProductTypeWine,
			RatePerUnit:    1.07, // $1.07 per gallon for still wine up to 14% ABV
			UnitType:       UnitTypeGallon,
			EffectiveDate:  time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate: nil,
			Description:    "Federal excise tax on wine (still wine, up to 14% ABV)",
			Jurisdiction:   "federal",
		},
		{
			ProductType:    ProductTypeSpirits,
			RatePerUnit:    13.50, // $13.50 per proof gallon
			UnitType:       UnitTypeGallon,
			EffectiveDate:  time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate: nil,
			Description:    "Federal excise tax on distilled spirits",
			Jurisdiction:   "federal",
		},
		{
			ProductType:    ProductTypeBeer,
			RatePerUnit:    0.58, // $0.58 per gallon (converted from barrel rate)
			UnitType:       UnitTypeGallon,
			EffectiveDate:  time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpirationDate: nil,
			Description:    "Federal excise tax on beer (per gallon)",
			Jurisdiction:   "federal",
		},
	}
}

// validateRate validates a tax rate.
func validateRate(rate *TaxRate) error {
	if !isValidProductType(rate.ProductType) {
		return fmt.Errorf("invalid product type: %s", rate.ProductType)
	}

	if !isValidUnitType(rate.UnitType) {
		return fmt.Errorf("invalid unit type: %s", rate.UnitType)
	}

	if rate.RatePerUnit < 0 {
		return fmt.Errorf("rate per unit must be non-negative")
	}

	if rate.Jurisdiction == "" {
		return fmt.Errorf("jurisdiction is required")
	}

	if rate.ExpirationDate != nil && rate.ExpirationDate.Before(rate.EffectiveDate) {
		return fmt.Errorf("expiration date must be after effective date")
	}

	return nil
}
