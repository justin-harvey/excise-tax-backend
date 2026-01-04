// Package tax provides tax calculation and excise tax operations.
package tax

import (
	"time"
)

// ProductType represents the type of alcoholic beverage.
type ProductType string

const (
	ProductTypeBeer    ProductType = "beer"
	ProductTypeWine    ProductType = "wine"
	ProductTypeSpirits ProductType = "spirits"
	ProductTypeOther   ProductType = "other"
)

// UnitType represents the unit of measurement.
type UnitType string

const (
	UnitTypeGallon UnitType = "gallon"
	UnitTypeBarrel UnitType = "barrel"
	UnitTypeCase   UnitType = "case"
	UnitTypeLiter  UnitType = "liter"
)

// TaxRate represents an excise tax rate for a product type.
type TaxRate struct {
	ProductType    ProductType `json:"product_type"`
	RatePerUnit    float64     `json:"rate_per_unit"`   // Tax rate per unit
	UnitType       UnitType    `json:"unit_type"`       // Unit of measurement
	EffectiveDate  time.Time   `json:"effective_date"`  // When this rate becomes effective
	ExpirationDate *time.Time  `json:"expiration_date"` // When this rate expires (nil = no expiration)
	Description    string      `json:"description"`     // Human-readable description
	Jurisdiction   string      `json:"jurisdiction"`    // e.g., "federal", "california", "oregon"
}

// ProductionItem represents a single production line item.
type ProductionItem struct {
	ProductType ProductType `json:"product_type"` // Type of product (beer, wine, spirits)
	ProductName string      `json:"product_name"` // Name of the product
	UnitType    UnitType    `json:"unit_type"`    // Unit of measurement
	Quantity    float64     `json:"quantity"`     // Quantity produced
	ABV         float64     `json:"abv"`          // Alcohol by volume (percentage)
	Notes       string      `json:"notes"`        // Optional notes
}

// TaxCalculation represents the result of a tax calculation.
type TaxCalculation struct {
	TotalTaxAmount     float64                 `json:"total_tax_amount"`
	BreakdownByProduct []TaxCalculationProduct `json:"breakdown_by_product"`
	CalculatedAt       time.Time               `json:"calculated_at"`
	Jurisdiction       string                  `json:"jurisdiction"`
}

// TaxCalculationProduct represents tax calculation for a single product.
type TaxCalculationProduct struct {
	ProductType ProductType `json:"product_type"`
	ProductName string      `json:"product_name"`
	Quantity    float64     `json:"quantity"`
	UnitType    UnitType    `json:"unit_type"`
	RatePerUnit float64     `json:"rate_per_unit"`
	TaxAmount   float64     `json:"tax_amount"`
}

// ProductionData represents a collection of production items.
type ProductionData struct {
	Items []ProductionItem `json:"items"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult represents the result of validation.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}
