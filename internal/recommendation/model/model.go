package model

type UserReflection struct {
	ReflectID      string  `json:"reflect_id"`
	Summary        string  `json:"summary"`
	PrimaryEmotion string  `json:"primary_emotion"`
	StrugglePoint  string  `json:"struggle_point"`
	WeightedScore  float64 `json:"weighted_reflection_score"`
}

type RecommendedPath struct {
	PathID              string  `json:"path_id"`
	Title               string  `json:"title"`
	Description         string  `json:"description,omitempty"`
	CoverImgURL         string  `json:"cover_img_url,omitempty"`
	Objective           string  `json:"objective,omitempty"`
	RecommendationScore float64 `json:"recommendation_score"`
	Reason              string  `json:"reason"`
}

type RecommendPathResponse struct {
	UserPersonaQuery string            `json:"user_persona_query"`
	RecommendedPaths []RecommendedPath `json:"recommended_paths"`
}
