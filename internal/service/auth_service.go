package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/config"
	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents the input for user registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// LoginRequest represents the input for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// TokenPair represents an access/refresh token pair returned on login or refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthService defines the contract for authentication operations.
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req LoginRequest) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}

// authService implements AuthService.
type authService struct {
	userRepo     domain.UserRepository
	tokenRepo    domain.TokenRepository
	categoryRepo domain.CategoryRepository
	jwtCfg       config.JWTConfig
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(
	userRepo domain.UserRepository,
	tokenRepo domain.TokenRepository,
	categoryRepo domain.CategoryRepository,
	jwtCfg config.JWTConfig,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		categoryRepo: categoryRepo,
		jwtCfg:       jwtCfg,
	}
}

// Register creates a new user with hashed password and default categories.
func (s *authService) Register(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	if ve := validator.Validate(req); ve != nil {
		return nil, ve
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := s.categoryRepo.CreateDefaultsForUser(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("creating default categories: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns a token pair.
func (s *authService) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	if ve := validator.Validate(req); ve != nil {
		return nil, ve
	}

	attempts, lockedUntil, err := s.userRepo.GetFailedAttempts(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("checking failed attempts: %w", err)
	}
	if attempts >= 5 && time.Now().Before(lockedUntil) {
		return nil, domain.ErrRateLimited
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		_ = s.userRepo.IncrementFailedAttempts(ctx, req.Email)
		return nil, domain.ErrUnauthorized
	}

	_ = s.userRepo.ResetFailedAttempts(ctx, req.Email)

	pair, err := s.generateTokenPair(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

// RefreshToken validates a refresh token, revokes it, and issues a new pair (rotation).
func (s *authService) RefreshToken(ctx context.Context, refreshTokenStr string) (*TokenPair, error) {
	token, err := s.tokenRepo.GetByToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, domain.ErrUnauthorized
	}

	if token.Revoked {
		return nil, domain.ErrUnauthorized
	}

	if err := s.tokenRepo.Revoke(ctx, token.ID); err != nil {
		return nil, fmt.Errorf("revoking refresh token: %w", err)
	}

	pair, err := s.generateTokenPair(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

// generateTokenPair creates a new access token (JWT) and refresh token, persisting the refresh token.
func (s *authService) generateTokenPair(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	now := time.Now()

	accessClaims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": jwt.NewNumericDate(now.Add(s.jwtCfg.AccessTokenExpiry)),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshTokenStr := uuid.New().String()
	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     refreshTokenStr,
		ExpiresAt: now.Add(s.jwtCfg.RefreshTokenExpiry),
		Revoked:   false,
		CreatedAt: now,
	}

	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
	}, nil
}
