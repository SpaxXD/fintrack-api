package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/repository/sqlc"
)

// Compile-time assertion that AccountRepository implements domain.AccountRepository.
var _ domain.AccountRepository = (*AccountRepository)(nil)

// AccountRepository implements domain.AccountRepository using sqlc-generated queries.
type AccountRepository struct {
	q *sqlc.Queries
}

// NewAccountRepository creates a new AccountRepository.
func NewAccountRepository(db sqlc.DBTX) *AccountRepository {
	return &AccountRepository{
		q: sqlc.New(db),
	}
}

// Create inserts a new account into the database.
func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	row, err := r.q.CreateAccount(ctx, account.UserID, account.Name, string(account.Type), account.Balance)
	if err != nil {
		return mapPgError(err)
	}

	account.ID = row.ID
	account.CreatedAt = row.CreatedAt
	account.UpdatedAt = row.UpdatedAt
	return nil
}

// GetByID retrieves an account by its ID (excludes soft-deleted).
func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	row, err := r.q.GetAccountByID(ctx, id)
	if err != nil {
		return nil, mapPgError(err)
	}
	return toDomainAccount(row), nil
}

// ListByUserID retrieves all non-deleted accounts for a user.
func (r *AccountRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	rows, err := r.q.ListAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, mapPgError(err)
	}

	accounts := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, *toDomainAccount(row))
	}
	return accounts, nil
}

// Update updates an account's name and type.
func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) error {
	row, err := r.q.UpdateAccount(ctx, account.ID, account.Name, string(account.Type))
	if err != nil {
		return mapPgError(err)
	}

	account.UpdatedAt = row.UpdatedAt
	return nil
}

// SoftDelete marks an account as deleted by setting deleted_at.
func (r *AccountRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.q.SoftDeleteAccount(ctx, id)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// UpdateBalance atomically adds delta to the account's balance.
func (r *AccountRepository) UpdateBalance(ctx context.Context, id uuid.UUID, delta int64) error {
	err := r.q.UpdateAccountBalance(ctx, id, delta)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// toDomainAccount converts a sqlc Account model to a domain Account entity.
func toDomainAccount(row sqlc.Account) *domain.Account {
	account := &domain.Account{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Type:      domain.AccountType(row.Type),
		Balance:   row.Balance,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		account.DeletedAt = &t
	}

	return account
}
