-- name: GetUserByProviderEmail :one
SELECT *
FROM users
WHERE provider = $1 AND email = $2 AND deleted_at IS NULL;

-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (provider, email, first_name, last_name, profile_image, role) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET first_name = $2,
    last_name = $3,
    profile_image = $4,
    birth_date = $5,
    phone = $6,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
