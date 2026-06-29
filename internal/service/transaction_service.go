package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/validator"
)

// CreateTransactionRequest holds input data for creating a transaction.
type CreateTransactionRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Type        string  `json:"type" validate:"required,oneof=income expense"`
	AccountID   string  `json:"account_id" validate:"required,uuid"`
	CategoryID  *string `json:"category_id" validate:"omitempty,uuid"`
	Description string  `json:"description"`
	Date        string  `json:"date" validate:"required"` // YYYY-MM-DD
}

// UpdateTransactionRequest holds input data for updating a transaction.
type UpdateTransactionRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Type        string  `json:"type" validate:"required,oneof=income expense"`
	AccountID   string  `json:"account_id" validate:"required,uuid"`
	CategoryID  *string `json:"category_id" validate:"omitempty,uuid"`
	Description string  `json:"description"`
	Date        string  `json:"date" validate:"required"`
}

// TransactionService handles business logic for transaction management.
type TransactionService struct {
	txRepo      domain.TransactionRepository
	accountRepo domain.AccountRepository
	categoryRepo domain.CategoryRepository
}

// NewTransactionService creates a new TransactionService with the given repositories.
func NewTransactionService(
	txRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	categoryRepo domain.CategoryRepository,
) *TransactionService {
	return &TransactionService{
		txRepo:      txRepo,
		accountRepo: accountRepo,
		categoryRepo: categoryRepo,
	}
}

// Create validates the input, creates a transaction, and updates the account balance atomically.
func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, req CreateTransactionRequest) (*domain.Transaction, error) {
	if ve := validator.Validate(req); ve != nil {
		return nil, ve
	}

	// Validate date format YYYY-MM-DD
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, &domain.ValidationError{
			Fields: []domain.FieldError{
				{Field: "Date", Message: "must be a valid date in YYYY-MM-DD format"},
			},
		}
	}

	// Convert amount to cents
	amountCents, err := validator.ToCents(req.Amount)
	if err != nil {
		return nil, &domain.ValidationError{
			Fields: []domain.FieldError{
				{Field: "Amount", Message: "must have at most 2 decimal places"},
			},
		}
	}

	// Parse and verify account ownership
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, &domain.ValidationError{
			Fields: []domain.FieldError{
				{Field: "AccountID", Message: "must be a valid UUID"},
			},
		}
	}

	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrNotFound
	}
	if account.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// If category_id is provided, verify ownership
	var categoryID *uuid.UUID
	if req.CategoryID != nil {
		catID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, &domain.ValidationError{
				Fields: []domain.FieldError{
					{Field: "CategoryID", Message: "must be a valid UUID"},
				},
			}
		}
		category, err := s.categoryRepo.GetByID(ctx, catID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, domain.ErrNotFound
		}
		if category.UserID != userID {
			return nil, domain.ErrForbidden
		}
		categoryID = &catID
	}

	// Create the transaction
	now := time.Now()
	transaction := &domain.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		AccountID:   accountID,
		CategoryID:  categoryID,
		Type:        domain.TransactionType(req.Type),
		Amount:      amountCents,
		Description: req.Description,
		Date:        date,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.txRepo.Create(ctx, transaction); err != nil {
		return nil, err
	}

	// Update account balance atomically
	delta := balanceDelta(transaction.Type, amountCents)
	if err := s.accountRepo.UpdateBalance(ctx, accountID, delta); err != nil {
		return nil, err
	}

	return transaction, nil
}

// List returns transactions for the given user with filters and pagination.
func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	filter.UserID = userID

	// Enforce default and max limits
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return s.txRepo.List(ctx, filter)
}

// Update validates the input, reverts old balance effect, applies new balance effect,
// and updates the transaction.
func (s *TransactionService) Update(ctx context.Context, userID uuid.UUID, txID uuid.UUID, req UpdateTransactionRequest) (*domain.Transaction, error) {
	if ve := validator.Validate(req); ve != nil {
		return nil, ve
	}

	// Validate date format
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, &domain.ValidationError{
			Fields: []domain.FieldError{
				{Field: "Date", Message: "must be a valid date in YYYY-MM-DD format"},
			},
		}
	}

	// Convert amount to cents
	amountCents, err := validator.ToCents(req.Amount)
	if err != nil {
		return nil, &domain.ValidationError{
			Fields: []domain.FieldError{
				{Field: "Amount", Message: "must have at most 2 decimal places"},
			},
		}
	}

	// Get existing transaction
	existing, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}

	// Check ownership
	if existing.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// Parse and verify new account ownership
	newAccountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, &domain.ValidationError{
			Fields: []domain.FieldError{
				{Field: "AccountID", Message: "must be a valid UUID"},
			},
		}
	}

	newAccount, err := s.accountRepo.GetByID(ctx, newAccountID)
	if err != nil {
		return nil, err
	}
	if newAccount == nil {
		return nil, domain.ErrNotFound
	}
	if newAccount.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// If new category_id is provided, verify ownership
	var newCategoryID *uuid.UUID
	if req.CategoryID != nil {
		catID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, &domain.ValidationError{
				Fields: []domain.FieldError{
					{Field: "CategoryID", Message: "must be a valid UUID"},
				},
			}
		}
		category, err := s.categoryRepo.GetByID(ctx, catID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, domain.ErrNotFound
		}
		if category.UserID != userID {
			return nil, domain.ErrForbidden
		}
		newCategoryID = &catID
	}

	// Revert old balance effect on old account
	revertDelta := balanceDelta(existing.Type, existing.Amount) * -1
	if err := s.accountRepo.UpdateBalance(ctx, existing.AccountID, revertDelta); err != nil {
		return nil, err
	}

	// Apply new balance effect on new account
	applyDelta := balanceDelta(domain.TransactionType(req.Type), amountCents)
	if err := s.accountRepo.UpdateBalance(ctx, newAccountID, applyDelta); err != nil {
		return nil, err
	}

	// Update the transaction fields
	existing.AccountID = newAccountID
	existing.CategoryID = newCategoryID
	existing.Type = domain.TransactionType(req.Type)
	existing.Amount = amountCents
	existing.Description = req.Description
	existing.Date = date
	existing.UpdatedAt = time.Now()

	if err := s.txRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete soft-deletes the transaction and reverts its balance effect on the account.
func (s *TransactionService) Delete(ctx context.Context, userID uuid.UUID, txID uuid.UUID) error {
	// Get transaction
	transaction, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return err
	}
	if transaction == nil {
		return domain.ErrNotFound
	}

	// Check ownership
	if transaction.UserID != userID {
		return domain.ErrForbidden
	}

	// Revert balance effect
	revertDelta := balanceDelta(transaction.Type, transaction.Amount) * -1
	if err := s.accountRepo.UpdateBalance(ctx, transaction.AccountID, revertDelta); err != nil {
		return err
	}

	// Soft-delete
	return s.txRepo.SoftDelete(ctx, txID)
}

// balanceDelta calculates the delta to apply to an account balance.
// income → positive delta (adds to balance)
// expense → negative delta (subtracts from balance)
func balanceDelta(txType domain.TransactionType, amount int64) int64 {
	if txType == domain.TransactionTypeIncome {
		return amount
	}
	return -amount
}
