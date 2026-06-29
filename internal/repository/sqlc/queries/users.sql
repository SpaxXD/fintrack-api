-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, password_hash, failed_attempts, locked_until, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, failed_attempts, locked_until, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, failed_attempts, locked_until, created_at, updated_at
FROM users
WHERE id = $1;

-- name: IncrementFailedAttempts :exec
UPDATE users
SET failed_attempts = failed_attempts + 1,
    locked_until = CASE
        WHEN failed_attempts + 1 >= 5 THEN NOW() + INTERVAL '15 minutes'
        ELSE locked_until
    END,
    updated_at = NOW()
WHERE email = $1;

-- name: ResetFailedAttempts :exec
UPDATE users
SET failed_attempts = 0,
    locked_until = NULL,
    updated_at = NOW()
WHERE email = $1;

-- name: GetFailedAttempts :one
SELECT failed_attempts, locked_until
FROM users
WHERE email = $1;
