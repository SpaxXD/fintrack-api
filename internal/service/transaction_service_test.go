package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/domain"
)

// --- Helper to set up TransactionService with mocks ---

type txTestEnv struct {
	svc          *TransactionService
	txRepo       *mockTransactionRepository
	accountRepo  *mockAccountRepository
	categoryRepo *mockCategoryRepository
	userID       uuid.UUID
	accountID    uuid.UUID
	categoryID   uuid.UUID
}

func setupTxTestEnv() *txTestEnv {
	txRepo := newMockTransactionRepository()
	accountRepo := newMockAccountRepository()
	categoryRepo := newMockCategoryRepository()
	svc := NewTransactionService(txRepo, accountRepo, categoryRepo)

	userID := uuid.New()
	accountID := uuid.New()
	categoryID := uuid.New()

	// Set up an account belonging to the user
	accountRepo.accounts[accountID] = &domain.Account{
		ID:      accountID,
		UserID:  userID,
		Name:    "Test Account",
		Type:    domain.AccountTypeChecking,
		Balance: 10000, // 100.00 in cents
	}

	// Set up a category belonging to the user
	categoryRepo.categories = append(categoryRepo.categories, &domain.Category{
		ID:     categoryID,
		UserID: userID,
		Name:   "Food",
		Type:   domain.CategoryTypeExpense,
	})

	return &txTestEnv{
		svc:          svc,
		txRepo:       txRepo,
		accountRepo:  accountRepo,
		categoryRepo: categoryRepo,
		userID:       userID,
		accountID:    accountID,
		categoryID:   categoryID,
	}
}

// --- Create Tests ---

