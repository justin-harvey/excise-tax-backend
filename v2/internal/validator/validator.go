package validator

import (
	"regexp"
	"strings"
)

// Validator holds validation errors
type Validator struct {
	Errors map[string]string
}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// Valid returns true if there are no errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message for a given field
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message if validation check fails
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// NotBlank checks that a field is not blank
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars checks that a field has no more than n characters
func MaxChars(value string, n int) bool {
	return len(value) <= n
}

// MinChars checks that a field has at least n characters
func MinChars(value string, n int) bool {
	return len(value) >= n
}

// In checks that a value is in a list of permitted values
func In(value string, list ...string) bool {
	for _, item := range list {
		if value == item {
			return true
		}
	}
	return false
}

// Matches checks that a value matches a regular expression pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique checks that all values in a slice are unique
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

// ValidXRPLAddress checks if a string is a valid XRPL address
func ValidXRPLAddress(address string) bool {
	if !strings.HasPrefix(address, "r") {
		return false
	}
	if len(address) < 25 || len(address) > 35 {
		return false
	}
	return true
}

// ValidTxHash checks if a string is a valid transaction hash
func ValidTxHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	matched, _ := regexp.MatchString("^[0-9A-Fa-f]+$", hash)
	return matched
}
