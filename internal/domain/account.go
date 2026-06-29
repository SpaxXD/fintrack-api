package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AccountType represents the type of a financial account.
type AccountType string

const (
	AccountTypeChecking   AccountType = "checking"
	AccountTypeSavings    AccountType = "savings"
	AccountTypeCreditCard AccountType = "credit_card"
	AccountTypeCash       AccountType = "cash"
	AccountTypeInvestment AccountType = "investment"
)

// Account represents a financial account owned by a user.
type Account struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"user_id"`
	Name      string      `json:"name"`
	Type      AccountType `json:"type"`
	Balance   int64       `json:"balance"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	DeletedAt *time.Time  `json:"deleted_at,omitempty"`
}

// AccountRepository defines the contract for account persistence operations.
type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Account, error)
	Update(ctx context.Context, account *Account) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	UpdateBalance(ctx context.Context, id uuid.UUID, delta int64) error
}
