package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/domain"
)

// mockAccountRepository is an in-memory mock for domain.AccountRepository.
type mockAccountRepository struct {
	accounts map[uuid.UUID]*domain.Account
	createFn func(ctx context.Context, account *domain.Account) error
}

func newMockAccountRepository() *mockAccountRepository {
	return &mockAccountRepository{
		accounts: make(map[uuid.UUID]*domain.Account),
	}
}

func (m *mockAccountRepository) Create(ctx context.Context, account *domain.Account) error {
	if m.createFn != nil {
		return m.createFn(ctx, account)
	}
	m.accounts[account.ID] = account
	return nil
}

func (m *mockAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	acc, ok := m.accounts[id]
	if !ok {
		return nil, nil
	}
	return acc, nil
}

func (m *mockAccountRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	var result []domain.Account
	for _, acc := range m.accounts {
		if acc.UserID == userID && acc.DeletedAt == nil {
			result = append(result, *acc)
		}
	}
	return result, nil
}

func (m *mockAccountRepository) Update(ctx context.Context, account *domain.Account) error {
	m.accounts[account.ID] = account
	return nil
}

func (m *mockAccountRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	delete(m.accounts, id)
	return nil
}

func (m *mockAccountRepository) UpdateBalance(ctx context.Context, id uuid.UUID, delta int64) error {
	acc, ok := m.accounts[id]
	if !ok {
		return domain.ErrNotFound
	}
	acc.Balance += delta
	return nil
}

func TestAccountService_Create_Success(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	userID := uuid.New()
	balance := int64(1000)
	req := CreateAccountRequest{
		Name:           "My Checking",
		Type:           "checking",
		InitialBalance: &balance,
	}

	account, err := svc.Create(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if account.Name != "My Checking" {
		t.Errorf("expected name 'My Checking', got %q", account.Name)
	}
	if account.Type != domain.AccountTypeChecking {
		t.Errorf("expected type 'checking', got %q", account.Type)
	}
	if account.Balance != 1000 {
		t.Errorf("expected balance 1000, got %d", account.Balance)
	}
	if account.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, account.UserID)
	}
}

func TestAccountService_Create_DefaultBalance(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	userID := uuid.New()
	req := CreateAccountRequest{
		Name: "Cash",
		Type: "cash",
	}

	account, err := svc.Create(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if account.Balance != 0 {
		t.Errorf("expected default balance 0, got %d", account.Balance)
	}
}

func TestAccountService_Create_ValidationError(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	tests := []struct {
		name string
		req  CreateAccountRequest
	}{
		{"empty name", CreateAccountRequest{Name: "", Type: "checking"}},
		{"invalid type", CreateAccountRequest{Name: "Valid", Type: "invalid_type"}},
		{"missing type", CreateAccountRequest{Name: "Valid", Type: ""}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), uuid.New(), tc.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			var ve *domain.ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
			}
		})
	}
}

func TestAccountService_List(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	userID := uuid.New()
	otherUserID := uuid.New()

	// Create accounts for our user
	repo.accounts[uuid.New()] = &domain.Account{ID: uuid.New(), UserID: userID, Name: "A1"}
	repo.accounts[uuid.New()] = &domain.Account{ID: uuid.New(), UserID: userID, Name: "A2"}
	// Create account for another user
	repo.accounts[uuid.New()] = &domain.Account{ID: uuid.New(), UserID: otherUserID, Name: "Other"}

	accounts, err := svc.List(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}
}

func TestAccountService_Update_Success(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	userID := uuid.New()
	accountID := uuid.New()
	repo.accounts[accountID] = &domain.Account{
		ID:     accountID,
		UserID: userID,
		Name:   "Old Name",
		Type:   domain.AccountTypeChecking,
	}

	req := UpdateAccountRequest{
		Name: "New Name",
		Type: "savings",
	}

	account, err := svc.Update(context.Background(), userID, accountID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if account.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %q", account.Name)
	}
	if account.Type != domain.AccountTypeSavings {
		t.Errorf("expected type 'savings', got %q", account.Type)
	}
}

func TestAccountService_Update_NotFound(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	req := UpdateAccountRequest{Name: "Name", Type: "checking"}
	_, err := svc.Update(context.Background(), uuid.New(), uuid.New(), req)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAccountService_Update_Forbidden(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	accountID := uuid.New()
	repo.accounts[accountID] = &domain.Account{
		ID:     accountID,
		UserID: ownerID,
		Name:   "Owner's Account",
		Type:   domain.AccountTypeChecking,
	}

	req := UpdateAccountRequest{Name: "Hacked", Type: "checking"}
	_, err := svc.Update(context.Background(), otherUserID, accountID, req)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAccountService_Update_ValidationError(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	req := UpdateAccountRequest{Name: "", Type: "checking"}
	_, err := svc.Update(context.Background(), uuid.New(), uuid.New(), req)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
}

func TestAccountService_Delete_Success(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	userID := uuid.New()
	accountID := uuid.New()
	repo.accounts[accountID] = &domain.Account{
		ID:     accountID,
		UserID: userID,
		Name:   "To Delete",
		Type:   domain.AccountTypeCash,
	}

	err := svc.Delete(context.Background(), userID, accountID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should be gone from the mock
	if _, exists := repo.accounts[accountID]; exists {
		t.Error("expected account to be deleted from repository")
	}
}

func TestAccountService_Delete_NotFound(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAccountService_Delete_Forbidden(t *testing.T) {
	repo := newMockAccountRepository()
	svc := NewAccountService(repo)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	accountID := uuid.New()
	repo.accounts[accountID] = &domain.Account{
		ID:     accountID,
		UserID: ownerID,
		Name:   "Owner's Account",
		Type:   domain.AccountTypeChecking,
	}

	err := svc.Delete(context.Background(), otherUserID, accountID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
