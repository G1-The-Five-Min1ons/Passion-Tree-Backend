package learningpath

type Node struct {
	ID          int    `json:"node_id"`
	PathID      int    `json:"path_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoURL    string `json:"video_url"`
	FileURL     string `json:"file_url"`
	Order       int    `json:"order"`
}

type CreateNodeRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	VideoURL    string `json:"video_url"`
	FileURL     string `json:"file_url"`
	Order       int    `json:"order"`
}

type UpdateNodeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoURL    string `json:"video_url"`
	FileURL     string `json:"file_url"`
	Order       int    `json:"order"`
}