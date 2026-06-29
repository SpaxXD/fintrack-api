package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/repository/sqlc"
)

// Compile-time assertion that TokenRepository implements domain.TokenRepository.
var _ domain.TokenRepository = (*TokenRepository)(nil)

// TokenRepository implements domain.TokenRepository using sqlc-generated queries.
type TokenRepository struct {
	q *sqlc.Queries
}

// NewTokenRepository creates a new TokenRepository.
func NewTokenRepository(db sqlc.DBTX) *TokenRepository {
	return &TokenRepository{
		q: sqlc.New(db),
	}
}

// Create inserts a new refresh token into the database.
func (r *TokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	row, err := r.q.CreateRefreshToken(ctx, token.UserID, token.Token, token.ExpiresAt)
	if err != nil {
		return mapPgError(err)
	}

	token.ID = row.ID
	token.Revoked = row.Revoked
	token.CreatedAt = row.CreatedAt
	return nil
}

// GetByToken retrieves a non-revoked refresh token by its token string.
func (r *TokenRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByToken(ctx, tokenStr)
	if err != nil {
		return nil, mapPgError(err)
	}
	return toDomainRefreshToken(row), nil
}

// Revoke marks a specific refresh token as revoked.
func (r *TokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	err := r.q.RevokeRefreshToken(ctx, id)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// RevokeAllByUserID revokes all active refresh tokens for a given user.
func (r *TokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	err := r.q.RevokeAllRefreshTokensByUserID(ctx, userID)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// toDomainRefreshToken converts a sqlc RefreshToken model to a domain RefreshToken entity.
func toDomainRefreshToken(row sqlc.RefreshToken) *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		Token:     row.Token,
		ExpiresAt: row.ExpiresAt,
		Revoked:   row.Revoked,
		CreatedAt: row.CreatedAt,
	}
}
