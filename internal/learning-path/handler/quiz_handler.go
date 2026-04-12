package handler

import (
	"context"
	"time"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetQuestions(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	questions, err := h.quizSvc.GetQuestions(ctx, nodeID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "retrieved questions successfully", "node_id", nodeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Questions retrieved successfully",
		"data":    questions,
	})
}

func (h *Handler) CreateQuestion(c *fiber.Ctx) error {
	nodeID := c.Params("node_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.CreateQuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	req.NodeID = nodeID

	id, err := h.quizSvc.AddQuestion(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "quiz question created successfully", "node_id", nodeID, "question_id", id)

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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.quizSvc.RemoveQuestion(ctx, question_id); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "quiz question deleted successfully", "question_id", question_id)

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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	var req model.CreateChoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}
	req.QuestionID = questionID

	id, err := h.quizSvc.AddChoice(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "choice added successfully", "question_id", questionID, "choice_id", id)

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
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.quizSvc.RemoveChoice(ctx, choice_id); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "choice deleted successfully", "choice_id", choice_id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Choice deleted successfully",
		"data": fiber.Map{
			"choice_id": choice_id,
		},
	})
}

func (h *Handler) UpdateQuestion(c *fiber.Ctx) error {
	questionID := c.Params("question_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.UpdateQuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	h.logger.InfoContext(ctx, "updating question", "question_id", questionID)

	if err := h.quizSvc.EditQuestion(ctx, questionID, req); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Question updated successfully",
		"data": fiber.Map{
			"question_id": questionID,
		},
	})
}

func (h *Handler) UpdateChoice(c *fiber.Ctx) error {
	choiceID := c.Params("choice_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.UpdateChoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	h.logger.InfoContext(ctx, "updating choice", "choice_id", choiceID)

	if err := h.quizSvc.EditChoice(ctx, choiceID, req); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Choice updated successfully",
		"data": fiber.Map{
			"choice_id": choiceID,
		},
	})
}
