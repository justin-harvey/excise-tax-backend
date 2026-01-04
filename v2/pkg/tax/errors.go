package tax

import (
	"errors"
	"fmt"
)

// Sentinel errors for common tax calculation issues
var (
	// ErrNoProductionItems indicates no production items were provided
	ErrNoProductionItems = errors.New("no production items provided")

	// ErrNoTaxRate indicates no applicable tax rate was found
	ErrNoTaxRate = errors.New("no applicable tax rate found")

	// ErrInvalidProductType indicates an unsupported product type
	ErrInvalidProductType = errors.New("invalid product type")

	// ErrInvalidUnitType indicates an unsupported unit type
	ErrInvalidUnitType = errors.New("invalid unit type")

	// ErrInvalidQuantity indicates an invalid quantity value
	ErrInvalidQuantity = errors.New("invalid quantity")

	// ErrInvalidABV indicates an invalid alcohol by volume value
	ErrInvalidABV = errors.New("invalid ABV value")

	// ErrInvalidJurisdiction indicates an unsupported jurisdiction
	ErrInvalidJurisdiction = errors.New("invalid jurisdiction")

	// ErrRateLoadFailed indicates tax rate loading failed
	ErrRateLoadFailed = errors.New("failed to load tax rates")
)

// ValidationErr represents a validation error with additional context
type ValidationErr struct {
	Field   string
	Message string
	Err     error
}

// Error implements the error interface
func (v *ValidationErr) Error() string {
	return fmt.Sprintf("validation error in field %s: %s", v.Field, v.Message)
}

// Unwrap returns the underlying error
func (v *ValidationErr) Unwrap() error {
	return v.Err
}

// CalculationError represents an error during tax calculation
type CalculationError struct {
	Message string
	Err     error
}

// Error implements the error interface
func (c *CalculationError) Error() string {
	return c.Message
}

// Unwrap returns the underlying error
func (c *CalculationError) Unwrap() error {
	return c.Err
}
