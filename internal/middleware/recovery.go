package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

// recoveryErrorResponse is the JSON structure returned on panic recovery.
// It intentionally omits internal details to prevent information leakage.
type recoveryErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// Recovery is a chi-compatible middleware that catches panics from downstream
// handlers, logs the panic value and stack trace using zerolog at error level,
// and returns an HTTP 500 response with a generic error body that never exposes
// internal details.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				requestID := GetRequestID(r.Context())

				log.Error().
					Interface("panic", rec).
					Str("stack_trace", string(stack)).
					Str("request_id", requestID).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("panic recovered")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				resp := recoveryErrorResponse{
					Code:      "internal_error",
					Message:   "An internal error occurred",
					RequestID: requestID,
				}

				_ = json.NewEncoder(w).Encode(resp)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
