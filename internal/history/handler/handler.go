package handler

import (
	"log"
	"passiontree/internal/history/service"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc service.ServiceHistory
}

func NewHandler(svc service.ServiceHistory) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperror.AppError); ok {
		if appErr.Log != nil {
			log.Printf("[APP ERROR] Code: %d, Msg: %s, Cause: %v", appErr.Code, appErr.Message, appErr.Log)
		}
		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"error":   appErr.Message,
		})
	}

	log.Printf("[UNKNOWN ERROR] %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "internal server error",
	})
}

func (h *Handler) GetUserHistory(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	historyList, err := h.svc.GetUserHistory(userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User history retrieved successfully",
		"data":    historyList,
	})
}