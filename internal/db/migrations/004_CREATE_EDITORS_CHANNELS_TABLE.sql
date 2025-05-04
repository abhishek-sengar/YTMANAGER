-- +goose Up
CREATE TABLE editors_channels (
    editor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    PRIMARY KEY (editor_id, channel_id)
);

-- +goose Down
DROP TABLE IF EXISTS editors_channels;
