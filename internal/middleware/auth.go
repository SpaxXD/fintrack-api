package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Auth returns a chi-compatible middleware that validates JWT tokens from the
// Authorization header and injects the authenticated user ID into the request
// context. The secret parameter is the HMAC-SHA256 signing key.
func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				writeAuthError(w, r, "unauthorized", "missing or malformed authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			segments := strings.Split(tokenStr, ".")
			if len(segments) != 3 {
				writeAuthError(w, r, "unauthorized", "invalid token structure")
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(secret), nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					writeAuthError(w, r, "token_expired", "access token has expired")
					return
				}
				writeAuthError(w, r, "unauthorized", "invalid or malformed token")
				return
			}

			if !token.Valid {
				writeAuthError(w, r, "unauthorized", "invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeAuthError(w, r, "unauthorized", "invalid token claims")
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				writeAuthError(w, r, "unauthorized", "missing user identifier in token")
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				writeAuthError(w, r, "unauthorized", "invalid user identifier in token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// authErrorResponse is the JSON structure returned on authentication failures.
type authErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// writeAuthError writes a 401 JSON error response with the given code and message.
func writeAuthError(w http.ResponseWriter, r *http.Request, code, message string) {
	reqID := GetRequestID(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	resp := authErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: reqID,
	}

	_ = json.NewEncoder(w).Encode(resp)
}
