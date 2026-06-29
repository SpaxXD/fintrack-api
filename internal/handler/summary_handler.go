package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/muriloabranches/fintrack-api/internal/middleware"
	"github.com/muriloabranches/fintrack-api/internal/service"
)

// SummaryHandler handles HTTP requests for financial summary endpoints.
type SummaryHandler struct {
	svc service.SummaryService
}

// NewSummaryHandler creates a new SummaryHandler with the given SummaryService.
func NewSummaryHandler(svc service.SummaryService) *SummaryHandler {
	return &SummaryHandler{svc: svc}
}

// GetSummary handles GET /api/v1/summary.
// Query params: date_from (YYYY-MM-DD, required), date_to (YYYY-MM-DD, required).
func (h *SummaryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	dateFrom, dateTo, ok := parseDateRange(w, r)
	if !ok {
		return
	}

	summary, err := h.svc.GetSummary(r.Context(), userID, dateFrom, dateTo)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// GetCategorySummary handles GET /api/v1/summary/categories.
// Query params: date_from (YYYY-MM-DD, required), date_to (YYYY-MM-DD, required).
func (h *SummaryHandler) GetCategorySummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	dateFrom, dateTo, ok := parseDateRange(w, r)
	if !ok {
		return
	}

	categories, err := h.svc.GetCategorySummary(r.Context(), userID, dateFrom, dateTo)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

// GetMonthlyTrend handles GET /api/v1/summary/trend.
// Query params: year (4-digit numeric, required, 2000-2100).
func (h *SummaryHandler) GetMonthlyTrend(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "year is required",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "year must be a numeric value",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	if year < 2000 || year > 2100 {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "year must be between 2000 and 2100",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return
	}

	trends, err := h.svc.GetMonthlyTrend(r.Context(), userID, year)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, trends)
}

// parseDateRange extracts and validates date_from and date_to query params.
// Returns parsed times and true on success, or writes an error response and returns false.
func parseDateRange(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	dateFromStr := r.URL.Query().Get("date_from")
	dateToStr := r.URL.Query().Get("date_to")

	if dateFromStr == "" {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "date_from is required",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return time.Time{}, time.Time{}, false
	}

	if dateToStr == "" {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "date_to is required",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return time.Time{}, time.Time{}, false
	}

	const dateLayout = "2006-01-02"

	dateFrom, err := time.Parse(dateLayout, dateFromStr)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "date_from must be in YYYY-MM-DD format",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return time.Time{}, time.Time{}, false
	}

	dateTo, err := time.Parse(dateLayout, dateToStr)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "date_to must be in YYYY-MM-DD format",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return time.Time{}, time.Time{}, false
	}

	if dateFrom.After(dateTo) {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Code:      "unprocessable_entity",
			Message:   "date_from must not be after date_to",
			RequestID: middleware.GetRequestID(r.Context()),
		})
		return time.Time{}, time.Time{}, false
	}

	return dateFrom, dateTo, true
}
