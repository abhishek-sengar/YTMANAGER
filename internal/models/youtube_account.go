package models

import "time"

type YouTubeAccount struct {
	ID           string    `db:"id"`
	UserID       string    `db:"user_id"`
	Email        string    `db:"email"`
	AccessToken  string    `db:"access_token"`
	RefreshToken string    `db:"refresh_token"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
