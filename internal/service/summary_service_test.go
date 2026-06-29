package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/domain"
)

// --- GetSummary Tests ---

func TestSummaryService_GetSummary_Empty(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	summary, err := svc.GetSummary(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary.TotalIncome != 0 {
		t.Errorf("expected total_income 0, got %d", summary.TotalIncome)
	}
	if summary.TotalExpense != 0 {
		t.Errorf("expected total_expense 0, got %d", summary.TotalExpense)
	}
	if summary.NetBalance != 0 {
		t.Errorf("expected net_balance 0, got %d", summary.NetBalance)
	}
}

func TestSummaryService_GetSummary_WithTransactions(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	accountID := uuid.New()

	txRepo.transactions = []*domain.Transaction{
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 5000, Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 3000, Date: time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeExpense, Amount: 2000, Date: time.Date(2024, 4, 20, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeExpense, Amount: 1500, Date: time.Date(2024, 7, 5, 0, 0, 0, 0, time.UTC)},
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	summary, err := svc.GetSummary(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary.TotalIncome != 8000 {
		t.Errorf("expected total_income 8000, got %d", summary.TotalIncome)
	}
	if summary.TotalExpense != 3500 {
		t.Errorf("expected total_expense 3500, got %d", summary.TotalExpense)
	}
	if summary.NetBalance != 4500 {
		t.Errorf("expected net_balance 4500, got %d", summary.NetBalance)
	}
}

func TestSummaryService_GetSummary_ExcludesOtherUsers(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	otherUserID := uuid.New()
	accountID := uuid.New()

	txRepo.transactions = []*domain.Transaction{
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 1000, Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: otherUserID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 9999, Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	summary, err := svc.GetSummary(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary.TotalIncome != 1000 {
		t.Errorf("expected total_income 1000, got %d", summary.TotalIncome)
	}
}

func TestSummaryService_GetSummary_NetBalanceNegative(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	accountID := uuid.New()

	txRepo.transactions = []*domain.Transaction{
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 1000, Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeExpense, Amount: 5000, Date: time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)},
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	summary, err := svc.GetSummary(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary.NetBalance != -4000 {
		t.Errorf("expected net_balance -4000, got %d", summary.NetBalance)
	}
}

// --- GetCategorySummary Tests ---

func TestSummaryService_GetCategorySummary_Empty(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	result, err := svc.GetCategorySummary(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}

func TestSummaryService_GetCategorySummary_GroupsByCategory(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	accountID := uuid.New()
	foodCatID := uuid.New()
	salaryCatID := uuid.New()

	catRepo.categories = []*domain.Category{
		{ID: foodCatID, UserID: userID, Name: "Food", Type: domain.CategoryTypeExpense},
		{ID: salaryCatID, UserID: userID, Name: "Salary", Type: domain.CategoryTypeIncome},
	}

	txRepo.transactions = []*domain.Transaction{
		{ID: uuid.New(), UserID: userID, AccountID: accountID, CategoryID: &foodCatID, Type: domain.TransactionTypeExpense, Amount: 1000, Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, CategoryID: &foodCatID, Type: domain.TransactionTypeExpense, Amount: 2000, Date: time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, CategoryID: &salaryCatID, Type: domain.TransactionTypeIncome, Amount: 5000, Date: time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)},
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	result, err := svc.GetCategorySummary(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 category summaries, got %d", len(result))
	}

	// Check by category
	catTotals := make(map[uuid.UUID]int64)
	catNames := make(map[uuid.UUID]string)
	for _, cs := range result {
		catTotals[cs.CategoryID] = cs.Total
		catNames[cs.CategoryID] = cs.CategoryName
	}
	if catTotals[foodCatID] != 3000 {
		t.Errorf("expected Food total 3000, got %d", catTotals[foodCatID])
	}
	if catTotals[salaryCatID] != 5000 {
		t.Errorf("expected Salary total 5000, got %d", catTotals[salaryCatID])
	}
	if catNames[foodCatID] != "Food" {
		t.Errorf("expected category name 'Food', got %q", catNames[foodCatID])
	}
	if catNames[salaryCatID] != "Salary" {
		t.Errorf("expected category name 'Salary', got %q", catNames[salaryCatID])
	}
}

func TestSummaryService_GetCategorySummary_SkipsNilCategory(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	accountID := uuid.New()
	catID := uuid.New()

	catRepo.categories = []*domain.Category{
		{ID: catID, UserID: userID, Name: "Food", Type: domain.CategoryTypeExpense},
	}

	txRepo.transactions = []*domain.Transaction{
		{ID: uuid.New(), UserID: userID, AccountID: accountID, CategoryID: nil, Type: domain.TransactionTypeExpense, Amount: 500, Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, CategoryID: &catID, Type: domain.TransactionTypeExpense, Amount: 1000, Date: time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)},
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	result, err := svc.GetCategorySummary(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 category summary (nil category skipped), got %d", len(result))
	}
	if result[0].Total != 1000 {
		t.Errorf("expected total 1000, got %d", result[0].Total)
	}
}

// --- GetMonthlyTrend Tests ---

func TestSummaryService_GetMonthlyTrend_Empty(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	trends, err := svc.GetMonthlyTrend(context.Background(), userID, 2024)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(trends) != 12 {
		t.Fatalf("expected 12 months, got %d", len(trends))
	}
	for i, tr := range trends {
		if tr.Month != i+1 {
			t.Errorf("month[%d]: expected month %d, got %d", i, i+1, tr.Month)
		}
		if tr.TotalIncome != 0 {
			t.Errorf("month[%d]: expected total_income 0, got %d", i, tr.TotalIncome)
		}
		if tr.TotalExpense != 0 {
			t.Errorf("month[%d]: expected total_expense 0, got %d", i, tr.TotalExpense)
		}
	}
}

func TestSummaryService_GetMonthlyTrend_WithTransactions(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	accountID := uuid.New()

	txRepo.transactions = []*domain.Transaction{
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 5000, Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeExpense, Amount: 2000, Date: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 3000, Date: time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeExpense, Amount: 1000, Date: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)},
	}

	trends, err := svc.GetMonthlyTrend(context.Background(), userID, 2024)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(trends) != 12 {
		t.Fatalf("expected 12 months, got %d", len(trends))
	}

	// January: income 5000, expense 2000
	if trends[0].TotalIncome != 5000 {
		t.Errorf("Jan: expected income 5000, got %d", trends[0].TotalIncome)
	}
	if trends[0].TotalExpense != 2000 {
		t.Errorf("Jan: expected expense 2000, got %d", trends[0].TotalExpense)
	}

	// June: income 3000, expense 0
	if trends[5].TotalIncome != 3000 {
		t.Errorf("Jun: expected income 3000, got %d", trends[5].TotalIncome)
	}
	if trends[5].TotalExpense != 0 {
		t.Errorf("Jun: expected expense 0, got %d", trends[5].TotalExpense)
	}

	// December: income 0, expense 1000
	if trends[11].TotalIncome != 0 {
		t.Errorf("Dec: expected income 0, got %d", trends[11].TotalIncome)
	}
	if trends[11].TotalExpense != 1000 {
		t.Errorf("Dec: expected expense 1000, got %d", trends[11].TotalExpense)
	}

	// Months with no transactions should be zero
	if trends[1].TotalIncome != 0 || trends[1].TotalExpense != 0 {
		t.Errorf("Feb: expected zeros, got income=%d expense=%d", trends[1].TotalIncome, trends[1].TotalExpense)
	}
}

func TestSummaryService_GetMonthlyTrend_ExcludesOtherYears(t *testing.T) {
	txRepo := newMockTransactionRepository()
	catRepo := newMockCategoryRepository()
	svc := NewSummaryService(txRepo, catRepo)

	userID := uuid.New()
	accountID := uuid.New()

	txRepo.transactions = []*domain.Transaction{
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 5000, Date: time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 3000, Date: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), UserID: userID, AccountID: accountID, Type: domain.TransactionTypeIncome, Amount: 9000, Date: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)},
	}

	trends, err := svc.GetMonthlyTrend(context.Background(), userID, 2024)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Only the 2024 transaction should appear
	totalIncome := int64(0)
	for _, tr := range trends {
		totalIncome += tr.TotalIncome
	}
	if totalIncome != 3000 {
		t.Errorf("expected total income 3000 (only 2024), got %d", totalIncome)
	}
}
