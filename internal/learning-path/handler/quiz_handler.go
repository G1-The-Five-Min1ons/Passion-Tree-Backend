package handler

import (
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (h *Handler) GetQuestions(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	questions, err := h.quizSvc.GetQuestions(nodeID)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": questions})
}

func (h *Handler) CreateQuestion(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	var req model.CreateQuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	req.NodeID = nodeID

	id, err := h.quizSvc.AddQuestion(req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Question has been created and added to node successfully",
		"data": fiber.Map{
			"question_id": id,
		},
	})
}

func (h *Handler) DeleteQuestion(c *fiber.Ctx) error {
	question_id := c.Params("question_id")
	if err := h.quizSvc.RemoveQuestion(question_id); err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Question has been deleted successfully",
		"data": fiber.Map{
			"question_id": question_id,
		},
	})
}

func (h *Handler) CreateChoice(c *fiber.Ctx) error {
	questionID := c.Params("question_id")
	var req model.CreateChoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	req.QuestionID = questionID

	id, err := h.quizSvc.AddChoice(req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Choice added to question successfully",
		"data": fiber.Map{
			"choice_id": id,
		},
	})
}

func (h *Handler) DeleteChoice(c *fiber.Ctx) error {
	choice_id := c.Params("choice_id")
	if err := h.quizSvc.RemoveChoice(choice_id); err != nil {
		return h.handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Choice deleted successfully",
		"data": fiber.Map{
			"choice_id": choice_id,
		},
	})
}