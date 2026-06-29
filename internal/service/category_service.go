package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/validator"
)

// CreateCategoryRequest holds input data for creating a category.
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
	Type string `json:"type" validate:"required,oneof=income expense"`
}

// UpdateCategoryRequest holds input data for updating a category.
type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

// CategoryService handles business logic for category management.
type CategoryService struct {
	repo domain.CategoryRepository
}

// NewCategoryService creates a new CategoryService with the given repository.
func NewCategoryService(repo domain.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// Create validates the input and creates a new category for the given user.
func (s *CategoryService) Create(ctx context.Context, userID uuid.UUID, req CreateCategoryRequest) (*domain.Category, error) {
	if ve := validator.Validate(req); ve != nil {
		return nil, ve
	}

	category := &domain.Category{
		UserID:    userID,
		Name:      req.Name,
		Type:      domain.CategoryType(req.Type),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// List returns all non-deleted categories for the given user.
func (s *CategoryService) List(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// Update validates the input, checks ownership, and updates the category name.
func (s *CategoryService) Update(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, req UpdateCategoryRequest) (*domain.Category, error) {
	if ve := validator.Validate(req); ve != nil {
		return nil, ve
	}

	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	if category.UserID != userID {
		return nil, domain.ErrForbidden
	}

	category.Name = req.Name
	category.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// Delete checks ownership and soft-deletes the category.
func (s *CategoryService) Delete(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID) error {
	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}

	if category.UserID != userID {
		return domain.ErrForbidden
	}

	return s.repo.SoftDelete(ctx, categoryID)
}
