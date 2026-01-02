package handler

import (
    "github.com/gofiber/fiber/v2"
    "passiontree/internal/learning-path/model"
)

func (h *Handler) GetQuestions(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	questions, err := h.svc.GetQuestions(nodeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": questions})
}

func (h *Handler) CreateQuestion(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req model.CreateQuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.NodeID = nodeID

	id, err := h.svc.AddQuestion(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"question_id": id})
}

func (h *Handler) DeleteQuestion(c *fiber.Ctx) error {
	if err := h.svc.RemoveQuestion(c.Params("question_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "question deleted"})
}

func (h *Handler) CreateChoice(c *fiber.Ctx) error {
	questionID := c.Params("question_id")
	var req model.CreateChoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.QuestionID = questionID

	id, err := h.svc.AddChoice(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"choice_id": id})
}

func (h *Handler) DeleteChoice(c *fiber.Ctx) error {
	if err := h.svc.RemoveChoice(c.Params("choice_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "choice deleted"})
}