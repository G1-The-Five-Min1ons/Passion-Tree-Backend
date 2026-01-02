package learningpath

import "time"

type CreatePathRequest struct {
	Title       string `json:"title" binding:"required"`
	Objective   string `json:"objective"`
	Description string `json:"description"`
	CoverImgURL string `json:"cover_img_url"`
	Status      string `json:"status"`
	CreatorID   string `json:"creator_id"`
}

type UpdatePathRequest struct {
	Title       string `json:"title"`
	Objective   string `json:"objective"`
	Description string `json:"description"`
	CoverImgURL string `json:"cover_img_url"`
	Status      string `json:"status"`
}

type StartPathRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type CreateNodeRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	PathID      string `json:"path_id" binding:"required"`
}

type UpdateNodeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateMaterialRequest struct {
	Type   string `json:"type" binding:"required"`
	URL    string `json:"url" binding:"required"`
	NodeID string `json:"node_id" binding:"required"`
}

type CreateCommentRequest struct {
	ParentID *string `json:"parent_id"`
	Content  string  `json:"content" binding:"required"`
	NodeID   string  `json:"node_id" binding:"required"`
}

type CreateReactionRequest struct {
	ReactionType string `json:"reaction_type" binding:"required"`
	CommentID    string `json:"comment_id" binding:"required"`
}

type CreateMentionRequest struct {
	CommentID string `json:"comment_id" binding:"required"`
}

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

type LearningPath struct {
	PathID      string    `json:"path_id"`
	Title       string    `json:"title"`
	CoverImgURL string    `json:"cover_img_url"`
	Objective   string    `json:"objective"`
	Description string    `json:"description"`
	AvgRating   float64   `json:"avg_rating"`
	Status      string    `json:"status"`
	CreatorID   string    `json:"creator_id"`
	CreatedAt   time.Time `json:"create_at"`
	UpdatedAt   time.Time `json:"update_at"`
	Nodes       []Node    `json:"nodes,omitempty"`
}

type PathEnroll struct {
	EnrollID   string     `json:"enroll_id"`
	Status     string     `json:"status"`
	EnrollAt   time.Time  `json:"enroll_at"`
	CompleteAt *time.Time `json:"complete_at"`
}

type Node struct {
	NodeID      string         `json:"node_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	PathID      string         `json:"path_id"`
	Materials   []NodeMaterial `json:"materials,omitempty"`
}

type NodeMaterial struct {
	MaterialID string `json:"material_id"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	NodeID     string `json:"node_id"`
}

type NodeComment struct {
	CommentID string            `json:"comment_id"`
	ParentID  *string           `json:"parent_id"`
	Content   string            `json:"content"`
	CreatedAt time.Time         `json:"create_at"`
	EditAt    *time.Time        `json:"edit_at"`
	NodeID    string            `json:"node_id"`
	Reactions []CommentReaction `json:"reactions,omitempty"`
	Mentions  []CommentMention  `json:"mentions,omitempty"`
}

type CommentReaction struct {
	ReactionID   string `json:"reaction_id"`
	ReactionType string `json:"reaction_type"`
	CommentID    string `json:"comment_id"`
}

type CommentMention struct {
	MentionID string    `json:"mention_id"`
	CreatedAt time.Time `json:"create_at"`
	CommentID string    `json:"comment_id"`
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