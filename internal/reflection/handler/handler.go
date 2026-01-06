package handler

import (
    "log"

    "github.com/gofiber/fiber/v2"
    "passiontree/internal/pkg/apperror"
    "passiontree/internal/reflection/service"
)

type Handler struct {
    service service.reflectSvc
}

func NewHandler(s service.ReflectionService) *Handler {
    return &Handler{service: s}
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
    if appErr, ok := err.(apperror.AppError); ok {
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

