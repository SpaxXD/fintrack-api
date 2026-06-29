package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/middleware"
)

// mockSummaryService implements service.SummaryService for testing.
type mockSummaryService struct {
	getSummaryFn         func(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) (*domain.Summary, error)
	getCategorySummaryFn func(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]domain.CategorySummary, error)
	getMonthlyTrendFn    func(ctx context.Context, userID uuid.UUID, year int) ([]domain.MonthlyTrend, error)
}

func (m *mockSummaryService) GetSummary(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) (*domain.Summary, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, userID, dateFrom, dateTo)
	}
	return &domain.Summary{}, nil
}

func (m *mockSummaryService) GetCategorySummary(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]domain.CategorySummary, error) {
	if m.getCategorySummaryFn != nil {
		return m.getCategorySummaryFn(ctx, userID, dateFrom, dateTo)
	}
	return []domain.CategorySummary{}, nil
}

func (m *mockSummaryService) GetMonthlyTrend(ctx context.Context, userID uuid.UUID, year int) ([]domain.MonthlyTrend, error) {
	if m.getMonthlyTrendFn != nil {
		return m.getMonthlyTrendFn(ctx, userID, year)
	}
	return []domain.MonthlyTrend{}, nil
}

func newSummaryRequest(method, path string, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")
	r := httptest.NewRequest(method, path, nil)
	return r.WithContext(ctx)
}

