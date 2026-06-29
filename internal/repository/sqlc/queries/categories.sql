-- name: CreateCategory :one
INSERT INTO categories (user_id, name, type)
VALUES ($1, $2, $3)
RETURNING id, user_id, name, type, created_at, updated_at, deleted_at;

-- name: GetCategoryByID :one
SELECT id, user_id, name, type, created_at, updated_at, deleted_at
FROM categories
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListCategoriesByUserID :many
SELECT id, user_id, name, type, created_at, updated_at, deleted_at
FROM categories
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, user_id, name, type, created_at, updated_at, deleted_at;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateDefaultCategories :exec
INSERT INTO categories (user_id, name, type) VALUES
    ($1, 'Salário', 'income'),
    ($1, 'Freelance', 'income'),
    ($1, 'Investimentos', 'income'),
    ($1, 'Outros (Receita)', 'income'),
    ($1, 'Alimentação', 'expense'),
    ($1, 'Transporte', 'expense'),
    ($1, 'Moradia', 'expense'),
    ($1, 'Saúde', 'expense'),
    ($1, 'Educação', 'expense'),
    ($1, 'Lazer', 'expense'),
    ($1, 'Outros (Despesa)', 'expense');
