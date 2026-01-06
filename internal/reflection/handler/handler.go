package handler

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "passiontree/internal/reflection/service"
    "passiontree/internal/pkg/apperror"
)

type Handler struct {
    reflectSvc service.svc
}

func NewHandler(svc service.svc) *Handler {
    return &Handler{
        reflectSvc: svc,
    }
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
