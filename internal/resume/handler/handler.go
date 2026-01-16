package handler

import (
	"log"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/resume/service"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc service.ResumeService
}

func NewHandler(svc service.ResumeService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) GetResume(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	pathID := c.Query("path_id")

	if userID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id is required"))
	}
	if pathID == "" {
		return h.handleError(c, apperror.NewBadRequest("path_id is required"))
	}

	ctx := c.UserContext()
	
	resp, err := h.svc.GetResumeNode(ctx, userID, pathID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "GetResume successfully",
		"data":    resp,
	})
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperror.AppError); ok {
		if appErr.Log != nil {
			log.Printf("[APP ERROR] Code: %d, Msg: %s, Cause: %v", appErr.Code, appErr.Message, appErr.Log)
		}
		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"error": appErr.Message,
		})
	}

	log.Printf("[UNKNOWN ERROR] %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error": "internal server error",
	})
}