package validator

import (
	"testing"
)

type TestStruct struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Age      int    `validate:"required,gte=18,lte=100"`
	Phone    string `validate:"omitempty,phone"`
}

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Error("expected validator to be created")
	}
	if v.validate == nil {
		t.Error("expected validate field to be initialized")
	}
}

func TestValidate_Success(t *testing.T) {
	v := New()
	data := TestStruct{
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
		Phone:    "123-456-7890",
	}

	err := v.Validate(data)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidate_RequiredFieldMissing(t *testing.T) {
	v := New()
	data := TestStruct{
		Password: "password123",
		Age:      25,
	}

	err := v.Validate(data)
	if err == nil {
		t.Error("expected validation error for missing email")
	}
}

func TestValidate_InvalidEmail(t *testing.T) {
	v := New()
	data := TestStruct{
		Email:    "invalid-email",
		Password: "password123",
		Age:      25,
	}

	err := v.Validate(data)
	if err == nil {
		t.Error("expected validation error for invalid email")
	}
}

func TestValidate_PasswordTooShort(t *testing.T) {
	v := New()
	data := TestStruct{
		Email:    "test@example.com",
		Password: "short",
		Age:      25,
	}

	err := v.Validate(data)
	if err == nil {
		t.Error("expected validation error for short password")
	}
}

func TestFormatErrors(t *testing.T) {
	v := New()
	data := TestStruct{
		Email:    "invalid",
		Password: "short",
		Age:      10,
	}

	err := v.Validate(data)
	if err == nil {
		t.Fatal("expected validation error")
	}

	errors := v.FormatErrors(err)
	if errors == nil {
		t.Error("expected errors map to be returned")
	}

	if _, ok := errors["email"]; !ok {
		t.Error("expected email error")
	}

	if _, ok := errors["password"]; !ok {
		t.Error("expected password error")
	}

	if _, ok := errors["age"]; !ok {
		t.Error("expected age error")
	}
}

func TestValidatePhone_Valid(t *testing.T) {
	v := New()

	validPhones := []string{
		"1234567890",
		"123-456-7890",
		"(123) 456-7890",
		"+1 123 456 7890",
		"11234567890",
	}

	for _, phone := range validPhones {
		err := v.ValidateVar(phone, "phone")
		if err != nil {
			t.Errorf("expected phone '%s' to be valid, got error: %v", phone, err)
		}
	}
}

func TestValidatePhone_Invalid(t *testing.T) {
	v := New()

	invalidPhones := []string{
		"123",
		"12345",
		"123456789",
		"abcdefghij",
		"123-45-6789", // Too few digits
	}

	for _, phone := range invalidPhones {
		err := v.ValidateVar(phone, "phone")
		if err == nil {
			t.Errorf("expected phone '%s' to be invalid", phone)
		}
	}
}

func TestValidateFEIN_Valid(t *testing.T) {
	v := New()

	validFEINs := []string{
		"12-3456789",
		"123456789",
	}

	for _, fein := range validFEINs {
		err := v.ValidateVar(fein, "fein")
		if err != nil {
			t.Errorf("expected FEIN '%s' to be valid, got error: %v", fein, err)
		}
	}
}

func TestValidateFEIN_Invalid(t *testing.T) {
	v := New()

	invalidFEINs := []string{
		"12345",
		"1234567890", // Too many digits
		"12-345678",  // Too few digits
		"ab-cdefghi",
		"",
	}

	for _, fein := range invalidFEINs {
		err := v.ValidateVar(fein, "fein")
		if err == nil {
			t.Errorf("expected FEIN '%s' to be invalid", fein)
		}
	}
}

func TestValidateStateCode_Valid(t *testing.T) {
	v := New()

	validStates := []string{
		"CA", "NY", "TX", "FL", "IL",
		"ca", "ny", "tx", // Should work with lowercase too
		"DC", "PR",
	}

	for _, state := range validStates {
		err := v.ValidateVar(state, "state_code")
		if err != nil {
			t.Errorf("expected state code '%s' to be valid, got error: %v", state, err)
		}
	}
}

func TestValidateStateCode_Invalid(t *testing.T) {
	v := New()

	invalidStates := []string{
		"XX",
		"ABC",
		"1A",
		"",
		"California",
	}

	for _, state := range invalidStates {
		err := v.ValidateVar(state, "state_code")
		if err == nil {
			t.Errorf("expected state code '%s' to be invalid", state)
		}
	}
}

func TestValidateZipCode_Valid(t *testing.T) {
	v := New()

	validZips := []string{
		"12345",
		"12345-6789",
	}

	for _, zip := range validZips {
		err := v.ValidateVar(zip, "zip")
		if err != nil {
			t.Errorf("expected ZIP code '%s' to be valid, got error: %v", zip, err)
		}
	}
}

func TestValidateZipCode_Invalid(t *testing.T) {
	v := New()

	invalidZips := []string{
		"1234",
		"123456",
		"12345-678",
		"abcde",
		"",
	}

	for _, zip := range invalidZips {
		err := v.ValidateVar(zip, "zip")
		if err == nil {
			t.Errorf("expected ZIP code '%s' to be invalid", zip)
		}
	}
}

func TestValidateStrongPassword_Valid(t *testing.T) {
	v := New()

	validPasswords := []string{
		"Password1!",
		"MyP@ssw0rd",
		"Str0ng!Pass",
		"Abcdef1!ghij",
	}

	for _, password := range validPasswords {
		err := v.ValidateVar(password, "strong_password")
		if err != nil {
			t.Errorf("expected password '%s' to be valid, got error: %v", password, err)
		}
	}
}

func TestValidateStrongPassword_Invalid(t *testing.T) {
	v := New()

	invalidPasswords := []string{
		"password",      // No uppercase, no number, no special
		"PASSWORD",      // No lowercase, no number, no special
		"Password",      // No number, no special
		"Password1",     // No special
		"Pass1!",        // Too short
		"password1!",    // No uppercase
		"PASSWORD1!",    // No lowercase
		"PasswordWord!", // No number
	}

	for _, password := range invalidPasswords {
		err := v.ValidateVar(password, "strong_password")
		if err == nil {
			t.Errorf("expected password '%s' to be invalid", password)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Email", "email"},
		{"FirstName", "first_name"},
		{"UserID", "user_i_d"},
		{"HTTPRequest", "h_t_t_p_request"},
		{"lowercase", "lowercase"},
	}

	for _, test := range tests {
		result := toSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("toSnakeCase(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestFormatErrorsList(t *testing.T) {
	v := New()
	data := TestStruct{
		Email:    "invalid",
		Password: "short",
		Age:      10,
	}

	err := v.Validate(data)
	if err == nil {
		t.Fatal("expected validation error")
	}

	errors := v.FormatErrorsList(err)
	if len(errors) == 0 {
		t.Error("expected errors list to be returned")
	}

	// Check that errors have the expected structure
	for _, e := range errors {
		if e.Field == "" {
			t.Error("expected field name to be set")
		}
		if e.Message == "" {
			t.Error("expected error message to be set")
		}
	}
}

func TestFormatErrors_NilError(t *testing.T) {
	v := New()
	errors := v.FormatErrors(nil)

	if errors != nil {
		t.Error("expected nil for nil error")
	}
}

func TestFormatErrorsList_NilError(t *testing.T) {
	v := New()
	errors := v.FormatErrorsList(nil)

	if errors != nil {
		t.Error("expected nil for nil error")
	}
}
