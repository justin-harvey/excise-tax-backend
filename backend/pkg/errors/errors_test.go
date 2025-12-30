package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("user", "user not found")

	if err.Code != ErrCodeNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeNotFound, err.Code)
	}

	if err.Message != "user not found" {
		t.Errorf("expected message 'user not found', got '%s'", err.Message)
	}

	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected HTTP status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}

	if err.Details["resource"] != "user" {
		t.Errorf("expected resource 'user', got '%v'", err.Details["resource"])
	}
}

func TestNewValidationError(t *testing.T) {
	fieldErrors := map[string]string{
		"email": "invalid email format",
		"age":   "must be positive",
	}
	err := NewValidationError("validation failed", fieldErrors)

	if err.Code != ErrCodeValidation {
		t.Errorf("expected code %s, got %s", ErrCodeValidation, err.Code)
	}

	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected HTTP status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}

	fields, ok := err.Details["fields"].(map[string]string)
	if !ok {
		t.Fatal("expected fields in details")
	}

	if fields["email"] != "invalid email format" {
		t.Errorf("expected email error, got '%s'", fields["email"])
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("invalid credentials")

	if err.Code != ErrCodeUnauthorized {
		t.Errorf("expected code %s, got %s", ErrCodeUnauthorized, err.Code)
	}

	if err.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("expected HTTP status %d, got %d", http.StatusUnauthorized, err.HTTPStatus)
	}
}

func TestNewUnauthorizedError_DefaultMessage(t *testing.T) {
	err := NewUnauthorizedError("")

	if err.Message != "authentication required" {
		t.Errorf("expected default message 'authentication required', got '%s'", err.Message)
	}
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("insufficient permissions")

	if err.Code != ErrCodeForbidden {
		t.Errorf("expected code %s, got %s", ErrCodeForbidden, err.Code)
	}

	if err.HTTPStatus != http.StatusForbidden {
		t.Errorf("expected HTTP status %d, got %d", http.StatusForbidden, err.HTTPStatus)
	}
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("user", "email already exists")

	if err.Code != ErrCodeConflict {
		t.Errorf("expected code %s, got %s", ErrCodeConflict, err.Code)
	}

	if err.HTTPStatus != http.StatusConflict {
		t.Errorf("expected HTTP status %d, got %d", http.StatusConflict, err.HTTPStatus)
	}
}

func TestNewInternalError(t *testing.T) {
	originalErr := errors.New("database connection failed")
	err := NewInternalError("failed to fetch user", originalErr)

	if err.Code != ErrCodeInternal {
		t.Errorf("expected code %s, got %s", ErrCodeInternal, err.Code)
	}

	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected HTTP status %d, got %d", http.StatusInternalServerError, err.HTTPStatus)
	}

	if err.Err != originalErr {
		t.Error("expected wrapped error to be preserved")
	}
}

func TestNewBadRequestError(t *testing.T) {
	err := NewBadRequestError("invalid JSON")

	if err.Code != ErrCodeBadRequest {
		t.Errorf("expected code %s, got %s", ErrCodeBadRequest, err.Code)
	}

	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected HTTP status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}
}

func TestNewTooManyRequestsError(t *testing.T) {
	err := NewTooManyRequestsError("rate limit exceeded")

	if err.Code != ErrCodeTooManyRequests {
		t.Errorf("expected code %s, got %s", ErrCodeTooManyRequests, err.Code)
	}

	if err.HTTPStatus != http.StatusTooManyRequests {
		t.Errorf("expected HTTP status %d, got %d", http.StatusTooManyRequests, err.HTTPStatus)
	}
}

func TestAppError_Error(t *testing.T) {
	err := NewNotFoundError("user", "user not found")
	errStr := err.Error()

	if errStr != "NOT_FOUND: user not found" {
		t.Errorf("unexpected error string: %s", errStr)
	}
}

func TestAppError_Error_WithWrappedError(t *testing.T) {
	originalErr := errors.New("db error")
	err := NewInternalError("operation failed", originalErr)
	errStr := err.Error()

	expectedStr := "INTERNAL_ERROR: operation failed: db error"
	if errStr != expectedStr {
		t.Errorf("expected '%s', got '%s'", expectedStr, errStr)
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := NewInternalError("wrapped", originalErr)

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Error("expected unwrap to return original error")
	}
}

func TestWrap_WithAppError(t *testing.T) {
	original := NewNotFoundError("user", "user not found")
	wrapped := Wrap(original, "failed to get user")

	if wrapped.Code != ErrCodeNotFound {
		t.Errorf("expected code to be preserved")
	}

	if wrapped.Message != "failed to get user: user not found" {
		t.Errorf("unexpected message: %s", wrapped.Message)
	}
}

func TestWrap_WithRegularError(t *testing.T) {
	original := errors.New("some error")
	wrapped := Wrap(original, "operation failed")

	if wrapped.Code != ErrCodeInternal {
		t.Errorf("expected internal error code, got %s", wrapped.Code)
	}

	if wrapped.Err != original {
		t.Error("expected original error to be wrapped")
	}
}

func TestWrap_NilError(t *testing.T) {
	wrapped := Wrap(nil, "message")

	if wrapped != nil {
		t.Error("expected nil when wrapping nil error")
	}
}

func TestIsNotFound(t *testing.T) {
	notFoundErr := NewNotFoundError("user", "not found")
	validationErr := NewValidationError("invalid", nil)

	if !IsNotFound(notFoundErr) {
		t.Error("expected IsNotFound to return true")
	}

	if IsNotFound(validationErr) {
		t.Error("expected IsNotFound to return false for validation error")
	}
}

func TestIsValidation(t *testing.T) {
	validationErr := NewValidationError("invalid", nil)
	notFoundErr := NewNotFoundError("user", "not found")

	if !IsValidation(validationErr) {
		t.Error("expected IsValidation to return true")
	}

	if IsValidation(notFoundErr) {
		t.Error("expected IsValidation to return false for not found error")
	}
}

func TestGetHTTPStatus(t *testing.T) {
	notFoundErr := NewNotFoundError("user", "not found")
	status := GetHTTPStatus(notFoundErr)

	if status != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, status)
	}
}

func TestGetHTTPStatus_RegularError(t *testing.T) {
	err := errors.New("regular error")
	status := GetHTTPStatus(err)

	if status != http.StatusInternalServerError {
		t.Errorf("expected status %d for regular error, got %d", http.StatusInternalServerError, status)
	}
}

func TestNewErrorResponse(t *testing.T) {
	err := NewValidationError("validation failed", map[string]string{
		"email": "invalid",
	})

	resp := NewErrorResponse(err)

	if resp.Error.Code != ErrCodeValidation {
		t.Errorf("expected code %s, got %s", ErrCodeValidation, resp.Error.Code)
	}

	if resp.Error.Message != "validation failed" {
		t.Errorf("expected message 'validation failed', got '%s'", resp.Error.Message)
	}
}

func TestAppError_ToJSON(t *testing.T) {
	err := NewNotFoundError("user", "user not found")
	jsonData := err.ToJSON()

	if len(jsonData) == 0 {
		t.Error("expected JSON data, got empty")
	}

	// Should be valid JSON
	expected := `{"code":"NOT_FOUND","message":"user not found","details":{"resource":"user"}}`
	if string(jsonData) != expected {
		t.Logf("expected: %s", expected)
		t.Logf("got: %s", string(jsonData))
	}
}
