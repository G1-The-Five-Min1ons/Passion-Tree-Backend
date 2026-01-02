package handler

import (
    "github.com/gofiber/fiber/v2"
    "passiontree/internal/learning-path/model"
)

func (h *Handler) GetComments(c *fiber.Ctx) error {
	comments, err := h.svc.GetNodeComments(c.Params("node_id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": comments})
}

func (h *Handler) CreateComment(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req model.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.NodeID = nodeID

	id, err := h.svc.AddComment(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"comment_id": id})
}

func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	if err := h.svc.RemoveComment(c.Params("comment_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "comment deleted"})
}

func (h *Handler) CreateReaction(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	var req model.CreateReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.CommentID = commentID

	if err := h.svc.AddReaction(req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "reaction added"})
}

func (h *Handler) CreateMention(c *fiber.Ctx) error {
	commentID := c.Params("comment_id")
	var req model.CreateMentionRequest
	c.BodyParser(&req)
	req.CommentID = commentID

	id, err := h.svc.AddMention(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"mention_id": id})
}