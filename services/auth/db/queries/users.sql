-- Get a single user by ID
-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- Get a user by email
-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
  AND deleted_at IS NULL;

-- List all users
-- name: ListUsers :many
SELECT *
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- Insert a new user
-- name: CreateUser :one
INSERT INTO users (
    id, email, firstname, lastname, phone, profile_image, birth_date, role
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id;

-- Update an existing user
-- name: UpdateUser :one
UPDATE users
SET
    email = $2,
    firstname = $3,
    lastname = $4,
    phone = $5,
    profile_image = $6,
    birth_date = $7,
    role = $8,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING $1;

-- Soft delete a user
-- name: DeleteUser :one
UPDATE users
SET deleted_at = NOW()
WHERE id = $1
RETURNING $1;

-- Hard delete a user (for admin use)
-- name: HardDeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING $1;