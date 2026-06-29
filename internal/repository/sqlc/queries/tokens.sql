-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token, expires_at, revoked, created_at;

-- name: GetRefreshTokenByToken :one
SELECT id, user_id, token, expires_at, revoked, created_at
FROM refresh_tokens
WHERE token = $1 AND revoked = FALSE;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked = TRUE
WHERE id = $1;

-- name: RevokeAllRefreshTokensByUserID :exec
UPDATE refresh_tokens
SET revoked = TRUE
WHERE user_id = $1 AND revoked = FALSE;
