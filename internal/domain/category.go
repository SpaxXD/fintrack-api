package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CategoryType represents the type of a transaction category.
type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

// Category represents a transaction category owned by a user.
type Category struct {
	ID        uuid.UUID    `json:"id"`
	UserID    uuid.UUID    `json:"user_id"`
	Name      string       `json:"name"`
	Type      CategoryType `json:"type"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at,omitempty"`
}

// CategoryRepository defines the contract for category persistence operations.
type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Category, error)
	Update(ctx context.Context, category *Category) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	CreateDefaultsForUser(ctx context.Context, userID uuid.UUID) error
}
