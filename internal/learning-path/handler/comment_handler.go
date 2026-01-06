package handler

import (
	"passiontree/internal/learning-path/model"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetComments(c *fiber.Ctx) error {
	comments, err := h.commentSvc.GetNodeComments(c.Params("node_id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    comments,
	})
}

func (h *Handler) CreateComment(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req model.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.NodeID = nodeID

	id, err := h.commentSvc.AddComment(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Comment created successfully",
		"data": fiber.Map{
			"comment_id": id,
		},
	})
}

func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	if err := h.commentSvc.RemoveComment(commentID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Comment deleted successfully",
		"data": fiber.Map{
			"comment_id": commentID,
		},
	})
}

func (h *Handler) CreateReaction(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	var req model.CreateReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.CommentID = commentID

	if err := h.commentSvc.AddReaction(req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Reaction added successfully",
		"data": fiber.Map{
			"comment_id": commentID,
		},
	})
}

func (h *Handler) CreateMention(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	var req model.CreateMentionRequest
	c.BodyParser(&req)
	req.CommentID = commentID

	id, err := h.commentSvc.AddMention(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Mention created successfully",
		"data": fiber.Map{
			"mention_id": id,
		},
	})
}
