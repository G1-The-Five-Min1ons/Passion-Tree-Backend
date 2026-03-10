package model

import "time"

// Setting represents a user setting
type Setting struct {
	SettingID string    `db:"id" json:"setting_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Key       string    `db:"key" json:"key"`
	Value     string    `db:"value" json:"value"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// SettingRequest represents a request to create/update a setting
type SettingRequest struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}
