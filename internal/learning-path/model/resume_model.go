package model

type ResumeResponse struct {
	CurrentNode *Node  `json:"current_node"`
	Message     string `json:"message"`
}
