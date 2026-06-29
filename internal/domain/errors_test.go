package domain

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	sentinels := []struct {
		err  error
		name string
	}{
		{ErrNotFound, "ErrNotFound"},
		{ErrConflict, "ErrConflict"},
		{ErrForbidden, "ErrForbidden"},
		{ErrUnauthorized, "ErrUnauthorized"},
		{ErrTokenExpired, "ErrTokenExpired"},
		{ErrValidation, "ErrValidation"},
		{ErrRateLimited, "ErrRateLimited"},
		{ErrInternalError, "ErrInternalError"},
		{ErrInvalidAmount, "ErrInvalidAmount"},
	}

	for _, s := range sentinels {
		t.Run(s.name, func(t *testing.T) {
			if s.err == nil {
				t.Errorf("%s should not be nil", s.name)
			}
			if s.err.Error() == "" {
				t.Errorf("%s should have a non-empty message", s.name)
			}
		})
	}
}

func TestValidationError_Error_NoFields(t *testing.T) {
	ve := &ValidationError{}
	got := ve.Error()
	if got != "validation error" {
		t.Errorf("expected 'validation error', got %q", got)
	}
}

func TestValidationError_Error_SingleField(t *testing.T) {
	ve := &ValidationError{
		Fields: []FieldError{
			{Field: "email", Message: "is required"},
		},
	}
	got := ve.Error()
	expected := "validation error: email: is required"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestValidationError_Error_MultipleFields(t *testing.T) {
	ve := &ValidationError{
		Fields: []FieldError{
			{Field: "email", Message: "is required"},
			{Field: "password", Message: "must be at least 8 characters"},
		},
	}
	got := ve.Error()
	expected := "validation error: email: is required; password: must be at least 8 characters"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	ve := &ValidationError{
		Fields: []FieldError{
			{Field: "name", Message: "too long"},
		},
	}

	if !errors.Is(ve, ErrValidation) {
		t.Error("ValidationError should unwrap to ErrValidation")
	}
}

func TestValidationError_NotMatchOtherErrors(t *testing.T) {
	ve := &ValidationError{
		Fields: []FieldError{
			{Field: "name", Message: "too long"},
		},
	}

	if errors.Is(ve, ErrNotFound) {
		t.Error("ValidationError should not match ErrNotFound")
	}
	if errors.Is(ve, ErrUnauthorized) {
		t.Error("ValidationError should not match ErrUnauthorized")
	}
}
