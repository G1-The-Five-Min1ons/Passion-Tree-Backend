package recommendation

type RecommendedItem struct {
	ID    int     `json:"learningpath_id"`
	Name  string  `json:"name"`
	Score float64 `json:"recommendation_score"`
}

type GeneralRequest struct {
	UserID int `json:"user_id"`
}

type TreeRequest struct {
	UserID         int   `json:"user_id"`
	CompletedNodes []int `json:"completed_nodes,omitempty"`
}