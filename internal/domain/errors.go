package domain

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors de domínio.
var (
	ErrNotFound      = errors.New("resource not found")
	ErrConflict      = errors.New("resource already exists")
	ErrForbidden     = errors.New("access denied")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrTokenExpired  = errors.New("token expired")
	ErrValidation    = errors.New("validation error")
	ErrRateLimited   = errors.New("rate limited")
	ErrInternalError = errors.New("internal error")
	ErrInvalidAmount = errors.New("invalid amount")
)

// FieldError represents a single field validation failure.
type FieldError struct {
	Field   string
	Message string
}

// ValidationError carries details about invalid fields.
// It implements the error interface and wraps ErrValidation so that
// errors.Is(ve, ErrValidation) returns true.
type ValidationError struct {
	Fields []FieldError
}

// Error returns a human-readable message listing all invalid fields.
func (ve *ValidationError) Error() string {
	if len(ve.Fields) == 0 {
		return "validation error"
	}

	msgs := make([]string, 0, len(ve.Fields))
	for _, f := range ve.Fields {
		msgs = append(msgs, fmt.Sprintf("%s: %s", f.Field, f.Message))
	}

	return fmt.Sprintf("validation error: %s", strings.Join(msgs, "; "))
}

// Unwrap returns ErrValidation so errors.Is works correctly.
func (ve *ValidationError) Unwrap() error {
	return ErrValidation
}
