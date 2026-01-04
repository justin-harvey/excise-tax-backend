package tax_test

import (
	"os"
	"testing"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/pkg/tax"
)

func TestDefaultFederalRates(t *testing.T) {
	rates := tax.DefaultFederalRates()

	if len(rates) == 0 {
		t.Error("DefaultFederalRates() returned empty slice")
	}

	// Check that we have the expected federal rates
	expectedRates := map[string]struct {
		productType tax.ProductType
		unitType    tax.UnitType
		rate        float64
	}{
		"federal-beer-barrel-2024":    {tax.ProductTypeBeer, tax.UnitTypeBarrel, 18.00},
		"federal-beer-gallon-2024":    {tax.ProductTypeBeer, tax.UnitTypeGallon, 0.58},
		"federal-wine-gallon-2024":    {tax.ProductTypeWine, tax.UnitTypeGallon, 1.07},
		"federal-spirits-gallon-2024": {tax.ProductTypeSpirits, tax.UnitTypeGallon, 13.50},
	}

	foundRates := make(map[string]bool)
	for _, rate := range rates {
		expected, exists := expectedRates[rate.ID]
		if !exists {
			t.Errorf("unexpected rate ID: %s", rate.ID)
			continue
		}

		if rate.ProductType != expected.productType {
			t.Errorf("rate %s: product type = %v, want %v", rate.ID, rate.ProductType, expected.productType)
		}

		if rate.UnitType != expected.unitType {
			t.Errorf("rate %s: unit type = %v, want %v", rate.ID, rate.UnitType, expected.unitType)
		}

		if rate.RatePerUnit != expected.rate {
			t.Errorf("rate %s: rate = %v, want %v", rate.ID, rate.RatePerUnit, expected.rate)
		}

		if rate.Jurisdiction != "federal" {
			t.Errorf("rate %s: jurisdiction = %v, want federal", rate.ID, rate.Jurisdiction)
		}

		foundRates[rate.ID] = true
	}

	// Verify all expected rates were found
	for id := range expectedRates {
		if !foundRates[id] {
			t.Errorf("missing expected rate: %s", id)
		}
	}
}

func TestLoadRatesFromJSON(t *testing.T) {
	// Create a temporary file with test rates
	tempFile, err := os.CreateTemp("", "test-rates-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test JSON data
	testJSON := `[
		{
			"id": "test-beer-barrel",
			"product_type": "beer",
			"unit_type": "barrel",
			"jurisdiction": "test",
			"rate_per_unit": 15.00,
			"effective_date": "2024-01-01T00:00:00Z",
			"description": "Test beer rate",
			"regulatory_basis": "Test statute"
		},
		{
			"id": "test-wine-gallon",
			"product_type": "wine",
			"unit_type": "gallon",
			"jurisdiction": "test",
			"rate_per_unit": 2.50,
			"effective_date": "2024-01-01T00:00:00Z",
			"expiration_date": "2025-01-01T00:00:00Z",
			"description": "Test wine rate",
			"regulatory_basis": "Test regulation"
		}
	]`

	if err := os.WriteFile(tempFile.Name(), []byte(testJSON), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test loading rates
	rates, err := tax.LoadRatesFromJSON(tempFile.Name())
	if err != nil {
		t.Errorf("LoadRatesFromJSON() error = %v", err)
	}

	if len(rates) != 2 {
		t.Errorf("LoadRatesFromJSON() loaded %d rates, want 2", len(rates))
	}

	// Verify first rate
	if rates[0].ID != "test-beer-barrel" {
		t.Errorf("rate 0 ID = %v, want test-beer-barrel", rates[0].ID)
	}
	if rates[0].ProductType != tax.ProductTypeBeer {
		t.Errorf("rate 0 product type = %v, want beer", rates[0].ProductType)
	}
	if rates[0].RatePerUnit != 15.00 {
		t.Errorf("rate 0 rate = %v, want 15.00", rates[0].RatePerUnit)
	}

	// Verify second rate (with expiration date)
	if rates[1].ID != "test-wine-gallon" {
		t.Errorf("rate 1 ID = %v, want test-wine-gallon", rates[1].ID)
	}
	if rates[1].ExpirationDate == nil {
		t.Error("rate 1 should have expiration date")
	}
}

func TestLoadRatesFromJSON_InvalidFile(t *testing.T) {
	// Test nonexistent file
	_, err := tax.LoadRatesFromJSON("nonexistent-file.json")
	if err == nil {
		t.Error("LoadRatesFromJSON() should error for nonexistent file")
	}

	// Test invalid JSON
	tempFile, err := os.CreateTemp("", "invalid-rates-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if err := os.WriteFile(tempFile.Name(), []byte("invalid json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = tax.LoadRatesFromJSON(tempFile.Name())
	if err == nil {
		t.Error("LoadRatesFromJSON() should error for invalid JSON")
	}
}

func TestSaveRatesToJSON(t *testing.T) {
	rates := []tax.TaxRate{
		{
			ID:              "test-save-beer",
			ProductType:     tax.ProductTypeBeer,
			UnitType:        tax.UnitTypeBarrel,
			Jurisdiction:    "test",
			RatePerUnit:     20.00,
			EffectiveDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Description:     "Test save rate",
			RegulatoryBasis: "Test code",
		},
	}

	tempFile, err := os.CreateTemp("", "save-rates-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Save rates to file
	err = tax.SaveRatesToJSON(rates, tempFile.Name())
	if err != nil {
		t.Errorf("SaveRatesToJSON() error = %v", err)
	}

	// Load them back and verify
	loadedRates, err := tax.LoadRatesFromJSON(tempFile.Name())
	if err != nil {
		t.Errorf("failed to load saved rates: %v", err)
	}

	if len(loadedRates) != 1 {
		t.Errorf("loaded %d rates, want 1", len(loadedRates))
	}

	if loadedRates[0].ID != "test-save-beer" {
		t.Errorf("loaded rate ID = %v, want test-save-beer", loadedRates[0].ID)
	}

	if loadedRates[0].RatePerUnit != 20.00 {
		t.Errorf("loaded rate = %v, want 20.00", loadedRates[0].RatePerUnit)
	}
}
