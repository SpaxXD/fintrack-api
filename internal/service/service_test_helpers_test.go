package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/domain"
)

// mockTransactionRepository is an in-memory mock for domain.TransactionRepository.
type mockTransactionRepository struct {
	transactions []*domain.Transaction
}

func newMockTransactionRepository() *mockTransactionRepository {
	return &mockTransactionRepository{}
}

func (m *mockTransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	m.transactions = append(m.transactions, tx)
	return nil
}

func (m *mockTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	for _, tx := range m.transactions {
		if tx.ID == id && tx.DeletedAt == nil {
			return tx, nil
		}
	}
	return nil, nil
}

func (m *mockTransactionRepository) List(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	var result []domain.Transaction
	for _, tx := range m.transactions {
		if tx.UserID != filter.UserID {
			continue
		}
		if tx.DeletedAt != nil {
			continue
		}
		if filter.DateFrom != nil && tx.Date.Before(*filter.DateFrom) {
			continue
		}
		if filter.DateTo != nil && tx.Date.After(*filter.DateTo) {
			continue
		}
		if filter.AccountID != nil && tx.AccountID != *filter.AccountID {
			continue
		}
		if filter.CategoryID != nil && (tx.CategoryID == nil || *tx.CategoryID != *filter.CategoryID) {
			continue
		}
		if filter.Type != nil && tx.Type != *filter.Type {
			continue
		}
		result = append(result, *tx)
	}
	// Apply pagination
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(result) {
		return []domain.Transaction{}, nil
	}
	result = result[offset:]
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > len(result) {
		limit = len(result)
	}
	return result[:limit], nil
}

func (m *mockTransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	for i, existing := range m.transactions {
		if existing.ID == tx.ID {
			m.transactions[i] = tx
			return nil
		}
	}
	return domain.ErrNotFound
}

func (m *mockTransactionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	for _, tx := range m.transactions {
		if tx.ID == id {
			now := time.Now()
			tx.DeletedAt = &now
			return nil
		}
	}
	return domain.ErrNotFound
}

// mockCategoryRepository is an in-memory mock for domain.CategoryRepository.
type mockCategoryRepository struct {
	categories []*domain.Category
}

func newMockCategoryRepository() *mockCategoryRepository {
	return &mockCategoryRepository{}
}

func (m *mockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	m.categories = append(m.categories, category)
	return nil
}

func (m *mockCategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	for _, c := range m.categories {
		if c.ID == id && c.DeletedAt == nil {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockCategoryRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	var result []domain.Category
	for _, c := range m.categories {
		if c.UserID == userID && c.DeletedAt == nil {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	for i, c := range m.categories {
		if c.ID == category.ID {
			m.categories[i] = category
			return nil
		}
	}
	return domain.ErrNotFound
}

func (m *mockCategoryRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	for _, c := range m.categories {
		if c.ID == id {
			now := time.Now()
			c.DeletedAt = &now
			return nil
		}
	}
	return domain.ErrNotFound
}

func (m *mockCategoryRepository) CreateDefaultsForUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}
