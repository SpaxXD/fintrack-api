package validator

import (
	"errors"
	"testing"

	"github.com/muriloabranches/fintrack-api/internal/domain"
)

func TestToCents_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected int64
	}{
		{"whole number", 10.0, 1000},
		{"one decimal", 12.5, 1250},
		{"two decimals", 99.99, 9999},
		{"zero", 0.0, 0},
		{"small value", 0.01, 1},
		{"negative value", -5.50, -550},
		{"large value", 999999.99, 99999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToCents(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("ToCents(%v) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCents_InvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		input float64
	}{
		{"three decimals", 10.123},
		{"four decimals", 5.1234},
		{"many decimals", 1.999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ToCents(tt.input)
			if err == nil {
				t.Fatalf("expected error for input %v, got nil", tt.input)
			}
			if !errors.Is(err, domain.ErrInvalidAmount) {
				t.Errorf("expected ErrInvalidAmount, got %v", err)
			}
		})
	}
}

func TestFromCents(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected float64
	}{
		{"1000 cents", 1000, 10.0},
		{"1250 cents", 1250, 12.50},
		{"9999 cents", 9999, 99.99},
		{"0 cents", 0, 0.0},
		{"1 cent", 1, 0.01},
		{"negative cents", -550, -5.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromCents(tt.input)
			if result != tt.expected {
				t.Errorf("FromCents(%d) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

// Test structs for validation tests
type validStruct struct {
	Name  string `validate:"required,min=1,max=100"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=0"`
}

type invalidStruct struct {
	Name  string `validate:"required,min=1,max=100"`
	Email string `validate:"required,email"`
}

func TestValidate_ValidStruct(t *testing.T) {
	s := validStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	result := Validate(s)
	if result != nil {
		t.Fatalf("expected nil for valid struct, got %v", result)
	}
}

func TestValidate_InvalidStruct(t *testing.T) {
	s := invalidStruct{
		Name:  "", // required, min=1
		Email: "not-an-email",
	}

	result := Validate(s)
	if result == nil {
		t.Fatal("expected ValidationError for invalid struct, got nil")
	}

	if len(result.Fields) == 0 {
		t.Fatal("expected at least one field error")
	}

	// Check that it wraps ErrValidation
	if !errors.Is(result, domain.ErrValidation) {
		t.Error("expected ValidationError to wrap ErrValidation")
	}

	// Verify field errors contain expected fields
	fieldMap := make(map[string]string)
	for _, fe := range result.Fields {
		fieldMap[fe.Field] = fe.Message
	}

	if _, ok := fieldMap["Name"]; !ok {
		t.Error("expected field error for 'Name'")
	}
	if _, ok := fieldMap["Email"]; !ok {
		t.Error("expected field error for 'Email'")
	}
}

func TestValidate_RequiredFieldMessage(t *testing.T) {
	type reqStruct struct {
		Title string `validate:"required"`
	}

	result := Validate(reqStruct{Title: ""})
	if result == nil {
		t.Fatal("expected ValidationError, got nil")
	}

	if len(result.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(result.Fields))
	}

	if result.Fields[0].Message != "is required" {
		t.Errorf("expected message 'is required', got '%s'", result.Fields[0].Message)
	}
}

func TestValidate_MinFieldMessage(t *testing.T) {
	type minStruct struct {
		Password string `validate:"min=8"`
	}

	result := Validate(minStruct{Password: "short"})
	if result == nil {
		t.Fatal("expected ValidationError, got nil")
	}

	if len(result.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(result.Fields))
	}

	if result.Fields[0].Message != "must be at least 8 characters" {
		t.Errorf("expected message 'must be at least 8 characters', got '%s'", result.Fields[0].Message)
	}
}

func TestValidate_EmailFieldMessage(t *testing.T) {
	type emailStruct struct {
		Email string `validate:"email"`
	}

	result := Validate(emailStruct{Email: "invalid"})
	if result == nil {
		t.Fatal("expected ValidationError, got nil")
	}

	if len(result.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(result.Fields))
	}

	if result.Fields[0].Message != "must be a valid email address" {
		t.Errorf("expected message 'must be a valid email address', got '%s'", result.Fields[0].Message)
	}
}

func TestValidate_OneofFieldMessage(t *testing.T) {
	type oneofStruct struct {
		Type string `validate:"oneof=income expense"`
	}

	result := Validate(oneofStruct{Type: "invalid"})
	if result == nil {
		t.Fatal("expected ValidationError, got nil")
	}

	if len(result.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(result.Fields))
	}

	expected := "must be one of: income expense"
	if result.Fields[0].Message != expected {
		t.Errorf("expected message '%s', got '%s'", expected, result.Fields[0].Message)
	}
}
