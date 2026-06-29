package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/middleware"
	"github.com/muriloabranches/fintrack-api/internal/service"
)

// CategoryHandler handles HTTP requests for category management.
type CategoryHandler struct {
	svc *service.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler with the given CategoryService.
func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// Create handles POST /api/v1/categories.
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req service.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "invalid request body",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	category, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, category)
}

// List handles GET /api/v1/categories.
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	categories, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

// Update handles PUT /api/v1/categories/{id}.
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	idParam := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "invalid category id",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	var req service.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "invalid request body",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	category, err := h.svc.Update(r.Context(), userID, categoryID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, category)
}

// Delete handles DELETE /api/v1/categories/{id}.
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	idParam := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "invalid category id",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	if err := h.svc.Delete(r.Context(), userID, categoryID); err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
