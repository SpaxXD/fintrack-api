package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TransactionType represents the type of a financial transaction.
type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

// Transaction represents a financial transaction.
type Transaction struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	AccountID   uuid.UUID       `json:"account_id"`
	CategoryID  *uuid.UUID      `json:"category_id,omitempty"`
	Type        TransactionType `json:"type"`
	Amount      int64           `json:"amount"`
	Description string          `json:"description"`
	Date        time.Time       `json:"date"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}

// TransactionFilter holds criteria for filtering transactions.
type TransactionFilter struct {
	UserID     uuid.UUID
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	Type       *TransactionType
	DateFrom   *time.Time
	DateTo     *time.Time
	Limit      int
	Offset     int
}

// TransactionRepository defines the contract for transaction persistence operations.
type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	List(ctx context.Context, filter TransactionFilter) ([]Transaction, error)
	Update(ctx context.Context, tx *Transaction) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
