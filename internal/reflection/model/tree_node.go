package model

import "time"

// TreeNode represents a node in the tree from Tree_Node table
type TreeNode struct {
	TreeNodeID string     `json:"tree_node_id"`
	NodeTitle  string     `json:"node_title"`
	NodeID     string     `json:"node_id"`
	NodeScore  *float64   `json:"node_score"`
	CreatedAt  time.Time  `json:"created_at"`
	TreeID     string     `json:"tree_id"`
	ChildNode  *string    `json:"child_node"`
	Sequence   int        `json:"sequence"`
}

// CreateTreeNodeRequest represents the request to create a tree node
type CreateTreeNodeRequest struct {
	NodeTitle string  `json:"node_title" binding:"required"`
	NodeID    string  `json:"node_id" binding:"required"`
	TreeID    string  `json:"tree_id" binding:"required"`
	ChildNode *string `json:"child_node"`
}

// UpdateTreeNodeRequest represents the request to update a tree node
type UpdateTreeNodeRequest struct {
	NodeTitle string   `json:"node_title"`
	NodeScore *float64 `json:"node_score"`
	ChildNode *string  `json:"child_node"`
}

// TreeNodeResponse represents the response for tree node operations
type TreeNodeResponse struct {
	TreeNodeID string     `json:"tree_node_id"`
	NodeTitle  string     `json:"node_title"`
	NodeID     string     `json:"node_id"`
	NodeScore  *float64   `json:"node_score"`
	CreatedAt  time.Time  `json:"created_at"`
	TreeID     string     `json:"tree_id"`
	ChildNode  *string    `json:"child_node"`
	Sequence   int        `json:"sequence"`
}
