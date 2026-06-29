package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/middleware"
	"github.com/muriloabranches/fintrack-api/internal/service"
)

// mockAccountService is a mock implementation of service.AccountService.
type mockAccountService struct {
	createFn func(ctx context.Context, userID uuid.UUID, req service.CreateAccountRequest) (*domain.Account, error)
	listFn   func(ctx context.Context, userID uuid.UUID) ([]domain.Account, error)
	updateFn func(ctx context.Context, userID uuid.UUID, id uuid.UUID, req service.UpdateAccountRequest) (*domain.Account, error)
	deleteFn func(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}

func (m *mockAccountService) Create(ctx context.Context, userID uuid.UUID, req service.CreateAccountRequest) (*domain.Account, error) {
	return m.createFn(ctx, userID, req)
}

func (m *mockAccountService) List(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	return m.listFn(ctx, userID)
}

func (m *mockAccountService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, req service.UpdateAccountRequest) (*domain.Account, error) {
	return m.updateFn(ctx, userID, id, req)
}

func (m *mockAccountService) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return m.deleteFn(ctx, userID, id)
}

func newTestContext(userID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-req-id")
	return ctx
}

func TestAccountHandler_Create(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		account := &domain.Account{
			ID:     uuid.New(),
			UserID: userID,
			Name:   "Checking",
			Type:   domain.AccountTypeChecking,
		}

		svc := &mockAccountService{
			createFn: func(_ context.Context, uid uuid.UUID, req service.CreateAccountRequest) (*domain.Account, error) {
				if uid != userID {
					t.Errorf("expected userID %v, got %v", userID, uid)
				}
				if req.Name != "Checking" {
					t.Errorf("expected name Checking, got %q", req.Name)
				}
				return account, nil
			},
		}

		handler := NewAccountHandler(svc)
		body := `{"name":"Checking","type":"checking"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBufferString(body))
		req = req.WithContext(newTestContext(userID))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("invalid body returns 422", func(t *testing.T) {
		svc := &mockAccountService{}
		handler := NewAccountHandler(svc)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBufferString("invalid json"))
		req = req.WithContext(newTestContext(userID))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("service error propagates", func(t *testing.T) {
		svc := &mockAccountService{
			createFn: func(_ context.Context, _ uuid.UUID, _ service.CreateAccountRequest) (*domain.Account, error) {
				return nil, domain.ErrValidation
			},
		}

		handler := NewAccountHandler(svc)
		body := `{"name":"","type":"invalid"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBufferString(body))
		req = req.WithContext(newTestContext(userID))
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})
}

func TestAccountHandler_List(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		accounts := []domain.Account{
			{ID: uuid.New(), UserID: userID, Name: "Savings", Type: domain.AccountTypeSavings},
		}

		svc := &mockAccountService{
			listFn: func(_ context.Context, uid uuid.UUID) ([]domain.Account, error) {
				if uid != userID {
					t.Errorf("expected userID %v, got %v", userID, uid)
				}
				return accounts, nil
			},
		}

		handler := NewAccountHandler(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
		req = req.WithContext(newTestContext(userID))
		w := httptest.NewRecorder()

		handler.List(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result []domain.Account
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 account, got %d", len(result))
		}
	})

	t.Run("service error propagates", func(t *testing.T) {
		svc := &mockAccountService{
			listFn: func(_ context.Context, _ uuid.UUID) ([]domain.Account, error) {
				return nil, domain.ErrNotFound
			},
		}

		handler := NewAccountHandler(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
		req = req.WithContext(newTestContext(userID))
		w := httptest.NewRecorder()

		handler.List(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestAccountHandler_Update(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()

	t.Run("success", func(t *testing.T) {
		account := &domain.Account{
			ID:     accountID,
			UserID: userID,
			Name:   "Updated",
			Type:   domain.AccountTypeSavings,
		}

		svc := &mockAccountService{
			updateFn: func(_ context.Context, uid, id uuid.UUID, req service.UpdateAccountRequest) (*domain.Account, error) {
				if uid != userID {
					t.Errorf("expected userID %v, got %v", userID, uid)
				}
				if id != accountID {
					t.Errorf("expected accountID %v, got %v", accountID, id)
				}
				return account, nil
			},
		}

		handler := NewAccountHandler(svc)
		body := `{"name":"Updated","type":"savings"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/accounts/"+accountID.String(), bytes.NewBufferString(body))

		// Set up chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx := context.WithValue(newTestContext(userID), chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Update(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("invalid ID returns 422", func(t *testing.T) {
		svc := &mockAccountService{}
		handler := NewAccountHandler(svc)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/accounts/not-a-uuid", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "not-a-uuid")
		ctx := context.WithValue(newTestContext(userID), chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Update(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}

		var resp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Code != "unprocessable_entity" {
			t.Errorf("expected code unprocessable_entity, got %q", resp.Code)
		}
	})

	t.Run("invalid body returns 422", func(t *testing.T) {
		svc := &mockAccountService{}
		handler := NewAccountHandler(svc)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/accounts/"+accountID.String(), bytes.NewBufferString("bad"))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx := context.WithValue(newTestContext(userID), chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Update(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("forbidden error propagates", func(t *testing.T) {
		svc := &mockAccountService{
			updateFn: func(_ context.Context, _, _ uuid.UUID, _ service.UpdateAccountRequest) (*domain.Account, error) {
				return nil, domain.ErrForbidden
			},
		}

		handler := NewAccountHandler(svc)
		body := `{"name":"Test","type":"checking"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/accounts/"+accountID.String(), bytes.NewBufferString(body))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx := context.WithValue(newTestContext(userID), chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Update(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestAccountHandler_Delete(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()

	t.Run("success returns 204", func(t *testing.T) {
		svc := &mockAccountService{
			deleteFn: func(_ context.Context, uid, id uuid.UUID) error {
				if uid != userID {
					t.Errorf("expected userID %v, got %v", userID, uid)
				}
				if id != accountID {
					t.Errorf("expected accountID %v, got %v", accountID, id)
				}
				return nil
			},
		}

		handler := NewAccountHandler(svc)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/"+accountID.String(), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx := context.WithValue(newTestContext(userID), chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
		}
	})

	t.Run("invalid ID returns 422", func(t *testing.T) {
		svc := &mockAccountService{}
		handler := NewAccountHandler(svc)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/bad-id", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "bad-id")
		ctx := context.WithValue(newTestContext(userID), chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("not found error propagates", func(t *testing.T) {
		svc := &mockAccountService{
			deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
				return domain.ErrNotFound
			},
		}

		handler := NewAccountHandler(svc)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/"+accountID.String(), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx := context.WithValue(newTestContext(userID), chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}
