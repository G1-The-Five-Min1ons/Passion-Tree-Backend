package model

type CreateNodeRequest struct {
	Title       string                      `json:"title" binding:"required"`
	Description string                      `json:"description"`
	PathID      string                      `json:"path_id" binding:"required"`
	Sequence    string                      `json:"sequence" binding:"required"`
	Materials   []CreateMaterialRequest     `json:"material,omitempty"`
	Questions   []CreateQuestionWithChoices `json:"Question,omitempty"`
}

type UpdateNodeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Node struct {
	NodeID      string         `json:"node_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Sequence    int            `json:"sequence"`
	PathID      string         `json:"path_id"`
	Materials   []NodeMaterial `json:"materials,omitempty"`
}

type NodeMaterial struct {
	MaterialID string `json:"material_id"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	NodeID     string `json:"node_id"`
}

type CreateMaterialRequest struct {
	Type   string `json:"type" binding:"required"`
	URL    string `json:"url" binding:"required"`
	NodeID string `json:"node_id"`
}

type ReorderNodesRequest struct {
	NodeIDs []string `json:"node_ids" binding:"required"`
}
