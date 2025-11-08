-- migrate:up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    provider VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    phone VARCHAR(10),
    profile_image VARCHAR,
    birth_date DATE,
    role VARCHAR(50) NOT NULL DEFAULT 'user'
);

-- migrate:down
DROP TABLE IF EXISTS users;
