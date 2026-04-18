package model

import (
	pathmodel "passiontree/internal/learning-path/model"
)

type UserReflection struct {
	ReflectID      string  `json:"reflect_id"`
	Summary        string  `json:"summary"`
	PrimaryEmotion string  `json:"primary_emotion"`
	StrugglePoint  string  `json:"struggle_point"`
	WeightedScore  float64 `json:"weighted_reflection_score"`
}

type RecommendedPath struct {
	pathmodel.LearningPath
	RecommendationScore float64 `json:"recommendation_score"`
	Reason              string  `json:"reason"`
}

type RecommendPathResponse struct {
	UserPersonaQuery string            `json:"user_persona_query"`
	RecommendedPaths []RecommendedPath `json:"recommended_paths"`
}

type BatchRecommendationResult struct {
	UserID           string   `json:"user_id"`
	RecommendedPaths []string `json:"recommended_paths"`
}

type BatchRecommendResponse struct {
	Success bool                        `json:"success"`
	Message string                      `json:"message"`
	Data    []BatchRecommendationResult `json:"data"`
}
