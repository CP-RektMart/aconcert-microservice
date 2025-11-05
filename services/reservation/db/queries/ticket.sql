-- name: CreateTicket :one
INSERT INTO Ticket (
    reservation_id,
    zone_number,
    row_number,
    col_number,
    event_id
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetTicket :one
SELECT * FROM Ticket
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetTicketByID :one
SELECT * FROM Ticket
WHERE id = $1
LIMIT 1;

-- name: ListTickets :many
SELECT * FROM Ticket
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListTicketsByReservationID :many
SELECT * FROM Ticket
WHERE reservation_id = $1 AND deleted_at IS NULL
ORDER BY zone_number, row_number, col_number;

-- name: ListTicketsBySeat :many
SELECT * FROM Ticket
WHERE zone_number = $1
    AND row_number = $2
    AND col_number = $3
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetTicketBySeat :one
SELECT * FROM Ticket
WHERE zone_number = $1
    AND row_number = $2
    AND col_number = $3
    AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateTicket :one
UPDATE Ticket
SET
    reservation_id = COALESCE(sqlc.narg(reservation_id), reservation_id),
    zone_number = COALESCE(sqlc.narg(zone_number), zone_number),
    row_number = COALESCE(sqlc.narg(row_number), row_number),
    col_number = COALESCE(sqlc.narg(col_number), col_number),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTicketReservation :one
UPDATE Ticket
SET reservation_id = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTicket :exec
UPDATE Ticket
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteTicketsByReservationID :exec
UPDATE Ticket
SET deleted_at = NOW(), updated_at = NOW()
WHERE reservation_id = $1 AND deleted_at IS NULL;

-- name: HardDeleteTicket :exec
DELETE FROM Ticket
WHERE id = $1;

-- name: CountTicketsByReservationID :one
SELECT COUNT(*) FROM Ticket
WHERE reservation_id = $1 AND deleted_at IS NULL;

-- name: CheckSeatAvailability :one
SELECT EXISTS(
    SELECT 1 FROM Ticket
    WHERE zone_number = $1
        AND row_number = $2
        AND col_number = $3
        AND deleted_at IS NULL
) AS is_taken;

-- name: CheckSeatAvailabilityForEvent :one
SELECT EXISTS(
    SELECT 1 FROM Ticket
    WHERE event_id = $1
        AND zone_number = $2
        AND row_number = $3
        AND col_number = $4
        AND deleted_at IS NULL
) AS is_taken;

-- name: CreateTicketWithAvailabilityCheck :one
WITH seat_check AS (
    SELECT EXISTS(
        SELECT 1 FROM Ticket
        WHERE event_id = $1
            AND zone_number = $2
            AND row_number = $3
            AND col_number = $4
            AND deleted_at IS NULL
    ) AS is_taken
)
INSERT INTO Ticket (
    reservation_id,
    zone_number,
    row_number,
    col_number,
    event_id
)
SELECT $5, $2, $3, $4, $1
FROM seat_check
WHERE NOT is_taken
RETURNING *;
