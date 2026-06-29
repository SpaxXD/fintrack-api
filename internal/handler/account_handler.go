package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/middleware"
	"github.com/muriloabranches/fintrack-api/internal/service"
)

// AccountHandler handles HTTP requests for account management.
type AccountHandler struct {
	svc service.AccountService
}

// NewAccountHandler creates a new AccountHandler with the given AccountService.
func NewAccountHandler(svc service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

// Create handles POST /api/v1/accounts.
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req service.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "invalid request body",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	account, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, account)
}

// List handles GET /api/v1/accounts.
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	accounts, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, accounts)
}

// Update handles PUT /api/v1/accounts/{id}.
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	idParam := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(idParam)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "invalid account id",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	var req service.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "invalid request body",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	account, err := h.svc.Update(r.Context(), userID, accountID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, account)
}

// Delete handles DELETE /api/v1/accounts/{id}.
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	idParam := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(idParam)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "invalid account id",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	if err := h.svc.Delete(r.Context(), userID, accountID); err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
