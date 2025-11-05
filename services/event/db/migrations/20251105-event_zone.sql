-- migrate:up
CREATE TABLE event_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL,
    location_id VARCHAR NOT NULL,
    zone_number INT NOT NULL,
    price FLOAT NOT NULL DEFAULT 0,
    color VARCHAR NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_sold_out BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- migrate:down
DROP TABLE IF EXISTS event_zones;
