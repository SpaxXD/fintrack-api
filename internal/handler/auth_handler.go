package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/service"
)

// UserResponse represents the public user profile returned on registration.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// refreshRequest represents the JSON body for the refresh endpoint.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthHandler handles authentication-related HTTP endpoints.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler with the given AuthService.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/v1/auth/register.
// It decodes the request body, delegates to AuthService, and returns the created user profile.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, err)
		return
	}

	user, err := h.authService.Register(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	resp := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Login handles POST /api/v1/auth/login.
// It decodes credentials, delegates to AuthService, and returns a token pair.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, err)
		return
	}

	pair, err := h.authService.Login(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, pair)
}

// Refresh handles POST /api/v1/auth/refresh.
// It decodes the refresh token, delegates to AuthService, and returns a new token pair.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, err)
		return
	}

	pair, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, pair)
}
