-- +goose Up
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    video_path TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    editor_id UUID REFERENCES users(id) ON DELETE CASCADE,
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS projects;
