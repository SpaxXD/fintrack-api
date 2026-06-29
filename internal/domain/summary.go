package domain

import "github.com/google/uuid"

// Summary holds aggregated financial data for a given period.
type Summary struct {
	TotalIncome  int64 `json:"total_income"`
	TotalExpense int64 `json:"total_expense"`
	NetBalance   int64 `json:"net_balance"`
}

// CategorySummary holds aggregated transaction data for a specific category.
type CategorySummary struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Type         string    `json:"type"`
	Total        int64     `json:"total"`
}

// MonthlyTrend holds aggregated income and expense data for a specific month.
type MonthlyTrend struct {
	Month        int   `json:"month"`
	TotalIncome  int64 `json:"total_income"`
	TotalExpense int64 `json:"total_expense"`
}
