package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/middleware"
	"github.com/muriloabranches/fintrack-api/internal/service"
)

// TransactionHandler handles HTTP requests for transaction management.
type TransactionHandler struct {
	svc *service.TransactionService
}

// NewTransactionHandler creates a new TransactionHandler with the given service.
func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

// Create handles POST /api/v1/transactions
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, r, domain.ErrUnauthorized)
		return
	}

	var req service.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, err)
		return
	}

	transaction, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, transaction)
}

// List handles GET /api/v1/transactions
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, r, domain.ErrUnauthorized)
		return
	}

	filter, err := parseTransactionFilter(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	transactions, err := h.svc.List(r.Context(), userID, filter)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, transactions)
}

// Update handles PUT /api/v1/transactions/{id}
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, r, domain.ErrUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	txID, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, r, fmt.Errorf("invalid transaction id: %w", domain.ErrValidation))
		return
	}

	var req service.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, err)
		return
	}

	transaction, err := h.svc.Update(r.Context(), userID, txID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, transaction)
}

// Delete handles DELETE /api/v1/transactions/{id}
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, r, domain.ErrUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	txID, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, r, fmt.Errorf("invalid transaction id: %w", domain.ErrValidation))
		return
	}

	if err := h.svc.Delete(r.Context(), userID, txID); err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseTransactionFilter extracts filter parameters from query string.
func parseTransactionFilter(r *http.Request) (domain.TransactionFilter, error) {
	var filter domain.TransactionFilter

	// account_id
	if v := r.URL.Query().Get("account_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return filter, fmt.Errorf("invalid account_id: %w", domain.ErrValidation)
		}
		filter.AccountID = &id
	}

	// category_id
	if v := r.URL.Query().Get("category_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return filter, fmt.Errorf("invalid category_id: %w", domain.ErrValidation)
		}
		filter.CategoryID = &id
	}

	// type
	if v := r.URL.Query().Get("type"); v != "" {
		txType := domain.TransactionType(v)
		if txType != domain.TransactionTypeIncome && txType != domain.TransactionTypeExpense {
			return filter, fmt.Errorf("invalid type, must be 'income' or 'expense': %w", domain.ErrValidation)
		}
		filter.Type = &txType
	}

	// date_from (YYYY-MM-DD)
	if v := r.URL.Query().Get("date_from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, fmt.Errorf("invalid date_from, expected YYYY-MM-DD: %w", domain.ErrValidation)
		}
		filter.DateFrom = &t
	}

	// date_to (YYYY-MM-DD)
	if v := r.URL.Query().Get("date_to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, fmt.Errorf("invalid date_to, expected YYYY-MM-DD: %w", domain.ErrValidation)
		}
		filter.DateTo = &t
	}

	// limit (default 20)
	if v := r.URL.Query().Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil || limit < 0 {
			return filter, fmt.Errorf("invalid limit: %w", domain.ErrValidation)
		}
		filter.Limit = limit
	} else {
		filter.Limit = 20
	}

	// offset (default 0)
	if v := r.URL.Query().Get("offset"); v != "" {
		offset, err := strconv.Atoi(v)
		if err != nil || offset < 0 {
			return filter, fmt.Errorf("invalid offset: %w", domain.ErrValidation)
		}
		filter.Offset = offset
	}

	return filter, nil
}
