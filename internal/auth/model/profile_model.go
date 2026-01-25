package model

import "time"

type Profile struct {
	ProfileID      string    `json:"profile_id"`
	AvatarURL      string    `json:"avatar_url"`
	RankName       string    `json:"rank_name"`
	LearningStreak int       `json:"learning_streak"`
	LearningCount  int       `json:"learning_count"`
	Location       string    `json:"location"`
	Bio            string    `json:"bio"`
	Level          int       `json:"level"`
	XP             int64     `json:"xp"`
	HourLearned    int       `json:"hour_learned"`
	UserID         string    `json:"user_id"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}
