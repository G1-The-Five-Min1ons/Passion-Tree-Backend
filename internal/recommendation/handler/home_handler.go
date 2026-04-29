package handler

import (
	"context"
	"time"

	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// GetHomeRecommendations godoc
// @Summary      Get home-screen recommendations for the authenticated user
// @Description  Returns the precomputed daily recommendation set for the home screen.
// @Tags         Recommendations
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Router       /home/recommendation [get]
func (h *Handler) GetHomeRecommendations(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Please log in again to continue",
		})
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "received request for home path recommendations", "user_id", userID)

	response, err := h.recSvc.RecommendHomePathsForUser(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Recommendations retrieved successfully",
		"data":    response,
	})
}

// TriggerBatchRecommendation godoc
// @Summary      Manually trigger the daily recommendation batch (admin)
// @Description  Starts an async job that recomputes home-screen recommendations for all users.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      202  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      403  {object}  apidoc.ErrorResponse
// @Failure      409  {object}  apidoc.ErrorResponse  "Batch already running"
// @Router       /batch/recommendation/trigger [post]
func (h *Handler) TriggerBatchRecommendation(c *fiber.Ctx) error {
	role, err := middleware.GetRoleFromContext(c)
	if err != nil || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "admin access required",
		})
	}

	const taskType = "recommendation_batch"

	// 2. 🚦 ด่านตรวจ: กันการรันซ้อน (Idempotency)
	task, ok := h.tasks.Create(taskType)
	if !ok {
		return h.handleError(c, apperror.NewConflict("Recommendation batch is already running"))
	}

	h.logger.InfoContext(c.UserContext(), "manual trigger for recommendation batch started", "task_id", task.TaskID)

	go func(taskID string) {
		// ใช้ Background Context เพราะงานอาจยาวกว่า 5 นาที และไม่ควรผูกกับ HTTP Request
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		err := h.recSvc.RunDailyRecommendationBatch(ctx)
		if err != nil {
			h.tasks.Fail(taskID, taskType, err.Error())
			h.logger.ErrorContext(ctx, "batch recommendation failed", "task_id", taskID, "error", err)
			return
		}

		h.tasks.Complete(taskID, taskType, "Recommendation batch finished successfully")
		h.logger.InfoContext(ctx, "batch recommendation completed", "task_id", taskID)
	}(task.TaskID)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": "Recommendation batch triggered in background",
		"data": fiber.Map{
			"task_id": task.TaskID,
			"status":  task.Status,
		},
	})
}
