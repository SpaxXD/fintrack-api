package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/muriloabranches/fintrack-api/internal/domain"
	"github.com/muriloabranches/fintrack-api/internal/repository/sqlc"
)

// Compile-time assertion that TransactionRepository implements domain.TransactionRepository.
var _ domain.TransactionRepository = (*TransactionRepository)(nil)

// TransactionRepository implements domain.TransactionRepository using sqlc-generated queries.
type TransactionRepository struct {
	q *sqlc.Queries
}

// NewTransactionRepository creates a new TransactionRepository.
func NewTransactionRepository(db sqlc.DBTX) *TransactionRepository {
	return &TransactionRepository{
		q: sqlc.New(db),
	}
}

// Create inserts a new transaction into the database.
func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	categoryID := uuidToPgtype(tx.CategoryID)
	date := timeToPgDate(tx.Date)

	row, err := r.q.CreateTransaction(ctx, tx.UserID, tx.AccountID, categoryID, string(tx.Type), tx.Amount, tx.Description, date)
	if err != nil {
		return mapPgError(err)
	}

	tx.ID = row.ID
	tx.CreatedAt = row.CreatedAt
	tx.UpdatedAt = row.UpdatedAt
	return nil
}

// GetByID retrieves a transaction by its ID.
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	row, err := r.q.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, mapPgError(err)
	}
	return toDomainTransaction(row), nil
}

// List retrieves transactions with dynamic filtering and pagination.
func (r *TransactionRepository) List(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	params := sqlc.ListTransactionsParams{
		UserID:     filter.UserID,
		AccountID:  uuidToPgtype(filter.AccountID),
		CategoryID: uuidToPgtype(filter.CategoryID),
		Type:       transactionTypeToPgText(filter.Type),
		DateFrom:   timePtrToPgDate(filter.DateFrom),
		DateTo:     timePtrToPgDate(filter.DateTo),
		Limit:      int32(limit),
		Offset:     int32(offset),
	}

	rows, err := r.q.ListTransactions(ctx, params)
	if err != nil {
		return nil, mapPgError(err)
	}

	transactions := make([]domain.Transaction, 0, len(rows))
	for _, row := range rows {
		transactions = append(transactions, *toDomainTransaction(row))
	}
	return transactions, nil
}

// Update updates an existing transaction.
func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	categoryID := uuidToPgtype(tx.CategoryID)
	date := timeToPgDate(tx.Date)

	row, err := r.q.UpdateTransaction(ctx, tx.ID, tx.AccountID, categoryID, string(tx.Type), tx.Amount, tx.Description, date)
	if err != nil {
		return mapPgError(err)
	}

	tx.UpdatedAt = row.UpdatedAt
	return nil
}

// SoftDelete marks a transaction as deleted by setting deleted_at.
func (r *TransactionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.q.SoftDeleteTransaction(ctx, id)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

// --- Helper functions ---

// uuidToPgtype converts a *uuid.UUID to pgtype.UUID for nullable UUID columns.
func uuidToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

// pgtypeToUUID converts a pgtype.UUID to *uuid.UUID.
func pgtypeToUUID(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	u := uuid.UUID(id.Bytes)
	return &u
}

// timeToPgDate converts a time.Time to pgtype.Date.
func timeToPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}

// timePtrToPgDate converts a *time.Time to pgtype.Date (null if nil).
func timePtrToPgDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{
		Time:  *t,
		Valid: true,
	}
}

// transactionTypeToPgText converts a *domain.TransactionType to pgtype.Text.
func transactionTypeToPgText(t *domain.TransactionType) pgtype.Text {
	if t == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: string(*t), Valid: true}
}

// pgDateToTime converts a pgtype.Date to time.Time.
func pgDateToTime(d pgtype.Date) time.Time {
	if !d.Valid {
		return time.Time{}
	}
	return d.Time
}

// toDomainTransaction converts a sqlc Transaction model to a domain Transaction entity.
func toDomainTransaction(row sqlc.Transaction) *domain.Transaction {
	tx := &domain.Transaction{
		ID:          row.ID,
		UserID:      row.UserID,
		AccountID:   row.AccountID,
		CategoryID:  pgtypeToUUID(row.CategoryID),
		Type:        domain.TransactionType(row.Type),
		Amount:      row.Amount,
		Description: row.Description,
		Date:        pgDateToTime(row.Date),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}

	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		tx.DeletedAt = &deletedAt
	}

	return tx
}
