-- Get a single event by ID
-- name: GetEventByID :one
SELECT *
FROM events
WHERE id = $1
  AND deleted_at IS NULL;

-- List all events
-- name: ListEvents :many
SELECT *
FROM events
WHERE deleted_at IS NULL
ORDER BY event_date DESC;

-- Insert a new event
-- name: CreateEvent :one
INSERT INTO events (
    id, name, description, location_id, artist, event_date, thumbnail, images
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id;

-- Update an existing event
-- name: UpdateEvent :one
UPDATE events
SET
    name = $2,
    description = $3,
    location_id = $4,
    artist = $5,
    event_date = $6,
    thumbnail = $7,
    images = $8,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id;

-- Soft delete an event
-- name: DeleteEvent :one
UPDATE events
SET deleted_at = NOW()
WHERE id = $1
RETURNING $1;

-- Hard delete an event (for admin use)
-- name: HardDeleteEvent :one
DELETE FROM events
WHERE id = $1
RETURNING $1;

