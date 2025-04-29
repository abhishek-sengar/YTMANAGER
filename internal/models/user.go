package models

import "time"

type User struct {
	ID                  string    `db:"id" json:"id"`
	Name                string    `db:"name" json:"name"`
	Email               string    `db:"email" json:"email"`
	PasswordHash        string    `db:"password_hash" json:"-"`
	Role                string    `db:"role" json:"role"` // "editor" or "owner"
	YoutubeAccessToken  string    `db:"youtube_access_token" json:"-"`
	YoutubeRefreshToken string    `db:"youtube_refresh_token" json:"-"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
}
