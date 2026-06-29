-- name: CreateAccount :one
INSERT INTO accounts (user_id, name, type, balance)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, name, type, balance, created_at, updated_at, deleted_at;

-- name: GetAccountByID :one
SELECT id, user_id, name, type, balance, created_at, updated_at, deleted_at
FROM accounts
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListAccountsByUserID :many
SELECT id, user_id, name, type, balance, created_at, updated_at, deleted_at
FROM accounts
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateAccount :one
UPDATE accounts
SET name = $2,
    type = $3,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, user_id, name, type, balance, created_at, updated_at, deleted_at;

-- name: SoftDeleteAccount :exec
UPDATE accounts
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateAccountBalance :exec
UPDATE accounts
SET balance = balance + $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
