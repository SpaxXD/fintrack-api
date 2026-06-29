package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/muriloabranches/fintrack-api/internal/domain"
)

// mockCategoryRepo is an in-memory mock of domain.CategoryRepository for testing.
type mockCategoryRepo struct {
	categories map[uuid.UUID]*domain.Category
	createErr  error
	getErr     error
	updateErr  error
	deleteErr  error
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{
		categories: make(map[uuid.UUID]*domain.Category),
	}
}

func (m *mockCategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	if m.createErr != nil {
		return m.createErr
	}
	category.ID = uuid.New()
	m.categories[category.ID] = category
	return nil
}

func (m *mockCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	cat, ok := m.categories[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cat, nil
}

func (m *mockCategoryRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	var result []domain.Category
	for _, c := range m.categories {
		if c.UserID == userID && c.DeletedAt == nil {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCategoryRepo) Update(ctx context.Context, category *domain.Category) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.categories[category.ID] = category
	return nil
}

func (m *mockCategoryRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.categories, id)
	return nil
}

func (m *mockCategoryRepo) CreateDefaultsForUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// --- Tests ---

func TestCategoryService_Create_Success(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	req := CreateCategoryRequest{
		Name: "Alimentação",
		Type: "expense",
	}

	cat, err := svc.Create(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cat.Name != "Alimentação" {
		t.Errorf("expected name 'Alimentação', got '%s'", cat.Name)
	}
	if cat.Type != domain.CategoryTypeExpense {
		t.Errorf("expected type 'expense', got '%s'", cat.Type)
	}
	if cat.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, cat.UserID)
	}
	if cat.ID == uuid.Nil {
		t.Error("expected a non-nil ID")
	}
}

func TestCategoryService_Create_ValidationError_EmptyName(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	req := CreateCategoryRequest{
		Name: "",
		Type: "income",
	}

	_, err := svc.Create(context.Background(), userID, req)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	ve, ok := err.(*domain.ValidationError)
	if !ok {
		t.Fatalf("expected *domain.ValidationError, got %T", err)
	}
	if len(ve.Fields) == 0 {
		t.Error("expected at least one field error")
	}
}

func TestCategoryService_Create_ValidationError_InvalidType(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	req := CreateCategoryRequest{
		Name: "Salário",
		Type: "invalid_type",
	}

	_, err := svc.Create(context.Background(), userID, req)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	_, ok := err.(*domain.ValidationError)
	if !ok {
		t.Fatalf("expected *domain.ValidationError, got %T", err)
	}
}

func TestCategoryService_Create_ValidationError_NameTooLong(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	longName := ""
	for i := 0; i < 51; i++ {
		longName += "a"
	}

	req := CreateCategoryRequest{
		Name: longName,
		Type: "income",
	}

	_, err := svc.Create(context.Background(), userID, req)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	_, ok := err.(*domain.ValidationError)
	if !ok {
		t.Fatalf("expected *domain.ValidationError, got %T", err)
	}
}

func TestCategoryService_List(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	// Create two categories
	svc.Create(context.Background(), userID, CreateCategoryRequest{Name: "Food", Type: "expense"})
	svc.Create(context.Background(), userID, CreateCategoryRequest{Name: "Salary", Type: "income"})

	// Another user's category
	otherUserID := uuid.New()
	svc.Create(context.Background(), otherUserID, CreateCategoryRequest{Name: "Transport", Type: "expense"})

	categories, err := svc.List(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(categories))
	}
}

func TestCategoryService_Update_Success(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	cat, _ := svc.Create(context.Background(), userID, CreateCategoryRequest{Name: "Food", Type: "expense"})

	updated, err := svc.Update(context.Background(), userID, cat.ID, UpdateCategoryRequest{Name: "Alimentação"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "Alimentação" {
		t.Errorf("expected name 'Alimentação', got '%s'", updated.Name)
	}
}

func TestCategoryService_Update_NotFound(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	_, err := svc.Update(context.Background(), userID, uuid.New(), UpdateCategoryRequest{Name: "New Name"})
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCategoryService_Update_Forbidden(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	ownerID := uuid.New()
	otherUserID := uuid.New()

	cat, _ := svc.Create(context.Background(), ownerID, CreateCategoryRequest{Name: "Food", Type: "expense"})

	_, err := svc.Update(context.Background(), otherUserID, cat.ID, UpdateCategoryRequest{Name: "Hacked"})
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCategoryService_Update_ValidationError(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	cat, _ := svc.Create(context.Background(), userID, CreateCategoryRequest{Name: "Food", Type: "expense"})

	_, err := svc.Update(context.Background(), userID, cat.ID, UpdateCategoryRequest{Name: ""})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	_, ok := err.(*domain.ValidationError)
	if !ok {
		t.Fatalf("expected *domain.ValidationError, got %T", err)
	}
}

func TestCategoryService_Delete_Success(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	cat, _ := svc.Create(context.Background(), userID, CreateCategoryRequest{Name: "Food", Type: "expense"})

	err := svc.Delete(context.Background(), userID, cat.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's gone from the repo
	_, err = repo.GetByID(context.Background(), cat.ID)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestCategoryService_Delete_NotFound(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	userID := uuid.New()

	err := svc.Delete(context.Background(), userID, uuid.New())
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCategoryService_Delete_Forbidden(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	ownerID := uuid.New()
	otherUserID := uuid.New()

	cat, _ := svc.Create(context.Background(), ownerID, CreateCategoryRequest{Name: "Food", Type: "expense"})

	err := svc.Delete(context.Background(), otherUserID, cat.ID)
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
