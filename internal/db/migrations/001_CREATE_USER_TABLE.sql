-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) NOT NULL,
    youtube_access_token TEXT,
    youtube_refresh_token TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);


-- +goose Down
DROP TABLE IF EXISTS users;