func TestSummaryHandler_GetSummary(t *testing.T) {
	testUserID := uuid.New()

	t.Run("returns summary for valid date range", func(t *testing.T) {
		svc := &mockSummaryService{
			getSummaryFn: func(_ context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) (*domain.Summary, error) {
				if userID != testUserID {
					t.Errorf("unexpected userID: %v", userID)
				}
				return &domain.Summary{
					TotalIncome:  150000,
					TotalExpense: 80000,
					NetBalance:   70000,
				}, nil
			},
		}
		h := NewSummaryHandler(svc)
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary?date_from=2024-01-01&date_to=2024-01-31", testUserID)

		h.GetSummary(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var resp domain.Summary
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if resp.TotalIncome != 150000 {
			t.Errorf("expected total_income=150000, got %d", resp.TotalIncome)
		}
		if resp.TotalExpense != 80000 {
			t.Errorf("expected total_expense=80000, got %d", resp.TotalExpense)
		}
		if resp.NetBalance != 70000 {
			t.Errorf("expected net_balance=70000, got %d", resp.NetBalance)
		}
	})

	t.Run("returns 422 when date_from is missing", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary?date_to=2024-01-31", testUserID)

		h.GetSummary(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp.Message != "date_from is required" {
			t.Errorf("expected 'date_from is required', got %q", resp.Message)
		}
	})

	t.Run("returns 422 when date_to is missing", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary?date_from=2024-01-01", testUserID)

		h.GetSummary(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp.Message != "date_to is required" {
			t.Errorf("expected 'date_to is required', got %q", resp.Message)
		}
	})

	t.Run("returns 422 for invalid date_from format", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary?date_from=01-01-2024&date_to=2024-01-31", testUserID)

		h.GetSummary(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp.Message != "date_from must be in YYYY-MM-DD format" {
			t.Errorf("unexpected message: %q", resp.Message)
		}
	})

	t.Run("returns 422 for invalid date_to format", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary?date_from=2024-01-01&date_to=invalid", testUserID)

		h.GetSummary(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp.Message != "date_to must be in YYYY-MM-DD format" {
			t.Errorf("unexpected message: %q", resp.Message)
		}
	})

	t.Run("returns 422 when date_from is after date_to", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary?date_from=2024-02-01&date_to=2024-01-01", testUserID)

		h.GetSummary(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp.Message != "date_from must not be after date_to" {
			t.Errorf("unexpected message: %q", resp.Message)
		}
	})

	t.Run("returns error from service", func(t *testing.T) {
		svc := &mockSummaryService{
			getSummaryFn: func(_ context.Context, _ uuid.UUID, _, _ time.Time) (*domain.Summary, error) {
				return nil, errors.New("db connection failed")
			},
		}
		h := NewSummaryHandler(svc)
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary?date_from=2024-01-01&date_to=2024-01-31", testUserID)

		h.GetSummary(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestSummaryHandler_GetCategorySummary(t *testing.T) {
	testUserID := uuid.New()
	catID := uuid.New()

	t.Run("returns category summary for valid date range", func(t *testing.T) {
		svc := &mockSummaryService{
			getCategorySummaryFn: func(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]domain.CategorySummary, error) {
				return []domain.CategorySummary{
					{CategoryID: catID, CategoryName: "Food", Type: "expense", Total: 45000},
				}, nil
			},
		}
		h := NewSummaryHandler(svc)
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/categories?date_from=2024-01-01&date_to=2024-12-31", testUserID)

		h.GetCategorySummary(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var resp []domain.CategorySummary
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if len(resp) != 1 {
			t.Fatalf("expected 1 category, got %d", len(resp))
		}
		if resp[0].CategoryName != "Food" {
			t.Errorf("expected 'Food', got %q", resp[0].CategoryName)
		}
		if resp[0].Total != 45000 {
			t.Errorf("expected total=45000, got %d", resp[0].Total)
		}
	})

	t.Run("returns 422 when date_from missing", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/categories?date_to=2024-01-31", testUserID)

		h.GetCategorySummary(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
	})
}

func TestSummaryHandler_GetMonthlyTrend(t *testing.T) {
	testUserID := uuid.New()

	t.Run("returns monthly trend for valid year", func(t *testing.T) {
		svc := &mockSummaryService{
			getMonthlyTrendFn: func(_ context.Context, _ uuid.UUID, year int) ([]domain.MonthlyTrend, error) {
				if year != 2024 {
					t.Errorf("expected year=2024, got %d", year)
				}
				trends := make([]domain.MonthlyTrend, 12)
				for i := range trends {
					trends[i].Month = i + 1
				}
				trends[0].TotalIncome = 500000
				trends[0].TotalExpense = 300000
				return trends, nil
			},
		}
		h := NewSummaryHandler(svc)
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/trend?year=2024", testUserID)

		h.GetMonthlyTrend(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var resp []domain.MonthlyTrend
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if len(resp) != 12 {
			t.Fatalf("expected 12 months, got %d", len(resp))
		}
		if resp[0].TotalIncome != 500000 {
			t.Errorf("expected jan income=500000, got %d", resp[0].TotalIncome)
		}
	})

	t.Run("returns 422 when year is missing", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/trend", testUserID)

		h.GetMonthlyTrend(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp.Message != "year is required" {
			t.Errorf("expected 'year is required', got %q", resp.Message)
		}
	})

	t.Run("returns 422 when year is not numeric", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/trend?year=abc", testUserID)

		h.GetMonthlyTrend(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp.Message != "year must be a numeric value" {
			t.Errorf("expected 'year must be a numeric value', got %q", resp.Message)
		}
	})

	t.Run("returns 422 when year is below 2000", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/trend?year=1999", testUserID)

		h.GetMonthlyTrend(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp.Message != "year must be between 2000 and 2100" {
			t.Errorf("unexpected message: %q", resp.Message)
		}
	})

	t.Run("returns 422 when year is above 2100", func(t *testing.T) {
		h := NewSummaryHandler(&mockSummaryService{})
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/trend?year=2101", testUserID)

		h.GetMonthlyTrend(w, r)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", w.Code)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp.Message != "year must be between 2000 and 2100" {
			t.Errorf("unexpected message: %q", resp.Message)
		}
	})

	t.Run("accepts boundary year 2000", func(t *testing.T) {
		svc := &mockSummaryService{
			getMonthlyTrendFn: func(_ context.Context, _ uuid.UUID, year int) ([]domain.MonthlyTrend, error) {
				if year != 2000 {
					t.Errorf("expected year=2000, got %d", year)
				}
				return make([]domain.MonthlyTrend, 12), nil
			},
		}
		h := NewSummaryHandler(svc)
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/trend?year=2000", testUserID)

		h.GetMonthlyTrend(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("accepts boundary year 2100", func(t *testing.T) {
		svc := &mockSummaryService{
			getMonthlyTrendFn: func(_ context.Context, _ uuid.UUID, year int) ([]domain.MonthlyTrend, error) {
				if year != 2100 {
					t.Errorf("expected year=2100, got %d", year)
				}
				return make([]domain.MonthlyTrend, 12), nil
			},
		}
		h := NewSummaryHandler(svc)
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/trend?year=2100", testUserID)

		h.GetMonthlyTrend(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("returns error from service", func(t *testing.T) {
		svc := &mockSummaryService{
			getMonthlyTrendFn: func(_ context.Context, _ uuid.UUID, _ int) ([]domain.MonthlyTrend, error) {
				return nil, errors.New("db error")
			},
		}
		h := NewSummaryHandler(svc)
		w := httptest.NewRecorder()
		r := newSummaryRequest(http.MethodGet, "/api/v1/summary/trend?year=2024", testUserID)

		h.GetMonthlyTrend(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}
