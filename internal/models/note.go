package models

import "time"

type Note struct {
	ID        string    `db:"id" json:"id"`
	ProjectID string    `db:"project_id" json:"project_id"`
	Timestamp int       `db:"timestamp" json:"timestamp"` // seconds
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
