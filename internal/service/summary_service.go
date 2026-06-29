package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/muriloabranches/fintrack-api/internal/domain"
)

// SummaryService defines the interface for financial summary operations.
type SummaryService interface {
	GetSummary(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) (*domain.Summary, error)
	GetCategorySummary(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]domain.CategorySummary, error)
	GetMonthlyTrend(ctx context.Context, userID uuid.UUID, year int) ([]domain.MonthlyTrend, error)
}

// summaryService implements SummaryService.
type summaryService struct {
	transactionRepo domain.TransactionRepository
	categoryRepo    domain.CategoryRepository
}

// NewSummaryService creates a new SummaryService with the given repositories.
func NewSummaryService(transactionRepo domain.TransactionRepository, categoryRepo domain.CategoryRepository) SummaryService {
	return &summaryService{
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
	}
}

// GetSummary calculates total income, total expense, and net balance for a given period.
func (s *summaryService) GetSummary(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) (*domain.Summary, error) {
	filter := domain.TransactionFilter{
		UserID:   userID,
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Limit:    100000,
		Offset:   0,
	}

	transactions, err := s.transactionRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var totalIncome, totalExpense int64
	for _, tx := range transactions {
		switch tx.Type {
		case domain.TransactionTypeIncome:
			totalIncome += tx.Amount
		case domain.TransactionTypeExpense:
			totalExpense += tx.Amount
		}
	}

	return &domain.Summary{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		NetBalance:   totalIncome - totalExpense,
	}, nil
}

// GetCategorySummary returns totals grouped by category for a given period.
func (s *summaryService) GetCategorySummary(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]domain.CategorySummary, error) {
	filter := domain.TransactionFilter{
		UserID:   userID,
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Limit:    100000,
		Offset:   0,
	}

	transactions, err := s.transactionRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Build a map of category ID → accumulated total and type.
	type catEntry struct {
		total int64
		txType string
	}
	catTotals := make(map[uuid.UUID]*catEntry)

	for _, tx := range transactions {
		if tx.CategoryID == nil {
			continue
		}
		entry, ok := catTotals[*tx.CategoryID]
		if !ok {
			entry = &catEntry{txType: string(tx.Type)}
			catTotals[*tx.CategoryID] = entry
		}
		entry.total += tx.Amount
	}

	if len(catTotals) == 0 {
		return []domain.CategorySummary{}, nil
	}

	// Fetch categories to resolve names.
	categories, err := s.categoryRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	catNameMap := make(map[uuid.UUID]string, len(categories))
	catTypeMap := make(map[uuid.UUID]string, len(categories))
	for _, c := range categories {
		catNameMap[c.ID] = c.Name
		catTypeMap[c.ID] = string(c.Type)
	}

	result := make([]domain.CategorySummary, 0, len(catTotals))
	for catID, entry := range catTotals {
		name := catNameMap[catID]
		if name == "" {
			name = "uncategorized"
		}
		catType := catTypeMap[catID]
		if catType == "" {
			catType = entry.txType
		}
		result = append(result, domain.CategorySummary{
			CategoryID:   catID,
			CategoryName: name,
			Type:         catType,
			Total:        entry.total,
		})
	}

	return result, nil
}

// GetMonthlyTrend returns income and expense totals for each of the 12 months of the given year.
func (s *summaryService) GetMonthlyTrend(ctx context.Context, userID uuid.UUID, year int) ([]domain.MonthlyTrend, error) {
	dateFrom := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(year, time.December, 31, 23, 59, 59, 999999999, time.UTC)

	filter := domain.TransactionFilter{
		UserID:   userID,
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Limit:    100000,
		Offset:   0,
	}

	transactions, err := s.transactionRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Initialize all 12 months with zeros.
	trends := make([]domain.MonthlyTrend, 12)
	for i := range trends {
		trends[i].Month = i + 1
	}

	// Accumulate totals by month.
	for _, tx := range transactions {
		monthIdx := int(tx.Date.Month()) - 1
		if monthIdx < 0 || monthIdx > 11 {
			continue
		}
		switch tx.Type {
		case domain.TransactionTypeIncome:
			trends[monthIdx].TotalIncome += tx.Amount
		case domain.TransactionTypeExpense:
			trends[monthIdx].TotalExpense += tx.Amount
		}
	}

	return trends, nil
}
