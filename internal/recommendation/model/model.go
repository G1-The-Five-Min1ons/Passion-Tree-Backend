package model

import "time"

type UserReflection struct {
	ReflectID      string  `json:"reflect_id"`
	Summary        string  `json:"summary"`
	PrimaryEmotion string  `json:"primary_emotion"`
	StrugglePoint  string  `json:"struggle_point"`
	WeightedScore  float64 `json:"weighted_reflection_score"`
}

type RecommendedPath struct {
	PathID              string     `json:"path_id"`
	Title               string     `json:"title"`
	CoverImgURL         string     `json:"cover_img_url"`
	Objective           string     `json:"objective"`
	Description         string     `json:"description"`
	Rating              float64    `json:"avg_rating"`
	Publish_status      string     `json:"publish_status"`
	CreatedAt           *time.Time `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at"`
	CreatorID           string     `json:"creator_id"`
	Instructor          string     `json:"instructor"`
	Modules             int        `json:"modules"`
	Students            int        `json:"student"`
	RecommendationScore float64    `json:"recommendation_score"`
	Reason              string     `json:"reason"`
}

type RecommendPathResponse struct {
	UserPersonaQuery string            `json:"user_persona_query"`
	RecommendedPaths []RecommendedPath `json:"recommended_paths"`
}
