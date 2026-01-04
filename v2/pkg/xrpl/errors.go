package xrpl

import "errors"

// Sentinel errors for common error conditions.
// Use errors.Is() to check for these specific errors.
var (
	// ErrNotConnected indicates the client is not connected to XRPL
	ErrNotConnected = errors.New("xrpl: not connected")

	// ErrInvalidAddress indicates an invalid XRPL address format
	ErrInvalidAddress = errors.New("xrpl: invalid address format")

	// ErrInvalidHash indicates an invalid transaction hash format
	ErrInvalidHash = errors.New("xrpl: invalid transaction hash format")

	// ErrTimeout indicates a request timeout
	ErrTimeout = errors.New("xrpl: request timeout")

	// ErrRequestFailed indicates the XRPL server rejected the request
	ErrRequestFailed = errors.New("xrpl: request failed")

	// ErrInvalidResponse indicates the server response was in an unexpected format
	ErrInvalidResponse = errors.New("xrpl: invalid response format")

	// ErrAlreadyConnected indicates the client is already connected
	ErrAlreadyConnected = errors.New("xrpl: already connected")

	// ErrConnectionClosed indicates the connection was closed
	ErrConnectionClosed = errors.New("xrpl: connection closed")
)
