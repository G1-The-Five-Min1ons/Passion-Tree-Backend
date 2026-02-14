package handler

import (
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"

	"passiontree/internal/upload/model"
)

func (h *Handler) GetPresignedIMGURL(c *fiber.Ctx) error {
	var req model.UploadImageRequest

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	result, err := h.service.GeneratePresignedURL(c.UserContext(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(c.UserContext(), "presigned URL request completed successfully")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Presigned URL generated successfully",
		"data":    result,
	})
}
