// Package validator provides request validation using go-playground/validator
// with custom validation rules for the excise tax portal.
//
// Example usage:
//
//	type CreateUserRequest struct {
//		Email    string `json:"email" validate:"required,email"`
//		Password string `json:"password" validate:"required,min=8"`
//		Phone    string `json:"phone" validate:"omitempty,phone"`
//	}
//
//	v := validator.New()
//	req := CreateUserRequest{
//		Email:    "test@example.com",
//		Password: "password123",
//	}
//
//	if err := v.Validate(req); err != nil {
//		// Handle validation error
//		fmt.Println(v.FormatErrors(err))
//	}
package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator with custom rules.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator with custom validation rules.
func New() *Validator {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("phone", validatePhone)
	v.RegisterValidation("fein", validateFEIN)
	v.RegisterValidation("state_code", validateStateCode)
	v.RegisterValidation("zip", validateZipCode)
	v.RegisterValidation("strong_password", validateStrongPassword)

	return &Validator{
		validate: v,
	}
}

// Validate validates a struct using the validation tags.
func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

// ValidateVar validates a single variable.
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// FormatErrors converts validation errors into a map of field names to error messages.
func (v *Validator) FormatErrors(err error) map[string]string {
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{
			"error": err.Error(),
		}
	}

	errors := make(map[string]string)
	for _, e := range validationErrors {
		fieldName := toSnakeCase(e.Field())
		errors[fieldName] = formatFieldError(e)
	}

	return errors
}

// formatFieldError formats a single validation error into a human-readable message.
func formatFieldError(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "fein":
		return fmt.Sprintf("%s must be a valid Federal Employer Identification Number (XX-XXXXXXX)", field)
	case "state_code":
		return fmt.Sprintf("%s must be a valid 2-letter state code", field)
	case "zip":
		return fmt.Sprintf("%s must be a valid ZIP code (XXXXX or XXXXX-XXXX)", field)
	case "strong_password":
		return fmt.Sprintf("%s must contain at least one uppercase letter, one lowercase letter, one number, and one special character", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	default:
		return fmt.Sprintf("%s failed validation for tag '%s'", field, e.Tag())
	}
}

// toSnakeCase converts a string from CamelCase to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// Custom validator functions

// validatePhone validates US phone numbers in various formats.
// Accepts: (123) 456-7890, 123-456-7890, 1234567890, +1 123 456 7890
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// Remove all non-digit characters
	digitsOnly := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Must have 10 digits, or 11 digits starting with 1
	if len(digitsOnly) == 10 {
		return true
	}
	if len(digitsOnly) == 11 && digitsOnly[0] == '1' {
		return true
	}
	return false
}

// validateFEIN validates Federal Employer Identification Numbers.
// Format: XX-XXXXXXX (9 digits total)
func validateFEIN(fl validator.FieldLevel) bool {
	fein := fl.Field().String()
	// Remove hyphens
	digitsOnly := strings.ReplaceAll(fein, "-", "")

	// Must be exactly 9 digits
	if len(digitsOnly) != 9 {
		return false
	}

	// All characters must be digits
	matched, _ := regexp.MatchString(`^\d{9}$`, digitsOnly)
	return matched
}

// validateStateCode validates US state codes (2 letters).
func validateStateCode(fl validator.FieldLevel) bool {
	code := strings.ToUpper(fl.Field().String())

	validStates := map[string]bool{
		"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true,
		"CO": true, "CT": true, "DE": true, "FL": true, "GA": true,
		"HI": true, "ID": true, "IL": true, "IN": true, "IA": true,
		"KS": true, "KY": true, "LA": true, "ME": true, "MD": true,
		"MA": true, "MI": true, "MN": true, "MS": true, "MO": true,
		"MT": true, "NE": true, "NV": true, "NH": true, "NJ": true,
		"NM": true, "NY": true, "NC": true, "ND": true, "OH": true,
		"OK": true, "OR": true, "PA": true, "RI": true, "SC": true,
		"SD": true, "TN": true, "TX": true, "UT": true, "VT": true,
		"VA": true, "WA": true, "WV": true, "WI": true, "WY": true,
		"DC": true, "PR": true, "VI": true, "GU": true, "AS": true,
		"MP": true,
	}

	return validStates[code]
}

// validateZipCode validates US ZIP codes.
// Accepts: XXXXX or XXXXX-XXXX
func validateZipCode(fl validator.FieldLevel) bool {
	zip := fl.Field().String()

	// 5-digit ZIP
	if matched, _ := regexp.MatchString(`^\d{5}$`, zip); matched {
		return true
	}

	// ZIP+4
	if matched, _ := regexp.MatchString(`^\d{5}-\d{4}$`, zip); matched {
		return true
	}

	return false
}

// validateStrongPassword validates that a password meets strength requirements.
// Requirements: at least 8 characters, contains uppercase, lowercase, number, and special character
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	// Check for uppercase
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	if !hasUpper {
		return false
	}

	// Check for lowercase
	hasLower, _ := regexp.MatchString(`[a-z]`, password)
	if !hasLower {
		return false
	}

	// Check for digit
	hasDigit, _ := regexp.MatchString(`\d`, password)
	if !hasDigit {
		return false
	}

	// Check for special character
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)
	if !hasSpecial {
		return false
	}

	return true
}

// ValidationError represents a structured validation error for API responses.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatErrorsList converts validation errors into a list of ValidationError structs.
func (v *Validator) FormatErrorsList(err error) []ValidationError {
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []ValidationError{
			{
				Field:   "error",
				Message: err.Error(),
			},
		}
	}

	errors := make([]ValidationError, 0, len(validationErrors))
	for _, e := range validationErrors {
		errors = append(errors, ValidationError{
			Field:   toSnakeCase(e.Field()),
			Message: formatFieldError(e),
		})
	}

	return errors
}
