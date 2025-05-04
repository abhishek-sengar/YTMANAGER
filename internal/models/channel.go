package models

import "time"

type Channel struct {
	ID               string    `db:"id"`
	OwnerID          string    `db:"owner_id"`
	YouTubeAccountID string    `db:"youtube_account_id"`
	YtChannelID      string    `db:"yt_channel_id"`
	Name             string    `db:"name"`
	IconURL          string    `db:"icon_url"`
	Email            string    `db:"email"`
	CreatedAt        time.Time `db:"created_at"`
}
