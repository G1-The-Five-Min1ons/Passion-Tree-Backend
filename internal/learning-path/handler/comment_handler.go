package handler

import (
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// GetComments returns all comments for a node. Public read-only.
func (h *Handler) GetComments(c *fiber.Ctx) error {
	ctx := c.UserContext()
	comments, err := h.commentSvc.GetNodeComments(ctx, c.Params("node_id"))
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Comments retrieved successfully",
		"data":    comments,
	})
}

// CreateComment creates a comment. The user_id is taken from the JWT token.
func (h *Handler) CreateComment(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// Extract user_id from the validated JWT token
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, apperror.NewUnauthorized("unauthorized: %s", err.Error()))
	}

	var req model.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	// Populate user_id from the token, not from the body
	req.UserID = userID
	// Populate node_id from the URL param, not from the body
	req.NodeID = c.Params("node_id")

	id, err := h.commentSvc.AddComment(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "comment created", "comment_id", id, "user_id", userID, "node_id", req.NodeID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user_id":    userID,
		"comment_id": id,
		"message":    req.Message,
		"create_at":  nil,
		"edit_at":    nil,
		"node_id":    req.NodeID,
		"parent_id":  req.ParentID,
	})
}

// UpdateComment updates a comment. Only the token owner can update their own comment.
// Returns 403 if the comment belongs to another user.
func (h *Handler) UpdateComment(c *fiber.Ctx) error {
	ctx := c.UserContext()
	commentID := c.Params("comment_id")

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, apperror.NewUnauthorized("unauthorized: %s", err.Error()))
	}

	var req model.UpdateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	if req.Message == "" {
		return h.handleError(c, apperror.NewBadRequest("message is required"))
	}

	if err := h.commentSvc.UpdateComment(ctx, userID, commentID, req.Message); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "comment updated", "comment_id", commentID, "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Comment updated successfully",
		"data": fiber.Map{
			"comment_id": commentID,
			"user_id":    userID,
			"message":    req.Message,
		},
	})
}

// DeleteComment deletes a comment. Only the token owner can delete their own comment.
func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	ctx := c.UserContext()

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, apperror.NewUnauthorized("unauthorized: %s", err.Error()))
	}

	if err := h.commentSvc.RemoveComment(ctx, userID, commentID); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "comment deleted", "comment_id", commentID, "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Comment deleted successfully",
		"data": fiber.Map{
			"comment_id": commentID,
		},
	})
}

// CreateReaction adds a reaction to a comment. Requires authentication.
func (h *Handler) CreateReaction(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	ctx := c.UserContext()

	if _, err := middleware.GetUserIDFromContext(c); err != nil {
		return h.handleError(c, apperror.NewUnauthorized("unauthorized: %s", err.Error()))
	}

	var req model.CreateReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	req.CommentID = commentID

	if err := h.commentSvc.AddReaction(ctx, req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reaction added", "comment_id", commentID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Reaction added successfully",
		"data": fiber.Map{
			"comment_id": commentID,
		},
	})
}

// CreateMention creates a mention on a comment.
// mentioner_user_id is sourced from the JWT token; mentioned_user_id is from the request body.
func (h *Handler) CreateMention(c *fiber.Ctx) error {
	ctx := c.UserContext()

	mentionerID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, apperror.NewUnauthorized("unauthorized: %s", err.Error()))
	}

	var req model.CreateMentionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	if req.MentionedUserID == "" {
		return h.handleError(c, apperror.NewBadRequest("mentioned_user_id is required"))
	}

	req.MentionerUserID = mentionerID
	req.CommentID = c.Params("comment_id")

	id, err := h.commentSvc.AddMention(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "mention created", "mention_id", id, "comment_id", req.CommentID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Mention created successfully",
		"data": fiber.Map{
			"mention_id":       id,
			"comment_id":       req.CommentID,
			"mentioner_id":     mentionerID,
			"mentioned_user_id": req.MentionedUserID,
		},
	})
}
