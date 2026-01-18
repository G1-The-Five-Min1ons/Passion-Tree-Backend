package model

import (
	nodeModel "passiontree/internal/learning-path/model"
)

type ResumeResponse struct {
	CurrentNode *nodeModel.Node `json:"current_node"`
	Message     string          `json:"message"`
}