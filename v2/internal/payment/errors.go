package payment

import (
	"errors"
	"fmt"
)

// Sentinel errors for common payment conditions
var (
	ErrInvalidPaymentType   = errors.New("invalid payment type")
	ErrInvalidState         = errors.New("invalid transaction state")
	ErrInvalidAmount        = errors.New("invalid payment amount")
	ErrInvalidAddress       = errors.New("invalid address")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentAlreadyExists = errors.New("payment already exists")
	ErrInvalidTransition    = errors.New("invalid state transition")
	ErrEventImmutable       = errors.New("events are immutable and cannot be modified")
	ErrEmptyEvents          = errors.New("payment must have at least one event")
	ErrXRPLNetwork          = errors.New("XRPL network error")
	ErrPaymentProcessing    = errors.New("payment processing error")
)

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NetworkError represents a network-related error
type NetworkError struct {
	Operation string
	Err       error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error during %s: %v", e.Operation, e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// ProcessingError represents a payment processing error
type ProcessingError struct {
	PaymentID string
	Stage     string
	Err       error
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("processing error for payment %s at stage %s: %v", e.PaymentID, e.Stage, e.Err)
}

func (e *ProcessingError) Unwrap() error {
	return e.Err
}
