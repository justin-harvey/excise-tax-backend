// Package errors provides custom error types with HTTP status code mapping
// for consistent error handling across the application.
//
// Example usage:
//
//	// Create a not found error
//	err := errors.NewNotFoundError("user", "user not found with ID: 123")
//
//	// Create a validation error
//	err := errors.NewValidationError("invalid email format", map[string]string{
//		"email": "must be a valid email address",
//	})
//
//	// Use in HTTP handler
//	if appErr, ok := err.(*errors.AppError); ok {
//		http.Error(w, appErr.Message, appErr.HTTPStatus)
//	}
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents application-specific error codes.
type ErrorCode string

const (
	// ErrCodeNotFound indicates a resource was not found.
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeValidation indicates a validation error.
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	// ErrCodeUnauthorized indicates an authentication error.
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	// ErrCodeForbidden indicates an authorization error.
	ErrCodeForbidden ErrorCode = "FORBIDDEN"
	// ErrCodeConflict indicates a conflict with existing data.
	ErrCodeConflict ErrorCode = "CONFLICT"
	// ErrCodeInternal indicates an internal server error.
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
	// ErrCodeBadRequest indicates a bad request.
	ErrCodeBadRequest ErrorCode = "BAD_REQUEST"
	// ErrCodeTooManyRequests indicates rate limiting.
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
)

// AppError represents an application error with additional context.
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Err        error                  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// ToJSON converts the error to a JSON byte slice for API responses.
func (e *AppError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// NewAppError creates a new AppError with the given parameters.
func NewAppError(code ErrorCode, message string, httpStatus int, details map[string]interface{}, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(resource string, message string) *AppError {
	if message == "" {
		message = fmt.Sprintf("%s not found", resource)
	}
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    message,
		HTTPStatus: http.StatusNotFound,
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string, fieldErrors map[string]string) *AppError {
	details := make(map[string]interface{})
	if len(fieldErrors) > 0 {
		details["fields"] = fieldErrors
	}
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Details:    details,
	}
}

// NewUnauthorizedError creates a new unauthorized error.
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "authentication required"
	}
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a new forbidden error.
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "access denied"
	}
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewConflictError creates a new conflict error.
func NewConflictError(resource string, message string) *AppError {
	if message == "" {
		message = fmt.Sprintf("%s already exists", resource)
	}
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}

// NewInternalError creates a new internal server error.
func NewInternalError(message string, err error) *AppError {
	if message == "" {
		message = "an internal error occurred"
	}
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewBadRequestError creates a new bad request error.
func NewBadRequestError(message string) *AppError {
	if message == "" {
		message = "bad request"
	}
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewTooManyRequestsError creates a new rate limit error.
func NewTooManyRequestsError(message string) *AppError {
	if message == "" {
		message = "too many requests, please try again later"
	}
	return &AppError{
		Code:       ErrCodeTooManyRequests,
		Message:    message,
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// Wrap wraps an existing error with additional context.
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, add context
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Code:       appErr.Code,
			Message:    fmt.Sprintf("%s: %s", message, appErr.Message),
			Details:    appErr.Details,
			HTTPStatus: appErr.HTTPStatus,
			Err:        appErr.Err,
		}
	}

	// Otherwise, create a new internal error
	return NewInternalError(message, err)
}

// IsNotFound checks if the error is a not found error.
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrCodeNotFound
	}
	return false
}

// IsValidation checks if the error is a validation error.
func IsValidation(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrCodeValidation
	}
	return false
}

// IsUnauthorized checks if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrCodeUnauthorized
	}
	return false
}

// IsForbidden checks if the error is a forbidden error.
func IsForbidden(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrCodeForbidden
	}
	return false
}

// IsConflict checks if the error is a conflict error.
func IsConflict(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrCodeConflict
	}
	return false
}

// IsInternal checks if the error is an internal error.
func IsInternal(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrCodeInternal
	}
	return false
}

// GetHTTPStatus returns the HTTP status code for an error.
// If the error is not an AppError, it returns 500 Internal Server Error.
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// ErrorResponse represents a standardized error response for APIs.
type ErrorResponse struct {
	Error struct {
		Code    ErrorCode              `json:"code"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details,omitempty"`
	} `json:"error"`
}

// NewErrorResponse creates a new ErrorResponse from an AppError.
func NewErrorResponse(err *AppError) *ErrorResponse {
	resp := &ErrorResponse{}
	resp.Error.Code = err.Code
	resp.Error.Message = err.Message
	resp.Error.Details = err.Details
	return resp
}
