package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/repository/sqlc"
)

// Compile-time assertion that CategoryRepository implements domain.CategoryRepository.
var _ domain.CategoryRepository = (*CategoryRepository)(nil)

// CategoryRepository implements domain.CategoryRepository using sqlc-generated queries.
type CategoryRepository struct {
	q *sqlc.Queries
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db sqlc.DBTX) *CategoryRepository {
	return &CategoryRepository{
		q: sqlc.New(db),
	}
}

// Create inserts a new category into the database.
func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	row, err := r.q.CreateCategory(ctx, category.UserID, category.Name, string(category.Type))
	if err != nil {
		return mapPgError(err)
	}

	category.ID = row.ID
	category.CreatedAt = row.CreatedAt
	category.UpdatedAt = row.UpdatedAt
	return nil
}

// GetByID retrieves a category by its ID (excluding soft-deleted).
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	row, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, mapPgError(err)
	}
	return toDomainCategory(row), nil
}

// ListByUserID returns all non-deleted categories for a given user.
func (r *CategoryRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	rows, err := r.q.ListCategoriesByUserID(ctx, userID)
	if err != nil {
		return nil, mapPgError(err)
	}

	categories := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, *toDomainCategory(row))
	}
	return categories, nil
}

// Update updates an existing category's name.
func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	row, err := r.q.UpdateCategory(ctx, category.ID, category.Name)
	if err != nil {
		return mapPgError(err)
	}

	category.UpdatedAt = row.UpdatedAt
	return nil
}

// SoftDelete marks a category as deleted by setting deleted_at.
func (r *CategoryRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.q.SoftDeleteCategory(ctx, id)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// CreateDefaultsForUser bulk-inserts the default set of categories for a new user.
func (r *CategoryRepository) CreateDefaultsForUser(ctx context.Context, userID uuid.UUID) error {
	err := r.q.CreateDefaultCategories(ctx, userID)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// toDomainCategory converts a sqlc Category model to a domain Category entity.
func toDomainCategory(row sqlc.Category) *domain.Category {
	cat := &domain.Category{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Type:      domain.CategoryType(row.Type),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		cat.DeletedAt = &t
	}

	return cat
}
