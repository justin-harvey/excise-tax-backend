package tax

import (
	"os"
	"testing"
	"time"
)

func TestDefaultFederalRates(t *testing.T) {
	rates := DefaultFederalRates()

	if len(rates) == 0 {
		t.Fatal("DefaultFederalRates() returned empty slice")
	}

	// Verify all rates are valid
	for i, rate := range rates {
		if err := validateRate(&rate); err != nil {
			t.Errorf("DefaultFederalRates()[%d] invalid: %v", i, err)
		}
	}

	// Check for expected rates
	hasBeerBarrel := false
	hasBeerGallon := false
	hasWineGallon := false
	hasSpiritsGallon := false

	for _, rate := range rates {
		if rate.ProductType == ProductTypeBeer && rate.UnitType == UnitTypeBarrel {
			hasBeerBarrel = true
			if rate.RatePerUnit != 18.00 {
				t.Errorf("Beer barrel rate = %.2f, want 18.00", rate.RatePerUnit)
			}
		}
		if rate.ProductType == ProductTypeBeer && rate.UnitType == UnitTypeGallon {
			hasBeerGallon = true
			if rate.RatePerUnit != 0.58 {
				t.Errorf("Beer gallon rate = %.2f, want 0.58", rate.RatePerUnit)
			}
		}
		if rate.ProductType == ProductTypeWine && rate.UnitType == UnitTypeGallon {
			hasWineGallon = true
			if rate.RatePerUnit != 1.07 {
				t.Errorf("Wine gallon rate = %.2f, want 1.07", rate.RatePerUnit)
			}
		}
		if rate.ProductType == ProductTypeSpirits && rate.UnitType == UnitTypeGallon {
			hasSpiritsGallon = true
			if rate.RatePerUnit != 13.50 {
				t.Errorf("Spirits gallon rate = %.2f, want 13.50", rate.RatePerUnit)
			}
		}
	}

	if !hasBeerBarrel {
		t.Error("Missing beer barrel rate")
	}
	if !hasBeerGallon {
		t.Error("Missing beer gallon rate")
	}
	if !hasWineGallon {
		t.Error("Missing wine gallon rate")
	}
	if !hasSpiritsGallon {
		t.Error("Missing spirits gallon rate")
	}
}

func TestLoadRatesFromJSON(t *testing.T) {
	// Create temp file with valid rates
	validJSON := `[
		{
			"product_type": "beer",
			"rate_per_unit": 18.00,
			"unit_type": "barrel",
			"effective_date": "2018-01-01T00:00:00Z",
			"expiration_date": null,
			"description": "Federal beer tax",
			"jurisdiction": "federal"
		},
		{
			"product_type": "wine",
			"rate_per_unit": 1.07,
			"unit_type": "gallon",
			"effective_date": "2018-01-01T00:00:00Z",
			"expiration_date": null,
			"description": "Federal wine tax",
			"jurisdiction": "federal"
		}
	]`

	tmpFile, err := os.CreateTemp("", "rates-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(validJSON); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Test loading valid file
	rates, err := LoadRatesFromJSON(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadRatesFromJSON() error = %v", err)
	}

	if len(rates) != 2 {
		t.Errorf("LoadRatesFromJSON() loaded %d rates, want 2", len(rates))
	}

	// Verify first rate
	if rates[0].ProductType != ProductTypeBeer {
		t.Errorf("rates[0].ProductType = %s, want beer", rates[0].ProductType)
	}
	if rates[0].RatePerUnit != 18.00 {
		t.Errorf("rates[0].RatePerUnit = %.2f, want 18.00", rates[0].RatePerUnit)
	}
}

func TestLoadRatesFromJSON_InvalidFile(t *testing.T) {
	_, err := LoadRatesFromJSON("nonexistent.json")
	if err == nil {
		t.Error("LoadRatesFromJSON() with nonexistent file should return error")
	}
}

func TestLoadRatesFromJSON_InvalidJSON(t *testing.T) {
	invalidJSON := `not valid json`

	tmpFile, err := os.CreateTemp("", "rates-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidJSON); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	_, err = LoadRatesFromJSON(tmpFile.Name())
	if err == nil {
		t.Error("LoadRatesFromJSON() with invalid JSON should return error")
	}
}

func TestLoadRatesFromJSON_InvalidRate(t *testing.T) {
	invalidJSON := `[
		{
			"product_type": "invalid_type",
			"rate_per_unit": 18.00,
			"unit_type": "barrel",
			"effective_date": "2018-01-01T00:00:00Z",
			"jurisdiction": "federal"
		}
	]`

	tmpFile, err := os.CreateTemp("", "rates-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidJSON); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	_, err = LoadRatesFromJSON(tmpFile.Name())
	if err == nil {
		t.Error("LoadRatesFromJSON() with invalid rate should return error")
	}
}

func TestValidateRate(t *testing.T) {
	effectiveDate := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	expirationDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	pastExpiration := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		rate    TaxRate
		wantErr bool
	}{
		{
			name: "valid rate",
			rate: TaxRate{
				ProductType:    ProductTypeBeer,
				RatePerUnit:    18.00,
				UnitType:       UnitTypeBarrel,
				EffectiveDate:  effectiveDate,
				ExpirationDate: nil,
				Jurisdiction:   "federal",
			},
			wantErr: false,
		},
		{
			name: "valid rate with expiration",
			rate: TaxRate{
				ProductType:    ProductTypeBeer,
				RatePerUnit:    18.00,
				UnitType:       UnitTypeBarrel,
				EffectiveDate:  effectiveDate,
				ExpirationDate: &expirationDate,
				Jurisdiction:   "federal",
			},
			wantErr: false,
		},
		{
			name: "invalid product type",
			rate: TaxRate{
				ProductType:   "invalid",
				RatePerUnit:   18.00,
				UnitType:      UnitTypeBarrel,
				EffectiveDate: effectiveDate,
				Jurisdiction:  "federal",
			},
			wantErr: true,
		},
		{
			name: "invalid unit type",
			rate: TaxRate{
				ProductType:   ProductTypeBeer,
				RatePerUnit:   18.00,
				UnitType:      "invalid",
				EffectiveDate: effectiveDate,
				Jurisdiction:  "federal",
			},
			wantErr: true,
		},
		{
			name: "negative rate",
			rate: TaxRate{
				ProductType:   ProductTypeBeer,
				RatePerUnit:   -18.00,
				UnitType:      UnitTypeBarrel,
				EffectiveDate: effectiveDate,
				Jurisdiction:  "federal",
			},
			wantErr: true,
		},
		{
			name: "missing jurisdiction",
			rate: TaxRate{
				ProductType:   ProductTypeBeer,
				RatePerUnit:   18.00,
				UnitType:      UnitTypeBarrel,
				EffectiveDate: effectiveDate,
				Jurisdiction:  "",
			},
			wantErr: true,
		},
		{
			name: "expiration before effective",
			rate: TaxRate{
				ProductType:    ProductTypeBeer,
				RatePerUnit:    18.00,
				UnitType:       UnitTypeBarrel,
				EffectiveDate:  effectiveDate,
				ExpirationDate: &pastExpiration,
				Jurisdiction:   "federal",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRate(&tt.rate)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
