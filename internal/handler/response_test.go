package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/middleware"
)

func TestWriteJSON(t *testing.T) {
	t.Run("sets Content-Type and status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"hello": "world"}

		writeJSON(w, http.StatusOK, data)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		var result map[string]string
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if result["hello"] != "world" {
			t.Errorf("expected hello=world, got %q", result["hello"])
		}
	})

	t.Run("handles nil data", func(t *testing.T) {
		w := httptest.NewRecorder()

		writeJSON(w, http.StatusNoContent, nil)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
		}
		if w.Body.Len() != 0 {
			t.Errorf("expected empty body, got %q", w.Body.String())
		}
	})

	t.Run("writes correct status for created", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]int{"id": 1}

		writeJSON(w, http.StatusCreated, data)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})
}

func TestMapDomainError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantStatus   int
		wantCode     string
	}{
		{
			name:       "ErrValidation",
			err:        domain.ErrValidation,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "unprocessable_entity",
		},
		{
			name:       "ValidationError (wraps ErrValidation)",
			err:        &domain.ValidationError{Fields: []domain.FieldError{{Field: "email", Message: "required"}}},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "unprocessable_entity",
		},
		{
			name:       "ErrUnauthorized",
			err:        domain.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "unauthorized",
		},
		{
			name:       "ErrTokenExpired",
			err:        domain.ErrTokenExpired,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "token_expired",
		},
		{
			name:       "ErrForbidden",
			err:        domain.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantCode:   "forbidden",
		},
		{
			name:       "ErrNotFound",
			err:        domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
		},
		{
			name:       "ErrConflict",
			err:        domain.ErrConflict,
			wantStatus: http.StatusConflict,
			wantCode:   "conflict",
		},
		{
			name:       "ErrRateLimited",
			err:        domain.ErrRateLimited,
			wantStatus: http.StatusTooManyRequests,
			wantCode:   "rate_limited",
		},
		{
			name:       "ErrInvalidAmount",
			err:        domain.ErrInvalidAmount,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "unprocessable_entity",
		},
		{
			name:       "unknown error maps to 500",
			err:        errors.New("something went wrong"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
		{
			name:       "wrapped ErrNotFound",
			err:        fmt.Errorf("account: %w", domain.ErrNotFound),
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
		},
		{
			name:       "wrapped ErrUnauthorized",
			err:        fmt.Errorf("auth: %w", domain.ErrUnauthorized),
			wantStatus: http.StatusUnauthorized,
			wantCode:   "unauthorized",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, code := mapDomainError(tc.err)
			if status != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, status)
			}
			if code != tc.wantCode {
				t.Errorf("expected code %q, got %q", tc.wantCode, code)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	const testRequestID = "test-request-id-123"

	newRequestWithID := func() *http.Request {
		ctx := context.WithValue(context.Background(), middleware.RequestIDKey, testRequestID)
		return httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	}

	t.Run("domain error returns proper response", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := newRequestWithID()

		writeError(w, r, domain.ErrNotFound)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		var resp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Code != "not_found" {
			t.Errorf("expected code not_found, got %q", resp.Code)
		}
		if resp.Message != domain.ErrNotFound.Error() {
			t.Errorf("expected message %q, got %q", domain.ErrNotFound.Error(), resp.Message)
		}
		if resp.RequestID != testRequestID {
			t.Errorf("expected request_id %q, got %q", testRequestID, resp.RequestID)
		}
		if resp.Details != nil {
			t.Errorf("expected no details, got %v", resp.Details)
		}
	})

	t.Run("500 error uses generic message", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := newRequestWithID()

		writeError(w, r, errors.New("sql: connection refused"))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var resp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Code != "internal_error" {
			t.Errorf("expected code internal_error, got %q", resp.Code)
		}
		if resp.Message != "an unexpected error occurred" {
			t.Errorf("expected generic message, got %q", resp.Message)
		}
		// Ensure internal details are NOT exposed
		if resp.Message == "sql: connection refused" {
			t.Error("internal error details leaked to response")
		}
	})

	t.Run("ValidationError includes details", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := newRequestWithID()

		ve := &domain.ValidationError{
			Fields: []domain.FieldError{
				{Field: "email", Message: "must be a valid email"},
				{Field: "password", Message: "must be at least 8 characters"},
			},
		}

		writeError(w, r, ve)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}

		var resp struct {
			Code      string             `json:"code"`
			Message   string             `json:"message"`
			RequestID string             `json:"request_id"`
			Details   []ValidationDetail `json:"details"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Code != "unprocessable_entity" {
			t.Errorf("expected code unprocessable_entity, got %q", resp.Code)
		}
		if len(resp.Details) != 2 {
			t.Fatalf("expected 2 details, got %d", len(resp.Details))
		}
		if resp.Details[0].Field != "email" {
			t.Errorf("expected first detail field=email, got %q", resp.Details[0].Field)
		}
		if resp.Details[0].Message != "must be a valid email" {
			t.Errorf("expected first detail message, got %q", resp.Details[0].Message)
		}
		if resp.Details[1].Field != "password" {
			t.Errorf("expected second detail field=password, got %q", resp.Details[1].Field)
		}
	})

	t.Run("request ID from context is included", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := newRequestWithID()

		writeError(w, r, domain.ErrForbidden)

		var resp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.RequestID != testRequestID {
			t.Errorf("expected request_id %q, got %q", testRequestID, resp.RequestID)
		}
	})

	t.Run("missing request ID returns empty string", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		writeError(w, r, domain.ErrUnauthorized)

		var resp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.RequestID != "" {
			t.Errorf("expected empty request_id, got %q", resp.RequestID)
		}
	})

	t.Run("Content-Type is application/json", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := newRequestWithID()

		writeError(w, r, domain.ErrConflict)

		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}
	})
}
