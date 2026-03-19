package model

import (
	"time"
)

type Reflection struct {
	ReflectID               string    `json:"reflect_id"`
	ReflectScore            string    `json:"reflect_score"`
	ReflectDescription      string    `json:"reflect_description"`
	Reflect                 string    `json:"reflect"`
	ProgressScore           string    `json:"progress_score"`
	ChallengeScore          string    `json:"challenge_score"`
	Summary                 string    `json:"summary"`
	SentimentAnalysis       string    `json:"sentiment_analysis"`
	PrimaryEmotion          *string   `json:"primary_emotion,omitempty"`
	StrugglePoint           string    `json:"struggle_point"`
	AIConfidentScore        float64   `json:"ai_confident_score"`
	ReflectionScore         float64   `json:"reflection_score"`
	WeightedReflectionScore float64   `json:"weighted_reflection_score"`
	CreatedAt               time.Time `json:"created_at"`
	TreeNodeID              string    `json:"tree_node_id"`
}

type CreateReflectionRequest struct {
	LearningReflect string `json:"learning_reflect"`
	MoodReflect     string `json:"mood_reflect"`
	FeelScore       int    `json:"feel_score"`
	ProgressScore   int    `json:"progress_score"`
	ChallengeScore  int    `json:"challenge_score"`
	TreeNodeID      string `json:"tree_node_id"`
}

type ReflectionResponse struct {
	ReflectID               string   `json:"reflect_id"`
	Summary                 string   `json:"summary"`
	SentimentAnalysis       string   `json:"sentiment_analysis"`
	PrimaryEmotion          *string  `json:"primary_emotion,omitempty"`
	StrugglePoint           string   `json:"struggle_point"`
	DevelopmentPlan         []string `json:"development_plan"`
	AIConfidentScore        float64  `json:"ai_confident_score"`
	ReflectionScore         float64  `json:"reflection_score"`
	WeightedReflectionScore float64  `json:"weighted_reflection_score"`
}

type UpdateReflectionRequest struct {
	LearningReflect string `json:"learning_reflect"`
	MoodReflect     string `json:"mood_reflect"`
	FeelScore       int    `json:"feel_score"`
	ProgressScore   int    `json:"progress_score"`
	ChallengeScore  int    `json:"challenge_score"`
}

// GetReflectionsFilter represents filter parameters for querying reflections
type GetReflectionsFilter struct {
	TreeNodeID *string `json:"tree_node_id,omitempty"`
	TreeID     *string `json:"tree_id,omitempty"`
	AlbumID    *string `json:"album_id,omitempty"`
	UserID     *string `json:"user_id,omitempty"`
	// Keyset cursor: fetch rows strictly older than this (create_at, reflect_id) pair.
	BeforeCreatedAt *time.Time `json:"before_created_at,omitempty"`
	BeforeReflectID *string    `json:"before_reflect_id,omitempty"`
	Limit      int     `json:"limit,omitempty"`
	// Deprecated: Use keyset cursor fields (BeforeCreatedAt + BeforeReflectID) instead.
	Offset     int     `json:"offset,omitempty"`
}

// ReflectionStats represents statistics for reflections
type ReflectionStats struct {
	TotalReflections   int     `json:"total_reflections"`
	ThisWeek           int     `json:"this_week"`
	ThisMonth          int     `json:"this_month"`
	UniqueAuthors      int     `json:"unique_authors"`
	AvgReflectionScore float64 `json:"avg_reflection_score"`
	TotalTrees         int     `json:"total_trees"`
	TotalAlbums        int     `json:"total_albums"`
}
