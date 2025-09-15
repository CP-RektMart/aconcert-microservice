-- migrate:up
CREATE TABLE events (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    name TEXT NOT NULL,
    description TEXT,
    location_id UUID NOT NULL,
    artist TEXT[] NOT NULL,
    event_date TIMESTAMPTZ NOT NULL,
    thumbnail TEXT,
    images TEXT[]
);

-- migrate:down
DROP TABLE IF EXISTS events;