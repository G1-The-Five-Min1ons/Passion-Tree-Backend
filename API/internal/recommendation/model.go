package recommendation

type RecommendedItem struct {
	ID    int     `json:"learningpath_id"`
	Name  string  `json:"name"`
	Score float64 `json:"recommendation_score"`
}
