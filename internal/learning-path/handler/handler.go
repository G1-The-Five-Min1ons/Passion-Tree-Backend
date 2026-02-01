package handler

import (
	"log/slog"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/database"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	pathSvc    service.ServiceLearningPath
	searchSvc  service.ServiceSearch
	nodeSvc    service.ServiceNode
	commentSvc service.ServiceComment
	quizSvc    service.ServiceQuiz
	logger     *slog.Logger
	storage    *database.StorageClient
}

func NewHandler(svc service.Service, logger *slog.Logger, storage *database.StorageClient) *Handler {
	return &Handler{
		pathSvc:    svc,
		searchSvc:  svc,
		nodeSvc:    svc,
		commentSvc: svc,
		quizSvc:    svc,
		logger:     logger,
		storage:    storage,
	}
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	ctx := c.UserContext()

	// Common Attributes
	logAttrs := []any{
		"method", c.Method(),
		"path", c.Path(),
		"ip", c.IP(),
		"user_agent", c.Get("User-Agent"),
		"request_id", c.GetRespHeader("X-Request-ID"), // Request ID Middleware
	}

	// Business/Handled Error)
	if appErr, ok := err.(*apperror.AppError); ok {
		if appErr.Log != nil {
			h.logger.WarnContext(ctx, "application handled error",
				append(logAttrs,
					"code", appErr.Code,
					"message", appErr.Message,
					"cause", appErr.Log,
				)...,
			)
		}

		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"error":   appErr.Message,
		})
	}

	// System Error (Unhandled Error)
	h.logger.ErrorContext(ctx, "unhandled system error",
		append(logAttrs, "error", err.Error())...,
	)

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "internal server error",
	})
}
