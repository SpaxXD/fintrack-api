package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// HeaderXRequestID is the HTTP header name used to propagate request IDs.
const HeaderXRequestID = "X-Request-ID"

// RequestID is a chi-compatible middleware that generates a UUID v4 for each
// incoming request, stores it in the request context, and sets it in the
// X-Request-ID response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()

		// Set the response header.
		w.Header().Set(HeaderXRequestID, id)

		// Store the request ID in the context using the shared key.
		ctx := context.WithValue(r.Context(), RequestIDKey, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
