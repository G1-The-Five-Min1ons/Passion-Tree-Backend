package handler

import (
	"context"
	"time"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

// GetQuestions godoc
// @Summary      List quiz questions for a node
// @Tags         Quiz
// @Produce      json
// @Security     BearerAuth
// @Param        node_id  path      string  true  "Node ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/nodes/{node_id}/questions [get]
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

// CreateQuestion godoc
// @Summary      Add a quiz question to a node
// @Tags         Quiz
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        node_id  path      string                       true  "Node ID"
// @Param        body     body      model.CreateQuestionRequest  true  "Question payload"
// @Success      201      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/nodes/{node_id}/questions [post]
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

// DeleteQuestion godoc
// @Summary      Delete a quiz question
// @Tags         Quiz
// @Produce      json
// @Security     BearerAuth
// @Param        question_id  path      string  true  "Question ID"
// @Success      200          {object}  apidoc.SuccessResponse
// @Failure      401          {object}  apidoc.ErrorResponse
// @Failure      404          {object}  apidoc.ErrorResponse
// @Router       /learningpaths/questions/{question_id} [delete]
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

// CreateChoice godoc
// @Summary      Add a choice to a quiz question
// @Tags         Quiz
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        question_id  path      string                     true  "Question ID"
// @Param        body         body      model.CreateChoiceRequest  true  "Choice payload"
// @Success      201          {object}  apidoc.SuccessResponse
// @Failure      400          {object}  apidoc.ErrorResponse
// @Failure      401          {object}  apidoc.ErrorResponse
// @Router       /learningpaths/questions/{question_id}/choices [post]
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

// DeleteChoice godoc
// @Summary      Delete a quiz choice
// @Tags         Quiz
// @Produce      json
// @Security     BearerAuth
// @Param        choice_id  path      string  true  "Choice ID"
// @Success      200        {object}  apidoc.SuccessResponse
// @Failure      401        {object}  apidoc.ErrorResponse
// @Failure      404        {object}  apidoc.ErrorResponse
// @Router       /learningpaths/questions/choices/{choice_id} [delete]
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

// UpdateQuestion godoc
// @Summary      Update a quiz question
// @Tags         Quiz
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        question_id  path      string                       true  "Question ID"
// @Param        body         body      model.UpdateQuestionRequest  true  "Updated fields"
// @Success      200          {object}  apidoc.SuccessResponse
// @Failure      400          {object}  apidoc.ErrorResponse
// @Failure      401          {object}  apidoc.ErrorResponse
// @Failure      404          {object}  apidoc.ErrorResponse
// @Router       /learningpaths/questions/{question_id} [put]
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

// UpdateChoice godoc
// @Summary      Update a quiz choice
// @Tags         Quiz
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        choice_id  path      string                     true  "Choice ID"
// @Param        body       body      model.UpdateChoiceRequest  true  "Updated fields"
// @Success      200        {object}  apidoc.SuccessResponse
// @Failure      400        {object}  apidoc.ErrorResponse
// @Failure      401        {object}  apidoc.ErrorResponse
// @Failure      404        {object}  apidoc.ErrorResponse
// @Router       /learningpaths/questions/choices/{choice_id} [put]
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
