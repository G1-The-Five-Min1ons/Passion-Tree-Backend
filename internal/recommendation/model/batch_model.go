package model

type UserInteraction struct {
	UserID string  `json:"user_id"`
	PathID string  `json:"path_id"`
	Score  float64 `json:"score"`
}

type UserProfile struct {
	UserID    string `json:"user_id"`
	Interests string `json:"interests"` 
}

type BatchRecommendPayload struct {
	Interactions []UserInteraction `json:"users_interactions"`
	Profiles     []UserProfile     `json:"users_profiles"`
}