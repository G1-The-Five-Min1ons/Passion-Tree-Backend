package model

import "time"

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