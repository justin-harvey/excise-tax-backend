# Tax Package

A Go library for calculating excise taxes on alcoholic beverages.

## Overview

The tax package provides a clean, functional API for calculating excise taxes on beer, wine, spirits, and other alcoholic beverages. It supports multiple jurisdictions, unit types, and time-based rate lookup.

## Features

- **Pure Functions**: All calculations are side-effect free
- **Multiple Jurisdictions**: Support for federal, state, and local tax rates
- **Flexible Units**: Handles gallons, barrels, cases, and liters
- **Time-Based Rates**: Effective and expiration date support
- **Comprehensive Validation**: Input validation with detailed error messages
- **Concurrent Safe**: All operations are safe for concurrent use
- **Functional Options**: Clean configuration using functional options pattern

## Installation

```bash
go get github.com/maxfelker/excise-tax-backend/v2/pkg/tax
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/maxfelker/excise-tax-backend/v2/pkg/tax"
)

func main() {
    // Create calculator with default federal rates
    calc, err := tax.New()
    if err != nil {
        log.Fatal(err)
    }

    // Define production data
    data := &tax.ProductionData{
        Items: []tax.ProductionItem{
            {
                ProductType: tax.ProductTypeBeer,
                ProductName: "Craft IPA",
                UnitType:    tax.UnitTypeBarrel,
                Quantity:    10,
                ABV:         6.5,
            },
        },
    }

    // Calculate taxes
    result, err := calc.Calculate(context.Background(), data, "federal", time.Now())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Total tax: $%.2f\n", result.TotalTaxAmount)
    // Output: Total tax: $180.00
}
```

### Custom Configuration

```go
// Custom logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Custom rates
customRates := []tax.TaxRate{
    {
        ID:              "custom-beer-barrel",
        ProductType:     tax.ProductTypeBeer,
        UnitType:        tax.UnitTypeBarrel,
        Jurisdiction:    "custom",
        RatePerUnit:     20.00,
        EffectiveDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
        Description:     "Custom beer rate",
        RegulatoryBasis: "Custom law",
    },
}

// Create calculator with custom configuration
calc, err := tax.New(
    tax.WithLogger(logger),
    tax.WithRates(customRates),
)
```

### Loading Rates from File

```go
// Create calculator with rates from JSON file
calc, err := tax.New(
    tax.WithRatesFromFile("rates.json"),
)
```

Example rates.json:
```json
[
  {
    "id": "state-beer-barrel-2024",
    "product_type": "beer",
    "unit_type": "barrel",
    "jurisdiction": "california",
    "rate_per_unit": 20.00,
    "effective_date": "2024-01-01T00:00:00Z",
    "description": "California beer excise tax",
    "regulatory_basis": "CA Revenue Code 123"
  }
]
```

## API Reference

### Types

#### ProductType
```go
const (
    ProductTypeBeer    ProductType = "beer"
    ProductTypeWine    ProductType = "wine"  
    ProductTypeSpirits ProductType = "spirits"
    ProductTypeOther   ProductType = "other"
)
```

#### UnitType
```go
const (
    UnitTypeGallon UnitType = "gallon"
    UnitTypeBarrel UnitType = "barrel"
    UnitTypeCase   UnitType = "case"
    UnitTypeLiter  UnitType = "liter"
)
```

#### ProductionItem
Represents a single item in a production batch:
```go
type ProductionItem struct {
    ProductType ProductType `json:"product_type"`
    ProductName string      `json:"product_name"`
    UnitType    UnitType    `json:"unit_type"`
    Quantity    float64     `json:"quantity"`
    ABV         float64     `json:"abv"` // Alcohol by Volume percentage
}
```

#### TaxRate
Defines tax rate for a specific product and jurisdiction:
```go
type TaxRate struct {
    ID              string      `json:"id"`
    ProductType     ProductType `json:"product_type"`
    UnitType        UnitType    `json:"unit_type"`
    Jurisdiction    string      `json:"jurisdiction"`
    RatePerUnit     float64     `json:"rate_per_unit"`
    EffectiveDate   time.Time   `json:"effective_date"`
    ExpirationDate  *time.Time  `json:"expiration_date,omitempty"`
    Description     string      `json:"description"`
    RegulatoryBasis string      `json:"regulatory_basis"`
}
```

### Calculator Methods

#### Calculate
```go
func (c *Calculator) Calculate(ctx context.Context, data *ProductionData, jurisdiction string, effectiveDate time.Time) (*TaxCalculation, error)
```
Calculates tax for production data with detailed breakdown.

#### CalculateForItem
```go
func (c *Calculator) CalculateForItem(ctx context.Context, item ProductionItem, jurisdiction string, effectiveDate time.Time) (float64, error)
```
Convenience method for calculating tax on a single item.

#### GetRate
```go
func (c *Calculator) GetRate(productType ProductType, unitType UnitType, jurisdiction string, effectiveDate time.Time) (*TaxRate, error)
```
Retrieves applicable tax rate for given criteria.

#### GetAllRates
```go
func (c *Calculator) GetAllRates() []TaxRate
```
Returns all loaded tax rates (copy to prevent modification).

#### ValidateProductionData
```go
func (c *Calculator) ValidateProductionData(data *ProductionData) []ValidationError
```
Validates production data without performing calculations.

### Default Federal Rates

The package includes current US federal excise tax rates:

- **Beer**: $18.00/barrel ($0.58/gallon)
- **Wine**: $1.07/gallon (≤14% ABV)
- **Spirits**: $13.50/gallon (≥80 proof)

These rates are based on TTB (Alcohol and Tobacco Tax and Trade Bureau) current schedules.

### Error Handling

The package provides detailed error information:

```go
// Sentinel errors
var (
    ErrNoProductionItems  = errors.New("no production items provided")
    ErrNoTaxRate         = errors.New("no applicable tax rate found")
    ErrInvalidProductType = errors.New("invalid product type")
    // ... more errors
)

// Structured validation errors
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// Calculation errors with context
type CalculationError struct {
    Message string
    Err     error
}
```

## Testing

Run tests:
```bash
go test ./pkg/tax/...
```

Run tests with coverage:
```bash
go test -cover ./pkg/tax/...
```

## Federal Tax Rate References

- [TTB Tax Rates](https://www.ttb.gov/tax-rates)
- [26 USC 5051](https://www.law.cornell.edu/uscode/text/26/5051) - Beer taxes
- [26 USC 5041](https://www.law.cornell.edu/uscode/text/26/5041) - Wine taxes  
- [26 USC 5001](https://www.law.cornell.edu/uscode/text/26/5001) - Distilled spirits taxes

## License

This package is part of the excise-tax-backend project and follows the same license terms.