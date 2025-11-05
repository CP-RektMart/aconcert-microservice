-- name: CreateReservation :one
INSERT INTO Reservation (
    user_id,
    event_id,
    status
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetReservation :one
SELECT * FROM Reservation
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetReservationByID :one
SELECT * FROM Reservation
WHERE id = $1
LIMIT 1;

-- name: ListReservations :many
SELECT * FROM Reservation
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListReservationsByUserID :many
SELECT * FROM Reservation
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListReservationsByEventID :many
SELECT * FROM Reservation
WHERE event_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListReservationsByStatus :many
SELECT * FROM Reservation
WHERE status = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateReservationStatus :one
UPDATE Reservation
SET status = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateReservation :one
UPDATE Reservation
SET
    user_id = COALESCE(sqlc.narg(user_id), user_id),
    event_id = COALESCE(sqlc.narg(event_id), event_id),
    status = COALESCE(sqlc.narg(status), status),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: DeleteReservation :exec
UPDATE Reservation
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: HardDeleteReservation :exec
DELETE FROM Reservation
WHERE id = $1;

-- name: CountReservationsByUserID :one
SELECT COUNT(*) FROM Reservation
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: CountReservationsByEventID :one
SELECT COUNT(*) FROM Reservation
WHERE event_id = $1 AND deleted_at IS NULL;
