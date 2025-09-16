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
