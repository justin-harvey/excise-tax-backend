package tax_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/pkg/tax"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []tax.Option
		wantErr bool
	}{
		{
			name: "default calculator",
			opts: nil,
		},
		{
			name: "with custom logger",
			opts: []tax.Option{
				tax.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),
			},
		},
		{
			name: "with custom rates",
			opts: []tax.Option{
				tax.WithRates([]tax.TaxRate{
					{
						ID:              "test-beer-gallon",
						ProductType:     tax.ProductTypeBeer,
						UnitType:        tax.UnitTypeGallon,
						Jurisdiction:    "test",
						RatePerUnit:     1.00,
						EffectiveDate:   time.Now().AddDate(-1, 0, 0),
						Description:     "Test rate",
						RegulatoryBasis: "Test",
					},
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc, err := tax.New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if calc == nil {
				t.Error("New() returned nil calculator")
			}
		})
	}
}

func TestCalculator_Calculate(t *testing.T) {
	calc, err := tax.New()
	if err != nil {
		t.Fatalf("failed to create calculator: %v", err)
	}

	tests := []struct {
		name             string
		data             *tax.ProductionData
		jurisdiction     string
		effectiveDate    time.Time
		wantErr          bool
		wantTotalTax     float64
		wantNumBreakdown int
	}{
		{
			name: "single beer item",
			data: &tax.ProductionData{
				Items: []tax.ProductionItem{
					{
						ProductType: tax.ProductTypeBeer,
						ProductName: "Test IPA",
						UnitType:    tax.UnitTypeBarrel,
						Quantity:    10,
						ABV:         6.5,
					},
				},
			},
			jurisdiction:     "federal",
			effectiveDate:    time.Now(),
			wantErr:          false,
			wantTotalTax:     180.0,
			wantNumBreakdown: 1,
		},
		{
			name:          "nil production data",
			data:          nil,
			jurisdiction:  "federal",
			effectiveDate: time.Now(),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.Calculate(context.Background(), tt.data, tt.jurisdiction, tt.effectiveDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if result.TotalTaxAmount != tt.wantTotalTax {
				t.Errorf("Calculate() total tax = %v, want %v", result.TotalTaxAmount, tt.wantTotalTax)
			}

			if len(result.BreakdownByProduct) != tt.wantNumBreakdown {
				t.Errorf("Calculate() breakdown count = %v, want %v", len(result.BreakdownByProduct), tt.wantNumBreakdown)
			}
		})
	}
}

func Example() {
	calc, err := tax.New()
	if err != nil {
		panic(err)
	}

	data := &tax.ProductionData{
		Items: []tax.ProductionItem{
			{
				ProductType: tax.ProductTypeBeer,
				ProductName: "Craft IPA",
				UnitType:    tax.UnitTypeBarrel,
				Quantity:    50,
				ABV:         6.5,
			},
		},
	}

	result, err := calc.Calculate(context.Background(), data, "federal", time.Now())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Total tax: $%.2f\n", result.TotalTaxAmount)
	fmt.Printf("Number of products: %d\n", len(result.BreakdownByProduct))

	// Output:
	// Total tax: $900.00
	// Number of products: 1
}
