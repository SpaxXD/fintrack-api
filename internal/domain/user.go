package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents a registered user in the system.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRepository defines the contract for user persistence operations.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	IncrementFailedAttempts(ctx context.Context, email string) error
	ResetFailedAttempts(ctx context.Context, email string) error
	GetFailedAttempts(ctx context.Context, email string) (int, time.Time, error)
}
