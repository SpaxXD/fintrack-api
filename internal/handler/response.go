package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/middleware"
)

// ErrorResponse is the standard error response format for the API.
type ErrorResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Details   interface{} `json:"details,omitempty"`
}

// ValidationDetail represents a single field validation error.
type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// writeError maps a domain error to an HTTP error response and writes it.
// Internal details are never exposed to the client.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code := mapDomainError(err)

	resp := ErrorResponse{
		Code:      code,
		RequestID: middleware.GetRequestID(r.Context()),
	}

	// For 500 errors, use a generic message to avoid leaking internals.
	if status == http.StatusInternalServerError {
		resp.Message = "an unexpected error occurred"
	} else {
		resp.Message = err.Error()
	}

	// If the error is a ValidationError, include field details.
	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		details := make([]ValidationDetail, 0, len(ve.Fields))
		for _, f := range ve.Fields {
			details = append(details, ValidationDetail{
				Field:   f.Field,
				Message: f.Message,
			})
		}
		resp.Details = details
	}

	writeJSON(w, status, resp)
}

// mapDomainError maps a domain error to an HTTP status code and error code string.
func mapDomainError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return http.StatusUnprocessableEntity, "unprocessable_entity"
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, domain.ErrTokenExpired):
		return http.StatusUnauthorized, "token_expired"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "conflict"
	case errors.Is(err, domain.ErrRateLimited):
		return http.StatusTooManyRequests, "rate_limited"
	case errors.Is(err, domain.ErrInvalidAmount):
		return http.StatusUnprocessableEntity, "unprocessable_entity"
	default:
		return http.StatusInternalServerError, "internal_error"
	}
}
