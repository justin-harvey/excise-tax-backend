package tax

import (
	"fmt"
	"time"
)

// ProductType represents the type of alcoholic beverage
type ProductType string

const (
	ProductTypeBeer    ProductType = "beer"
	ProductTypeWine    ProductType = "wine"
	ProductTypeSpirits ProductType = "spirits"
	ProductTypeOther   ProductType = "other"
)

// UnitType represents the unit of measurement for production quantities
type UnitType string

const (
	UnitTypeGallon UnitType = "gallon"
	UnitTypeBarrel UnitType = "barrel"
	UnitTypeCase   UnitType = "case"
	UnitTypeLiter  UnitType = "liter"
)

// TaxRate represents a tax rate for a specific product type and jurisdiction
type TaxRate struct {
	ID              string      `json:"id" yaml:"id"`
	ProductType     ProductType `json:"product_type" yaml:"product_type"`
	UnitType        UnitType    `json:"unit_type" yaml:"unit_type"`
	Jurisdiction    string      `json:"jurisdiction" yaml:"jurisdiction"`
	RatePerUnit     float64     `json:"rate_per_unit" yaml:"rate_per_unit"`
	EffectiveDate   time.Time   `json:"effective_date" yaml:"effective_date"`
	ExpirationDate  *time.Time  `json:"expiration_date,omitempty" yaml:"expiration_date,omitempty"`
	Description     string      `json:"description" yaml:"description"`
	RegulatoryBasis string      `json:"regulatory_basis" yaml:"regulatory_basis"`
}

// ProductionItem represents a single item in a production batch
type ProductionItem struct {
	ProductType ProductType `json:"product_type" yaml:"product_type"`
	ProductName string      `json:"product_name" yaml:"product_name"`
	UnitType    UnitType    `json:"unit_type" yaml:"unit_type"`
	Quantity    float64     `json:"quantity" yaml:"quantity"`
	ABV         float64     `json:"abv" yaml:"abv"` // Alcohol by Volume percentage
}

// ProductionData represents data about a production batch for tax calculation
type ProductionData struct {
	Items []ProductionItem `json:"items" yaml:"items"`
}

// TaxCalculation represents the result of a tax calculation
type TaxCalculation struct {
	TotalTaxAmount     float64                 `json:"total_tax_amount" yaml:"total_tax_amount"`
	BreakdownByProduct []TaxCalculationProduct `json:"breakdown_by_product" yaml:"breakdown_by_product"`
	CalculatedAt       time.Time               `json:"calculated_at" yaml:"calculated_at"`
	Jurisdiction       string                  `json:"jurisdiction" yaml:"jurisdiction"`
}

// TaxCalculationProduct represents tax calculation results for a specific product
type TaxCalculationProduct struct {
	ProductType ProductType `json:"product_type" yaml:"product_type"`
	ProductName string      `json:"product_name" yaml:"product_name"`
	Quantity    float64     `json:"quantity" yaml:"quantity"`
	UnitType    UnitType    `json:"unit_type" yaml:"unit_type"`
	RatePerUnit float64     `json:"rate_per_unit" yaml:"rate_per_unit"`
	TaxAmount   float64     `json:"tax_amount" yaml:"tax_amount"`
}

// ValidationError represents a validation error with field context
type ValidationError struct {
	Field   string `json:"field" yaml:"field"`
	Message string `json:"message" yaml:"message"`
}

// Error implements the error interface
func (v ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %s: %s", v.Field, v.Message)
}
