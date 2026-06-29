package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter for middleware compatibility.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// Logger returns a middleware that logs HTTP requests using zerolog in structured JSON.
// It records method, path, status code, duration in milliseconds, and request ID.
// Requests resulting in status >= 500 are logged at error level; others at info level.
func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := newResponseWriter(w)

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			durationMs := float64(duration.Nanoseconds()) / float64(time.Millisecond)

			requestID := GetRequestID(r.Context())

			var event *zerolog.Event
			if wrapped.statusCode >= http.StatusInternalServerError {
				event = logger.Error()

				userID := GetUserID(r.Context())
				if userID != [16]byte{} {
					event = event.Str("user_id", userID.String())
				}
			} else {
				event = logger.Info()
			}

			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", wrapped.statusCode).
				Float64("duration_ms", durationMs).
				Str("request_id", requestID).
				Msg("request completed")
		})
	}
}
