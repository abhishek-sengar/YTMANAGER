-- +goose Up
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    youtube_account_id UUID REFERENCES youtube_accounts(id),
    yt_channel_id TEXT NOT NULL,
    name TEXT NOT NULL,
    icon_url TEXT,
    email TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (owner_id, yt_channel_id)
);

-- +goose Down
DROP TABLE IF EXISTS channels;
