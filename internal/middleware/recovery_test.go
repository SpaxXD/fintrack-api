package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovery_PanicReturns500WithJSON(t *testing.T) {
	// Handler that panics.
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	// Wrap with Recovery middleware.
	handler := Recovery(panicking)

	// Inject a request_id into the context to simulate RequestID middleware.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), RequestIDKey, "test-request-id-123")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Verify status code.
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	// Verify Content-Type.
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	// Verify JSON body structure.
	var resp recoveryErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.Code != "internal_error" {
		t.Errorf("expected code %q, got %q", "internal_error", resp.Code)
	}
	if resp.Message != "An internal error occurred" {
		t.Errorf("expected message %q, got %q", "An internal error occurred", resp.Message)
	}
	if resp.RequestID != "test-request-id-123" {
		t.Errorf("expected request_id %q, got %q", "test-request-id-123", resp.RequestID)
	}
}

func TestRecovery_PanicDoesNotExposeInternalDetails(t *testing.T) {
	// Handler that panics with SQL-related content.
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("pq: duplicate key value violates unique constraint \"users_email_key\"")
	})

	handler := Recovery(panicking)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), RequestIDKey, "req-abc")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body := rr.Body.String()

	// The response must NOT contain stack traces, SQL errors, or panic values.
	sensitivePatterns := []string{
		"goroutine",
		"runtime",
		"panic",
		"duplicate key",
		"users_email_key",
		"pq:",
		".go:",
		"stack",
	}

	for _, pattern := range sensitivePatterns {
		if containsString(body, pattern) {
			t.Errorf("response body should not contain %q, got: %s", pattern, body)
		}
	}
}

func TestRecovery_NormalRequestPassesThrough(t *testing.T) {
	// Handler that responds normally.
	normal := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	handler := Recovery(normal)

	req := httptest.NewRequest(http.MethodGet, "/healthy", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	expected := `{"status":"ok"}`
	if rr.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rr.Body.String())
	}
}

func TestRecovery_PanicWithNilValue(t *testing.T) {
	// Handler that panics with nil.
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(nil)
	})

	handler := Recovery(panicking)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), RequestIDKey, "req-nil")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// In Go 1.21+, panic(nil) is caught by recover() and returns a *runtime.PanicNilError.
	// The middleware should still return 500.
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	var resp recoveryErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.Code != "internal_error" {
		t.Errorf("expected code %q, got %q", "internal_error", resp.Code)
	}
}

// containsString checks if s contains substr (case-sensitive).
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
