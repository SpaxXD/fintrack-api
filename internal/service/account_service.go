package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/validator"
)

// CreateAccountRequest holds the input for creating a new account.
type CreateAccountRequest struct {
	Name           string `json:"name" validate:"required,min=1,max=100"`
	Type           string `json:"type" validate:"required,oneof=checking savings credit_card cash investment"`
	InitialBalance *int64 `json:"initial_balance"`
}

// UpdateAccountRequest holds the input for updating an existing account.
type UpdateAccountRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
	Type string `json:"type" validate:"required,oneof=checking savings credit_card cash investment"`
}

// AccountService defines the interface for account business logic.
type AccountService interface {
	Create(ctx context.Context, userID uuid.UUID, req CreateAccountRequest) (*domain.Account, error)
	List(ctx context.Context, userID uuid.UUID) ([]domain.Account, error)
	Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, req UpdateAccountRequest) (*domain.Account, error)
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}

// accountService implements AccountService.
type accountService struct {
	repo domain.AccountRepository
}

// NewAccountService creates a new AccountService with the given repository.
func NewAccountService(repo domain.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

// Create validates the input, builds an Account entity, and persists it.
func (s *accountService) Create(ctx context.Context, userID uuid.UUID, req CreateAccountRequest) (*domain.Account, error) {
	if ve := validator.Validate(req); ve != nil {
		return nil, ve
	}

	var balance int64
	if req.InitialBalance != nil {
		balance = *req.InitialBalance
	}

	account := &domain.Account{
		ID:      uuid.New(),
		UserID:  userID,
		Name:    req.Name,
		Type:    domain.AccountType(req.Type),
		Balance: balance,
	}

	if err := s.repo.Create(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// List returns all non-deleted accounts for the given user.
func (s *accountService) List(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// Update validates the input, checks ownership, and updates the account.
func (s *accountService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, req UpdateAccountRequest) (*domain.Account, error) {
	if ve := validator.Validate(req); ve != nil {
		return nil, ve
	}

	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrNotFound
	}

	if account.UserID != userID {
		return nil, domain.ErrForbidden
	}

	account.Name = req.Name
	account.Type = domain.AccountType(req.Type)

	if err := s.repo.Update(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// Delete checks ownership and soft-deletes the account.
func (s *accountService) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if account == nil {
		return domain.ErrNotFound
	}

	if account.UserID != userID {
		return domain.ErrForbidden
	}

	return s.repo.SoftDelete(ctx, id)
}