func TestTransactionService_Create_IncomeSuccess(t *testing.T) {
	env := setupTxTestEnv()
	catStr := env.categoryID.String()

	req := CreateTransactionRequest{
		Amount:      50.00,
		Type:        "income",
		AccountID:   env.accountID.String(),
		CategoryID:  &catStr,
		Description: "Salary",
		Date:        "2024-01-15",
	}

	tx, err := env.svc.Create(context.Background(), env.userID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.Amount != 5000 {
		t.Errorf("expected amount 5000 cents, got %d", tx.Amount)
	}
	if tx.Type != domain.TransactionTypeIncome {
		t.Errorf("expected type 'income', got %q", tx.Type)
	}
	if tx.AccountID != env.accountID {
		t.Errorf("expected account_id %s, got %s", env.accountID, tx.AccountID)
	}
	if tx.CategoryID == nil || *tx.CategoryID != env.categoryID {
		t.Errorf("expected category_id %s", env.categoryID)
	}
	if tx.Description != "Salary" {
		t.Errorf("expected description 'Salary', got %q", tx.Description)
	}

	// Check balance was updated: 10000 + 5000 = 15000
	acc := env.accountRepo.accounts[env.accountID]
	if acc.Balance != 15000 {
		t.Errorf("expected balance 15000, got %d", acc.Balance)
	}
}

func TestTransactionService_Create_ExpenseSuccess(t *testing.T) {
	env := setupTxTestEnv()

	req := CreateTransactionRequest{
		Amount:      25.50,
		Type:        "expense",
		AccountID:   env.accountID.String(),
		Description: "Lunch",
		Date:        "2024-01-15",
	}

	tx, err := env.svc.Create(context.Background(), env.userID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.Amount != 2550 {
		t.Errorf("expected amount 2550 cents, got %d", tx.Amount)
	}
	if tx.CategoryID != nil {
		t.Errorf("expected nil category_id, got %v", tx.CategoryID)
	}

	// Check balance was updated: 10000 - 2550 = 7450
	acc := env.accountRepo.accounts[env.accountID]
	if acc.Balance != 7450 {
		t.Errorf("expected balance 7450, got %d", acc.Balance)
	}
}

func TestTransactionService_Create_ValidationErrors(t *testing.T) {
	env := setupTxTestEnv()

	tests := []struct {
		name string
		req  CreateTransactionRequest
	}{
		{"zero amount", CreateTransactionRequest{Amount: 0, Type: "income", AccountID: env.accountID.String(), Date: "2024-01-15"}},
		{"negative amount", CreateTransactionRequest{Amount: -10, Type: "income", AccountID: env.accountID.String(), Date: "2024-01-15"}},
		{"invalid type", CreateTransactionRequest{Amount: 10, Type: "transfer", AccountID: env.accountID.String(), Date: "2024-01-15"}},
		{"empty type", CreateTransactionRequest{Amount: 10, Type: "", AccountID: env.accountID.String(), Date: "2024-01-15"}},
		{"empty account_id", CreateTransactionRequest{Amount: 10, Type: "income", AccountID: "", Date: "2024-01-15"}},
		{"invalid account_id", CreateTransactionRequest{Amount: 10, Type: "income", AccountID: "not-uuid", Date: "2024-01-15"}},
		{"empty date", CreateTransactionRequest{Amount: 10, Type: "income", AccountID: env.accountID.String(), Date: ""}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := env.svc.Create(context.Background(), env.userID, tc.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var ve *domain.ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
			}
		})
	}
}

func TestTransactionService_Create_InvalidDate(t *testing.T) {
	env := setupTxTestEnv()

	req := CreateTransactionRequest{
		Amount:    10.00,
		Type:      "income",
		AccountID: env.accountID.String(),
		Date:      "15-01-2024", // wrong format
	}

	_, err := env.svc.Create(context.Background(), env.userID, req)
	if err == nil {
		t.Fatal("expected error for invalid date, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
}

func TestTransactionService_Create_InvalidAmountDecimals(t *testing.T) {
	env := setupTxTestEnv()

	req := CreateTransactionRequest{
		Amount:    10.123, // more than 2 decimal places
		Type:      "income",
		AccountID: env.accountID.String(),
		Date:      "2024-01-15",
	}

	_, err := env.svc.Create(context.Background(), env.userID, req)
	if err == nil {
		t.Fatal("expected error for invalid decimals, got nil")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
	}
}

func TestTransactionService_Create_AccountNotFound(t *testing.T) {
	env := setupTxTestEnv()

	req := CreateTransactionRequest{
		Amount:    10.00,
		Type:      "income",
		AccountID: uuid.New().String(), // non-existent account
		Date:      "2024-01-15",
	}

	_, err := env.svc.Create(context.Background(), env.userID, req)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTransactionService_Create_AccountForbidden(t *testing.T) {
	env := setupTxTestEnv()
	otherUserID := uuid.New()

	req := CreateTransactionRequest{
		Amount:    10.00,
		Type:      "income",
		AccountID: env.accountID.String(), // belongs to env.userID, not otherUserID
		Date:      "2024-01-15",
	}

	_, err := env.svc.Create(context.Background(), otherUserID, req)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTransactionService_Create_CategoryNotFound(t *testing.T) {
	env := setupTxTestEnv()
	nonExistentCat := uuid.New().String()

	req := CreateTransactionRequest{
		Amount:     10.00,
		Type:       "income",
		AccountID:  env.accountID.String(),
		CategoryID: &nonExistentCat,
		Date:       "2024-01-15",
	}

	_, err := env.svc.Create(context.Background(), env.userID, req)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTransactionService_Create_CategoryForbidden(t *testing.T) {
	env := setupTxTestEnv()

	// Create a category belonging to another user
	otherCatID := uuid.New()
	env.categoryRepo.categories = append(env.categoryRepo.categories, &domain.Category{
		ID:     otherCatID,
		UserID: uuid.New(), // another user
		Name:   "Other",
		Type:   domain.CategoryTypeExpense,
	})

	catStr := otherCatID.String()
	req := CreateTransactionRequest{
		Amount:     10.00,
		Type:       "expense",
		AccountID:  env.accountID.String(),
		CategoryID: &catStr,
		Date:       "2024-01-15",
	}

	_, err := env.svc.Create(context.Background(), env.userID, req)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

// --- List Tests ---

func TestTransactionService_List_DefaultPagination(t *testing.T) {
	env := setupTxTestEnv()

	// Create some transactions
	for i := 0; i < 5; i++ {
		env.txRepo.transactions = append(env.txRepo.transactions, &domain.Transaction{
			ID:     uuid.New(),
			UserID: env.userID,
			Type:   domain.TransactionTypeIncome,
			Amount: 1000,
		})
	}

	txs, err := env.svc.List(context.Background(), env.userID, domain.TransactionFilter{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(txs) != 5 {
		t.Errorf("expected 5 transactions, got %d", len(txs))
	}
}

func TestTransactionService_List_LimitEnforcement(t *testing.T) {
	env := setupTxTestEnv()

	// Limit > 100 should be capped to 100
	filter := domain.TransactionFilter{Limit: 200}
	_, err := env.svc.List(context.Background(), env.userID, filter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTransactionService_List_NegativeOffset(t *testing.T) {
	env := setupTxTestEnv()

	filter := domain.TransactionFilter{Offset: -5}
	_, err := env.svc.List(context.Background(), env.userID, filter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// --- Update Tests ---

func TestTransactionService_Update_Success(t *testing.T) {
	env := setupTxTestEnv()

	// Create an existing transaction
	txID := uuid.New()
	env.txRepo.transactions = append(env.txRepo.transactions, &domain.Transaction{
		ID:        txID,
		UserID:    env.userID,
		AccountID: env.accountID,
		Type:      domain.TransactionTypeIncome,
		Amount:    5000, // 50.00
	})

	// Simulate balance after creating income: 10000 + 5000 = 15000
	env.accountRepo.accounts[env.accountID].Balance = 15000

	req := UpdateTransactionRequest{
		Amount:      30.00, // change to 30.00
		Type:        "expense",
		AccountID:   env.accountID.String(),
		Description: "Updated",
		Date:        "2024-02-01",
	}

	tx, err := env.svc.Update(context.Background(), env.userID, txID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.Amount != 3000 {
		t.Errorf("expected amount 3000, got %d", tx.Amount)
	}
	if tx.Type != domain.TransactionTypeExpense {
		t.Errorf("expected type 'expense', got %q", tx.Type)
	}
	if tx.Description != "Updated" {
		t.Errorf("expected description 'Updated', got %q", tx.Description)
	}

	// Balance: 15000 - 5000 (revert income) - 3000 (apply expense) = 7000
	acc := env.accountRepo.accounts[env.accountID]
	if acc.Balance != 7000 {
		t.Errorf("expected balance 7000, got %d", acc.Balance)
	}
}

func TestTransactionService_Update_NotFound(t *testing.T) {
	env := setupTxTestEnv()

	req := UpdateTransactionRequest{
		Amount:    10.00,
		Type:      "income",
		AccountID: env.accountID.String(),
		Date:      "2024-01-15",
	}

	_, err := env.svc.Update(context.Background(), env.userID, uuid.New(), req)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTransactionService_Update_Forbidden(t *testing.T) {
	env := setupTxTestEnv()
	otherUserID := uuid.New()

	txID := uuid.New()
	env.txRepo.transactions = append(env.txRepo.transactions, &domain.Transaction{
		ID:        txID,
		UserID:    env.userID, // belongs to env.userID
		AccountID: env.accountID,
		Type:      domain.TransactionTypeIncome,
		Amount:    1000,
	})

	// Set up account for other user
	otherAccountID := uuid.New()
	env.accountRepo.accounts[otherAccountID] = &domain.Account{
		ID:     otherAccountID,
		UserID: otherUserID,
		Name:   "Other",
		Type:   domain.AccountTypeChecking,
	}

	req := UpdateTransactionRequest{
		Amount:    10.00,
		Type:      "income",
		AccountID: otherAccountID.String(),
		Date:      "2024-01-15",
	}

	_, err := env.svc.Update(context.Background(), otherUserID, txID, req)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

// --- Delete Tests ---

func TestTransactionService_Delete_IncomeSuccess(t *testing.T) {
	env := setupTxTestEnv()

	txID := uuid.New()
	env.txRepo.transactions = append(env.txRepo.transactions, &domain.Transaction{
		ID:        txID,
		UserID:    env.userID,
		AccountID: env.accountID,
		Type:      domain.TransactionTypeIncome,
		Amount:    3000, // 30.00
	})

	// Simulate balance after income: 10000 + 3000 = 13000
	env.accountRepo.accounts[env.accountID].Balance = 13000

	err := env.svc.Delete(context.Background(), env.userID, txID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Balance should revert: 13000 - 3000 = 10000
	acc := env.accountRepo.accounts[env.accountID]
	if acc.Balance != 10000 {
		t.Errorf("expected balance 10000, got %d", acc.Balance)
	}
}

func TestTransactionService_Delete_ExpenseSuccess(t *testing.T) {
	env := setupTxTestEnv()

	txID := uuid.New()
	env.txRepo.transactions = append(env.txRepo.transactions, &domain.Transaction{
		ID:        txID,
		UserID:    env.userID,
		AccountID: env.accountID,
		Type:      domain.TransactionTypeExpense,
		Amount:    2000, // 20.00
	})

	// Simulate balance after expense: 10000 - 2000 = 8000
	env.accountRepo.accounts[env.accountID].Balance = 8000

	err := env.svc.Delete(context.Background(), env.userID, txID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Balance should revert: 8000 + 2000 = 10000
	acc := env.accountRepo.accounts[env.accountID]
	if acc.Balance != 10000 {
		t.Errorf("expected balance 10000, got %d", acc.Balance)
	}
}

func TestTransactionService_Delete_NotFound(t *testing.T) {
	env := setupTxTestEnv()

	err := env.svc.Delete(context.Background(), env.userID, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTransactionService_Delete_Forbidden(t *testing.T) {
	env := setupTxTestEnv()

	txID := uuid.New()
	env.txRepo.transactions = append(env.txRepo.transactions, &domain.Transaction{
		ID:        txID,
		UserID:    env.userID,
		AccountID: env.accountID,
		Type:      domain.TransactionTypeIncome,
		Amount:    1000,
	})

	otherUserID := uuid.New()
	err := env.svc.Delete(context.Background(), otherUserID, txID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
