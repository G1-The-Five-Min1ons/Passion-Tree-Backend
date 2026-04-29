package handler

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"
)

// GetTeacherVerificationStatus godoc
// @Summary      Get the authenticated user's teacher application status
// @Description  Returns pending/approved/rejected state plus review metadata for the user's teacher application.
// @Tags         Teacher
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      404  {object}  apidoc.ErrorResponse  "No application submitted yet"
// @Router       /auth/teacher/verification-status [get]
func (h *Handler) GetTeacherVerificationStatus(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	status, err := h.userSvc.GetTeacherVerificationStatus(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Teacher verification status retrieved successfully",
		"data":    status,
	})
}

// ApplyForTeacher godoc
// @Summary      Apply for teacher role
// @Description  Submits a teacher-role application with phone, reason, and teaching history. Requires authentication.
// @Tags         Teacher
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.ApplyTeacherRequest  true  "Application payload"
// @Success      201   {object}  apidoc.MessageResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Failure      409   {object}  apidoc.ErrorResponse  "Application already exists"
// @Router       /auth/teacher/apply [post]
func (h *Handler) ApplyForTeacher(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req model.ApplyTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	req.Reason = strings.TrimSpace(req.Reason)
	req.TeachingHistory = strings.TrimSpace(req.TeachingHistory)

	if req.PhoneNumber == "" {
		return h.handleError(c, apperror.NewBadRequest("phone_number is required"))
	}
	if req.Reason == "" {
		return h.handleError(c, apperror.NewBadRequest("reason is required"))
	}
	if req.TeachingHistory == "" {
		return h.handleError(c, apperror.NewBadRequest("teaching_history is required"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.ApplyForTeacher(ctx, userID, req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "teacher application submitted successfully", "user_id", userID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Teacher application submitted successfully",
	})
}
