package handler

import (
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// GetComments godoc
// @Summary      List comments on a node
// @Tags         Comments
// @Produce      json
// @Security     BearerAuth
// @Param        node_id  path      string  true  "Node ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/nodes/{node_id}/comments [get]
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

// GetPathComments godoc
// @Summary      List comments on a learning path
// @Tags         Comments
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string  true  "Learning path ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/{path_id}/comments [get]
func (h *Handler) GetPathComments(c *fiber.Ctx) error {
	ctx := c.UserContext()
	comments, err := h.commentSvc.GetPathComments(ctx, c.Params("path_id"))
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Path comments retrieved successfully",
		"data":    comments,
	})
}

// CreateComment godoc
// @Summary      Create a comment on a node
// @Description  Author is taken from the JWT, not the body.
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        node_id  path      string                      true  "Node ID"
// @Param        body     body      model.CreateCommentRequest  true  "Comment payload"
// @Success      201      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/nodes/{node_id}/comments [post]
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

	req.UserID = userID
	nodeID := c.Params("node_id")
	req.NodeID = &nodeID

	id, err := h.commentSvc.AddComment(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "comment created successfully", "comment_id", id, "user_id", userID, "node_id", nodeID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Comment created successfully",
		"data": fiber.Map{
			"comment_id": id,
			"message":    req.Message,
			"user_id":    userID,
			"node_id":    nodeID,
			"parent_id":  req.ParentID,
		},
	})
}

// CreatePathComment godoc
// @Summary      Create a comment on a learning path
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string                      true  "Learning path ID"
// @Param        body     body      model.CreateCommentRequest  true  "Comment payload"
// @Success      201      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/{path_id}/comments [post]
func (h *Handler) CreatePathComment(c *fiber.Ctx) error {
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

	req.UserID = userID
	pathID := c.Params("path_id")
	req.PathID = &pathID

	id, err := h.commentSvc.AddComment(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "path comment created successfully", "comment_id", id, "user_id", userID, "path_id", pathID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Path comment created successfully",
		"data": fiber.Map{
			"comment_id": id,
			"message":    req.Message,
			"user_id":    userID,
			"path_id":    pathID,
			"parent_id":  req.ParentID,
		},
	})
}

// UpdateComment godoc
// @Summary      Update one of your own comments
// @Description  Returns 403 if the comment belongs to another user.
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        comment_id  path      string                      true  "Comment ID"
// @Param        body        body      model.UpdateCommentRequest  true  "Updated message"
// @Success      200         {object}  apidoc.SuccessResponse
// @Failure      400         {object}  apidoc.ErrorResponse
// @Failure      401         {object}  apidoc.ErrorResponse
// @Failure      403         {object}  apidoc.ErrorResponse
// @Failure      404         {object}  apidoc.ErrorResponse
// @Router       /learningpaths/comments/{comment_id} [put]
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

	h.logger.InfoContext(ctx, "comment updated successfully", "comment_id", commentID, "user_id", userID)

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

// DeleteComment godoc
// @Summary      Delete one of your own comments
// @Tags         Comments
// @Produce      json
// @Security     BearerAuth
// @Param        comment_id  path      string  true  "Comment ID"
// @Success      200         {object}  apidoc.SuccessResponse
// @Failure      401         {object}  apidoc.ErrorResponse
// @Failure      403         {object}  apidoc.ErrorResponse
// @Failure      404         {object}  apidoc.ErrorResponse
// @Router       /learningpaths/comments/{comment_id} [delete]
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

	h.logger.InfoContext(ctx, "comment deleted successfully", "comment_id", commentID, "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Comment deleted successfully",
		"data": fiber.Map{
			"comment_id": commentID,
		},
	})
}

// CreateReaction godoc
// @Summary      Toggle a reaction on a comment
// @Description  Adds the reaction if it doesn't exist; removes it if the same user has already reacted.
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        comment_id  path      string                       true  "Comment ID"
// @Param        body        body      model.CreateReactionRequest  true  "Reaction type"
// @Success      200         {object}  apidoc.SuccessResponse
// @Failure      400         {object}  apidoc.ErrorResponse
// @Failure      401         {object}  apidoc.ErrorResponse
// @Router       /learningpaths/comments/{comment_id}/reactions [post]
func (h *Handler) CreateReaction(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	ctx := c.UserContext()

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, apperror.NewUnauthorized("unauthorized: %s", err.Error()))
	}

	var req model.CreateReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	if req.ReactionType == "" {
		return h.handleError(c, apperror.NewBadRequest("reaction_type is required"))
	}
	if !model.IsValidReactionType(req.ReactionType) {
		return h.handleError(c, apperror.NewBadRequest("invalid reaction_type: must be one of like, love, haha, wow, sad, angry"))
	}
	req.CommentID = commentID
	req.UserID = userID

	added, err := h.commentSvc.ToggleReaction(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	msg := "Reaction removed successfully"
	if added {
		msg = "Reaction added successfully"
	}
	h.logger.InfoContext(ctx, msg, "comment_id", commentID, "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": msg,
		"data": fiber.Map{
			"comment_id": commentID,
			"added":      added,
		},
	})
}

// CreateMention godoc
// @Summary      Mention another user in a comment
// @Description  The mentioning user is taken from the JWT; the mentioned user comes from the request body.
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        comment_id  path      string                      true  "Comment ID"
// @Param        body        body      model.CreateMentionRequest  true  "Mentioned user payload"
// @Success      201         {object}  apidoc.SuccessResponse
// @Failure      400         {object}  apidoc.ErrorResponse
// @Failure      401         {object}  apidoc.ErrorResponse
// @Router       /learningpaths/comments/{comment_id}/mentions [post]
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

	h.logger.InfoContext(ctx, "mention created successfully", "mention_id", id, "comment_id", req.CommentID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Mention created successfully",
		"data": fiber.Map{
			"mention_id":        id,
			"comment_id":        req.CommentID,
			"mentioner_id":      mentionerID,
			"mentioned_user_id": req.MentionedUserID,
		},
	})
}
