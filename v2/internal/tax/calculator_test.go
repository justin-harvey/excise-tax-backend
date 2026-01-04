package tax

import (
	"context"
	"testing"
	"time"
)

func TestCalculator_Calculate(t *testing.T) {
	rates := DefaultFederalRates()
	calc := NewCalculator(rates)
	ctx := context.Background()
	effectiveDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		data         *ProductionData
		jurisdiction string
		wantTotalTax float64
		wantErr      bool
		wantErrType  error
	}{
		{
			name: "single beer barrel",
			data: &ProductionData{
				Items: []ProductionItem{
					{
						ProductType: ProductTypeBeer,
						ProductName: "IPA",
						UnitType:    UnitTypeBarrel,
						Quantity:    10,
						ABV:         6.5,
					},
				},
			},
			jurisdiction: "federal",
			wantTotalTax: 180.00, // 10 barrels * $18/barrel
			wantErr:      false,
		},
		{
			name: "multiple products",
			data: &ProductionData{
				Items: []ProductionItem{
					{
						ProductType: ProductTypeBeer,
						ProductName: "IPA",
						UnitType:    UnitTypeBarrel,
						Quantity:    5,
						ABV:         6.5,
					},
					{
						ProductType: ProductTypeWine,
						ProductName: "Chardonnay",
						UnitType:    UnitTypeGallon,
						Quantity:    100,
						ABV:         12.5,
					},
				},
			},
			jurisdiction: "federal",
			wantTotalTax: 197.00, // (5 * $18) + (100 * $1.07)
			wantErr:      false,
		},
		{
			name:         "nil production data",
			data:         nil,
			jurisdiction: "federal",
			wantErr:      true,
			wantErrType:  ErrNoProductionItems,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.Calculate(ctx, tt.data, tt.jurisdiction, effectiveDate)

			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrType != nil {
				if !containsError(err, tt.wantErrType) {
					t.Errorf("Calculate() error = %v, want error type %v", err, tt.wantErrType)
				}
				return
			}

			if !tt.wantErr && result != nil {
				if result.TotalTaxAmount != tt.wantTotalTax {
					t.Errorf("Calculate() total tax = %.2f, want %.2f", result.TotalTaxAmount, tt.wantTotalTax)
				}
			}
		})
	}
}

func TestCalculator_GetRate(t *testing.T) {
	rates := DefaultFederalRates()
	calc := NewCalculator(rates)
	effectiveDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	rate, err := calc.GetRate(ProductTypeBeer, UnitTypeBarrel, "federal", effectiveDate)
	if err != nil {
		t.Errorf("GetRate() error = %v", err)
	}
	if rate.RatePerUnit != 18.00 {
		t.Errorf("GetRate() rate = %.2f, want 18.00", rate.RatePerUnit)
	}
}

// containsError checks if err contains the target error in its chain
func containsError(err, target error) bool {
	if err == nil {
		return false
	}
	if err == target {
		return true
	}
	// Check unwrapping
	type unwrapper interface {
		Unwrap() error
	}
	if u, ok := err.(unwrapper); ok {
		return containsError(u.Unwrap(), target)
	}
	return false
}
