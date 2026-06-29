package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/repository/sqlc"
)

// Compile-time assertion that UserRepository implements domain.UserRepository.
var _ domain.UserRepository = (*UserRepository)(nil)

// UserRepository implements domain.UserRepository using sqlc-generated queries.
type UserRepository struct {
	q *sqlc.Queries
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db sqlc.DBTX) *UserRepository {
	return &UserRepository{
		q: sqlc.New(db),
	}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	row, err := r.q.CreateUser(ctx, user.Email, user.PasswordHash)
	if err != nil {
		return mapPgError(err)
	}

	user.ID = row.ID
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, mapPgError(err)
	}
	return toDomainUser(row), nil
}

// GetByID retrieves a user by their ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, mapPgError(err)
	}
	return toDomainUser(row), nil
}

// IncrementFailedAttempts increments the failed login attempts counter for a user.
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, email string) error {
	err := r.q.IncrementFailedAttempts(ctx, email)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// ResetFailedAttempts resets the failed login attempts counter for a user.
func (r *UserRepository) ResetFailedAttempts(ctx context.Context, email string) error {
	err := r.q.ResetFailedAttempts(ctx, email)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// GetFailedAttempts returns the number of failed login attempts and the lock expiration time.
func (r *UserRepository) GetFailedAttempts(ctx context.Context, email string) (int, time.Time, error) {
	row, err := r.q.GetFailedAttempts(ctx, email)
	if err != nil {
		return 0, time.Time{}, mapPgError(err)
	}

	var lockedUntil time.Time
	if row.LockedUntil.Valid {
		lockedUntil = row.LockedUntil.Time
	}

	return int(row.FailedAttempts), lockedUntil, nil
}

// toDomainUser converts a sqlc User model to a domain User entity.
func toDomainUser(row sqlc.User) *domain.User {
	return &domain.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

// mapPgError translates PostgreSQL errors to domain errors.
func mapPgError(err error) error {
	if err == nil {
		return nil
	}

	// pgx.ErrNoRows → domain.ErrNotFound
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	// Unique constraint violation (23505) → domain.ErrConflict
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrConflict
		}
	}

	return fmt.Errorf("database error: %w", err)
}
