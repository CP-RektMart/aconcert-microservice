-- name: CreateReservationTicket :exec
INSERT INTO ReservationTicket (
    reservation_id,
    ticket_id
) VALUES (
    $1, $2
);

-- name: GetReservationTickets :many
SELECT
    rt.reservation_id,
    rt.ticket_id,
    t.*
FROM ReservationTicket rt
JOIN Ticket t ON rt.ticket_id = t.id
WHERE rt.reservation_id = $1 AND t.deleted_at IS NULL;

-- name: GetTicketReservations :many
SELECT
    rt.reservation_id,
    rt.ticket_id,
    r.*
FROM ReservationTicket rt
JOIN Reservation r ON rt.reservation_id = r.id
WHERE rt.ticket_id = $1 AND r.deleted_at IS NULL;

-- name: ListAllReservationTickets :many
SELECT * FROM ReservationTicket;

-- name: DeleteReservationTicket :exec
DELETE FROM ReservationTicket
WHERE reservation_id = $1 AND ticket_id = $2;

-- name: DeleteReservationTicketsByReservationID :exec
DELETE FROM ReservationTicket
WHERE reservation_id = $1;

-- name: DeleteReservationTicketsByTicketID :exec
DELETE FROM ReservationTicket
WHERE ticket_id = $1;

-- name: CountTicketsInReservation :one
SELECT COUNT(*) FROM ReservationTicket
WHERE reservation_id = $1;

-- name: CountReservationsForTicket :one
SELECT COUNT(*) FROM ReservationTicket
WHERE ticket_id = $1;

-- name: CheckReservationTicketExists :one
SELECT EXISTS(
    SELECT 1 FROM ReservationTicket
    WHERE reservation_id = $1 AND ticket_id = $2
) AS exists;

-- name: BulkCreateReservationTickets :copyfrom
INSERT INTO ReservationTicket (
    reservation_id,
    ticket_id
) VALUES (
    $1, $2
);
