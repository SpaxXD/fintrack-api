package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestRequestID_SetsHeader(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	got := rec.Header().Get(HeaderXRequestID)
	if got == "" {
		t.Fatal("expected X-Request-ID header to be set, got empty string")
	}

	// Verify it is a valid UUID v4.
	parsed, err := uuid.Parse(got)
	if err != nil {
		t.Fatalf("expected valid UUID in X-Request-ID header, got %q: %v", got, err)
	}
	if parsed.Version() != 4 {
		t.Fatalf("expected UUID v4, got version %d", parsed.Version())
	}
}

func TestRequestID_StoresInContext(t *testing.T) {
	var ctxID string

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if ctxID == "" {
		t.Fatal("expected request ID in context, got empty string")
	}

	// The context value should match the response header.
	headerID := rec.Header().Get(HeaderXRequestID)
	if ctxID != headerID {
		t.Fatalf("context ID %q does not match header ID %q", ctxID, headerID)
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ids := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		id := rec.Header().Get(HeaderXRequestID)
		if _, exists := ids[id]; exists {
			t.Fatalf("duplicate request ID generated: %s", id)
		}
		ids[id] = struct{}{}
	}
}

func TestGetRequestID_EmptyWhenNotSet(t *testing.T) {
	ctx := context.Background()
	id := GetRequestID(ctx)
	if id != "" {
		t.Fatalf("expected empty string from context without request ID, got %q", id)
	}
}
