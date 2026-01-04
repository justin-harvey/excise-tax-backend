package tax

import (
	"errors"
	"fmt"
)

// Sentinel errors for common conditions.
var (
	// ErrInvalidProductType indicates an unsupported product type.
	ErrInvalidProductType = errors.New("invalid product type")

	// ErrInvalidUnitType indicates an unsupported unit type.
	ErrInvalidUnitType = errors.New("invalid unit type")

	// ErrInvalidQuantity indicates an invalid quantity value.
	ErrInvalidQuantity = errors.New("invalid quantity: must be greater than zero")

	// ErrInvalidABV indicates an invalid ABV value.
	ErrInvalidABV = errors.New("invalid ABV: must be between 0 and 100")

	// ErrNoTaxRate indicates no applicable tax rate was found.
	ErrNoTaxRate = errors.New("no applicable tax rate found")

	// ErrInvalidDate indicates an invalid date.
	ErrInvalidDate = errors.New("invalid date")

	// ErrNoProductionItems indicates no production items provided.
	ErrNoProductionItems = errors.New("no production items provided")
)

// ValidationError represents a validation error.
type ValidationErr struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationErr) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ValidationErr) Unwrap() error {
	return e.Err
}

// CalculationError represents an error during tax calculation.
type CalculationError struct {
	Message string
	Err     error
}

func (e *CalculationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("calculation error: %s (%v)", e.Message, e.Err)
	}
	return fmt.Sprintf("calculation error: %s", e.Message)
}

func (e *CalculationError) Unwrap() error {
	return e.Err
}
