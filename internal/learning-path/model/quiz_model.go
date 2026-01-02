package model

type CreateQuestionRequest struct {
	QuestionText string `json:"question_text" binding:"required"`
	Type         string `json:"type" binding:"required"`
	NodeID       string `json:"node_id" binding:"required"`
}

type CreateChoiceRequest struct {
	ChoiceText string `json:"choice_text" binding:"required"`
	IsCorrect  bool   `json:"is_correct"`
	Reasoning  string `json:"reasoning"`
	QuestionID string `json:"question_id" binding:"required"`
}

type NodeQuestion struct {
	QuestionID   string           `json:"question_id"`
	QuestionText string           `json:"question_text"`
	Type         string           `json:"type"`
	NodeID       string           `json:"node_id"`
	Choices      []QuestionChoice `json:"choices,omitempty"`
}

type QuestionChoice struct {
	ChoiceID   string `json:"choice_id"`
	ChoiceText string `json:"choice_text"`
	IsCorrect  bool   `json:"is_correct"`
	Reasoning  string `json:"reasoning"`
	QuestionID string `json:"question_id"`
}