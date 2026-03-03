package model

import "time"

type CreateCommentRequest struct {
	// UserID is populated server-side from the JWT token; clients must NOT send it
	UserID   string  `json:"-"`
	NodeID   string  `json:"node_id"`
	Message  string  `json:"message"`
	ParentID *string `json:"parent_id"`
}

type UpdateCommentRequest struct {
	Message string `json:"message"`
}

// Allowed reaction types
const (
	ReactionLike  = "like"
	ReactionLove  = "love"
	ReactionHaha  = "haha"
	ReactionWow   = "wow"
	ReactionSad   = "sad"
	ReactionAngry = "angry"
)

// AllowedReactionTypes is the set of valid reaction types.
var AllowedReactionTypes = map[string]bool{
	ReactionLike:  true,
	ReactionLove:  true,
	ReactionHaha:  true,
	ReactionWow:   true,
	ReactionSad:   true,
	ReactionAngry: true,
}

// IsValidReactionType checks whether the given reaction type is allowed.
func IsValidReactionType(rt string) bool {
	return AllowedReactionTypes[rt]
}

type CreateReactionRequest struct {
	UserID       string `json:"-"`
	ReactionType string `json:"reaction_type"`
	CommentID    string `json:"-"`
}

type CreateMentionRequest struct {
	// Set server-side from JWT token — who is doing the mentioning
	MentionerUserID string `json:"-"`
	// Sent by client — who is being mentioned
	MentionedUserID string `json:"mentioned_user_id"`
	// Set server-side from URL param
	CommentID string `json:"-"`
}
type NodeComment struct {
	UserID    string            `json:"user_id"`
	UserName  string            `json:"user_name,omitempty"`
	CommentID string            `json:"comment_id"`
	Message   string            `json:"message"`
	CreatedAt time.Time         `json:"create_at"`
	EditAt    *time.Time        `json:"edit_at"`
	NodeID    string            `json:"node_id"`
	ParentID  *string           `json:"parent_id"`
	Reactions []CommentReaction `json:"reactions,omitempty"`
	Mentions  []CommentMention  `json:"mentions,omitempty"`
}

type CommentReaction struct {
	ReactionID   string `json:"reaction_id"`
	ReactionType string `json:"reaction_type"`
	CommentID    string `json:"comment_id"`
	UserID       string `json:"user_id,omitempty"` // Added to track who reacted
}

type CommentMention struct {
	MentionID       string    `json:"mention_id"`
	CreatedAt       time.Time `json:"create_at"`
	CommentID       string    `json:"comment_id"`
	MentionedUserID string    `json:"mentioned_user_id"`
}
