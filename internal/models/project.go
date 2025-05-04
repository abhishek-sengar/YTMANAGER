package models

import "time"

type Project struct {
	ID          string    `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	VideoPath   string    `db:"video_path" json:"video_path"`
	Status      string    `db:"status" json:"status"` // pending, approved, rejected
	EditorID    string    `db:"editor_id" json:"editor_id"`
	OwnerID     string    `db:"owner_id" json:"owner_id"`
	ChannelID   string    `db:"channel_id" json:"channel_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
