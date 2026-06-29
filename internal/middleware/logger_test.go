package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// logEntry represents a parsed JSON log line.
type logEntry struct {
	Level      string  `json:"level"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Status     int     `json:"status"`
	DurationMs float64 `json:"duration_ms"`
	RequestID  string  `json:"request_id"`
	UserID     string  `json:"user_id"`
	Message    string  `json:"message"`
}

func parseLogEntry(t *testing.T, buf *bytes.Buffer) logEntry {
	t.Helper()
	var entry logEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v\nraw: %s", err, buf.String())
	}
	return entry
}

func TestLogger_LogsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Logger()

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	reqID := "test-request-id-123"
	ctx := context.WithValue(req.Context(), RequestIDKey, reqID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, &buf)

	if entry.Method != http.MethodGet {
		t.Errorf("expected method GET, got %s", entry.Method)
	}
	if entry.Path != "/api/v1/accounts" {
		t.Errorf("expected path /api/v1/accounts, got %s", entry.Path)
	}
	if entry.Status != http.StatusOK {
		t.Errorf("expected status 200, got %d", entry.Status)
	}
	if entry.DurationMs < 0 {
		t.Errorf("expected non-negative duration_ms, got %f", entry.DurationMs)
	}
	if entry.RequestID != reqID {
		t.Errorf("expected request_id %s, got %s", reqID, entry.RequestID)
	}
	if entry.Level != "info" {
		t.Errorf("expected level info, got %s", entry.Level)
	}
}

func TestLogger_LogsInfoLevelForSuccess(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Logger()

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, &buf)

	if entry.Level != "info" {
		t.Errorf("expected level info for 201, got %s", entry.Level)
	}
	if entry.Status != http.StatusCreated {
		t.Errorf("expected status 201, got %d", entry.Status)
	}
}

func TestLogger_LogsErrorLevelFor5xx(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Logger()

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, &buf)

	if entry.Level != "error" {
		t.Errorf("expected level error for 500, got %s", entry.Level)
	}
	if entry.Status != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", entry.Status)
	}
}

func TestLogger_IncludesUserIDInErrorLogs(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Logger()

	userID := uuid.New()

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, &buf)

	if entry.UserID != userID.String() {
		t.Errorf("expected user_id %s in error log, got %s", userID.String(), entry.UserID)
	}
}

func TestLogger_OmitsUserIDInInfoLogs(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Logger()

	userID := uuid.New()

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, &buf)

	if entry.UserID != "" {
		t.Errorf("expected user_id to be omitted in info log, got %s", entry.UserID)
	}
}

func TestLogger_OmitsUserIDInErrorWhenNotAvailable(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Logger()

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, &buf)

	if entry.UserID != "" {
		t.Errorf("expected user_id to be empty when not in context, got %s", entry.UserID)
	}
	if entry.Level != "error" {
		t.Errorf("expected level error for 502, got %s", entry.Level)
	}
}

func TestLogger_DefaultStatusIsOKIfNotSet(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Logger()

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write body without calling WriteHeader explicitly
		w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, &buf)

	if entry.Status != http.StatusOK {
		t.Errorf("expected default status 200, got %d", entry.Status)
	}
}

func TestLogger_DurationIsPositive(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Logger()

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, &buf)

	if entry.DurationMs < 0 {
		t.Errorf("expected non-negative duration_ms, got %f", entry.DurationMs)
	}
}

func TestLogger_RespectLogLevel(t *testing.T) {
	var buf bytes.Buffer
	// Set log level to error - info messages should be suppressed
	logger := zerolog.New(&buf).Level(zerolog.ErrorLevel)

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if buf.Len() != 0 {
		t.Errorf("expected no log output for info when level is error, got: %s", buf.String())
	}
}

func TestLogger_ErrorLevelStillLogsWhenConfigured(t *testing.T) {
	var buf bytes.Buffer
	// Set log level to error - error messages should still be logged
	logger := zerolog.New(&buf).Level(zerolog.ErrorLevel)

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if buf.Len() == 0 {
		t.Error("expected error log output when level is error and status is 500")
	}

	entry := parseLogEntry(t, &buf)
	if entry.Level != "error" {
		t.Errorf("expected level error, got %s", entry.Level)
	}
}
