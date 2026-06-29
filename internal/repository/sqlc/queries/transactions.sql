-- name: CreateTransaction :one
INSERT INTO transactions (user_id, account_id, category_id, type, amount, description, date)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, account_id, category_id, type, amount, description, date, created_at, updated_at, deleted_at;

-- name: GetTransactionByID :one
SELECT id, user_id, account_id, category_id, type, amount, description, date, created_at, updated_at, deleted_at
FROM transactions
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListTransactions :many
SELECT id, user_id, account_id, category_id, type, amount, description, date, created_at, updated_at, deleted_at
FROM transactions
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (sqlc.narg('account_id')::UUID IS NULL OR account_id = sqlc.narg('account_id'))
  AND (sqlc.narg('category_id')::UUID IS NULL OR category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('type')::VARCHAR IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('date_from')::DATE IS NULL OR date >= sqlc.narg('date_from'))
  AND (sqlc.narg('date_to')::DATE IS NULL OR date <= sqlc.narg('date_to'))
ORDER BY date DESC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateTransaction :one
UPDATE transactions
SET account_id = $2,
    category_id = $3,
    type = $4,
    amount = $5,
    description = $6,
    date = $7,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, user_id, account_id, category_id, type, amount, description, date, created_at, updated_at, deleted_at;

-- name: SoftDeleteTransaction :exec
UPDATE transactions
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
