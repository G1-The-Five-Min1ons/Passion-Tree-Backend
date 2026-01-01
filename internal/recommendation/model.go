package recommendation

type RecommendedItem struct {
	Path_id    string  `json:"learningpath_id"`
}

type GeneralRequest struct {
	User_id string `json:"user_id"`
}

type TreeRequest struct {
	User_id string `json:"user_id"`
	Tree_id string `json:"tree_id"`
}
