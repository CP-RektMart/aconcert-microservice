-- users.sql

CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    email TEXT NOT NULL UNIQUE,
    firstname TEXT NOT NULL,
    lastname TEXT NOT NULL,
    phone TEXT,
    profile_image TEXT,
    birth_date DATE,
    role TEXT NOT NULL CHECK (role IN ('USER', 'ADMIN'))
);