package handler

import (
	"context"
	"time"

	"passiontree/internal/mission/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// CreateTemplate godoc
// @Summary      Create a mission template (admin)
// @Tags         Missions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateTemplateRequest  true  "Mission template payload"
// @Success      201   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Failure      403   {object}  apidoc.ErrorResponse
// @Router       /admin/missions/templates [post]
func (h *Handler) CreateTemplate(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req model.CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	id, err := h.missionSvc.CreateTemplate(ctx, req, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Mission template created successfully",
		"data":    fiber.Map{"mission_id": id},
	})
}

// GetAllTemplates godoc
// @Summary      List all mission templates (admin)
// @Tags         Missions
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      403  {object}  apidoc.ErrorResponse
// @Router       /admin/missions/templates [get]
func (h *Handler) GetAllTemplates(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	templates, err := h.missionSvc.GetAllTemplates(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Retrieved all mission templates successfully",
		"data":    templates,
	})
}

// DeleteTemplate godoc
// @Summary      Delete a mission template (admin)
// @Tags         Missions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Mission template ID"
// @Success      200  {object}  apidoc.MessageResponse
// @Failure      400  {object}  apidoc.ErrorResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      403  {object}  apidoc.ErrorResponse
// @Failure      404  {object}  apidoc.ErrorResponse
// @Router       /admin/missions/templates/{id} [delete]
func (h *Handler) DeleteTemplate(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	missionID := c.Params("id")
	if missionID == "" {
		return h.handleError(c, apperror.NewBadRequest("mission id is required"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.missionSvc.DeleteTemplate(ctx, missionID, userID); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Mission template deleted successfully",
	})
}

// GetMyMissions godoc
// @Summary      List the authenticated user's active missions
// @Tags         Missions
// @Produce      json
// @Security     BearerAuth
// @Param        X-Client-Request-Id  header    string  false  "Optional client correlation ID echoed back in the response"
// @Success      200                  {object}  apidoc.SuccessResponse
// @Failure      401                  {object}  apidoc.ErrorResponse
// @Router       /user/missions [get]
func (h *Handler) GetMyMissions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}
	clientRequestID := c.Get("X-Client-Request-Id")

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	startedAt := time.Now()

	h.logger.InfoContext(ctx, "mission fetch request received",
		"user_id", userID,
		"client_request_id", clientRequestID,
		"method", c.Method(),
		"path", c.Path(),
		"ip", c.IP(),
		"request_id", c.GetRespHeader("X-Request-ID"),
	)

	missions, err := h.missionSvc.GetMyMissions(ctx, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "mission fetch failed",
			"user_id", userID,
			"client_request_id", clientRequestID,
			"elapsed_ms", time.Since(startedAt).Milliseconds(),
			"error", err,
		)
		return h.handleError(c, err)
	}
	if clientRequestID != "" {
		c.Set("X-Client-Request-Id", clientRequestID)
	}

	h.logger.InfoContext(ctx, "mission fetch response sent",
		"user_id", userID,
		"client_request_id", clientRequestID,
		"missions_count", len(missions),
		"status", fiber.StatusOK,
		"elapsed_ms", time.Since(startedAt).Milliseconds(),
	)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Retrieved active missions successfully",
		"data":    missions,
	})
}

// TriggerAutoAssign godoc
// @Summary      Manually trigger weekly mission auto-assignment for the current user
// @Description  Forces the weekly mission assignment routine to run immediately for the authenticated user.
// @Tags         Missions
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.MessageResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      500  {object}  apidoc.ErrorResponse
// @Router       /system/missions/auto-assign [post]
func (h *Handler) TriggerAutoAssign(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 60*time.Second)
	defer cancel()

	if err := h.missionSvc.AutoAssignWeeklyMissionsByUser(ctx, userID); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Weekly missions auto-assigned triggered successfully",
	})
}
