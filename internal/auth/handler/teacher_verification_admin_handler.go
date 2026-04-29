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

// ListTeacherApplications godoc
// @Summary      List teacher applications (admin)
// @Description  Returns teacher applications filtered by status. Defaults to pending.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        status  query     string  false  "Filter by status (pending, approved, rejected)"  Enums(pending, approved, rejected)
// @Success      200     {object}  apidoc.SuccessResponse
// @Failure      400     {object}  apidoc.ErrorResponse
// @Failure      401     {object}  apidoc.ErrorResponse
// @Failure      403     {object}  apidoc.ErrorResponse
// @Router       /auth/admin/teacher-applications [get]
func (h *Handler) ListTeacherApplications(c *fiber.Ctx) error {
	// Supported status values for filtering: pending (default), approved, rejected
	status := strings.ToLower(strings.TrimSpace(c.Query("status", model.TeacherApplicationStatusPending)))
	if status != model.TeacherApplicationStatusPending &&
		status != model.TeacherApplicationStatusApproved &&
		status != model.TeacherApplicationStatusRejected {
		return h.handleError(c, apperror.NewBadRequest("status must be one of: pending, approved, rejected"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	applications, err := h.userSvc.GetTeacherApplications(ctx, status)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Teacher applications retrieved successfully",
		"data":    applications,
	})
}

// ReviewTeacherApplication godoc
// @Summary      Review a teacher application (admin)
// @Description  Approves or rejects the specified teacher application.
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request_id  path      string                                 true  "Teacher application ID"
// @Param        body        body      model.ReviewTeacherApplicationRequest  true  "Decision payload"
// @Success      200         {object}  apidoc.MessageResponse
// @Failure      400         {object}  apidoc.ErrorResponse
// @Failure      401         {object}  apidoc.ErrorResponse
// @Failure      403         {object}  apidoc.ErrorResponse
// @Failure      404         {object}  apidoc.ErrorResponse
// @Router       /auth/admin/teacher-applications/{request_id} [put]
func (h *Handler) ReviewTeacherApplication(c *fiber.Ctx) error {
	adminID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	requestID := c.Params("request_id")
	if requestID == "" {
		return h.handleError(c, apperror.NewBadRequest("request_id is required"))
	}

	var req model.ReviewTeacherApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.Status != model.TeacherApplicationStatusApproved && req.Status != model.TeacherApplicationStatusRejected {
		return h.handleError(c, apperror.NewBadRequest("status must be either 'approved' or 'rejected'"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.ReviewTeacherApplication(ctx, requestID, adminID, req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.Info("admin reviewed teacher application",
		"admin_id", adminID,
		"request_id", requestID,
		"action", req.Status,
	)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Teacher application reviewed successfully",
	})
}
