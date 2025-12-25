package learningpath

type LearningPath struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
	Nodes       []Node `json:"nodes,omitempty"`
}

type Node struct {
	ID          int    `json:"node_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

type NodeProgress struct {
	Node
	Status string `json:"status"`
}

type CreatePathRequest struct {
	Title       string `json:"title" binding:"required"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
}

type UpdatePathRequest struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
}

type StartPathRequest struct {
	UserID int `json:"user_id" binding:"required"`
}