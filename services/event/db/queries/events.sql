-- Get a single event by ID
-- name: GetEventByID :one
SELECT *
FROM events
WHERE id = $1
  AND deleted_at IS NULL;

-- List events with optional search and pagination
-- name: ListEvents :many
SELECT *
FROM events
WHERE
  deleted_at IS NULL
  AND (
    -- The query parameter is a string. If it's empty, this condition is true for all rows.
    -- Otherwise, it performs a case-insensitive search on the event name.
    sqlc.arg('query')::text = '' OR name ILIKE '%' || sqlc.arg('query') || '%'
  )
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

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

-- name: GetEventZonesByEventID :many
SELECT *
FROM event_zones
WHERE event_id = $1
  AND deleted_at IS NULL;

-- name: CreateEventZone :one
INSERT INTO event_zones (
    event_id, location_id, zone_number, price, color, name, description
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING event_id;

-- name: UpdateEventZone :one
UPDATE event_zones
SET
    event_id = $1,
    location_id = $2,
    zone_number = $3,
    price = $4,
    color = $5,
    name = $6,
    description = $7,
    is_sold_out = $8,
    updated_at = NOW()
WHERE id = $9
  AND deleted_at IS NULL
RETURNING event_id;

-- name: DeleteEventZone :one
UPDATE event_zones
SET deleted_at = NOW()
WHERE id = $1
RETURNING $1;

-- name: GetEventZoneByID :one
SELECT *
FROM event_zones
WHERE id = $1
  AND deleted_at IS NULL;
