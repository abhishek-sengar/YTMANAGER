-- +goose Up
CREATE TABLE notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    timestamp INT NOT NULL, -- in seconds
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS notes;
