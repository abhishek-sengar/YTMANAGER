package models

type EditorChannel struct {
	EditorID  string `db:"editor_id"`
	ChannelID string `db:"channel_id"`
}
