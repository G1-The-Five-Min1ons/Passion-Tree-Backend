package handler

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/pkg/apperror"
)

type Handler struct {
	pathSvc    service.ServiceLearningPath
	searchSvc  service.ServiceSearch
	nodeSvc    service.ServiceNode
	commentSvc service.ServiceComment
	quizSvc    service.ServiceQuiz
	logger     *slog.Logger
}

func NewHandler(svc service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		pathSvc:    svc,
		searchSvc:  svc,
		nodeSvc:    svc,
		commentSvc: svc,
		quizSvc:    svc,
		logger:     logger,
	}
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	ctx := c.UserContext()

	if appErr, ok := err.(*apperror.AppError); ok {
		if appErr.Log != nil {
			h.logger.WarnContext(ctx, "application handled error",
				"code", appErr.Code,
				"message", appErr.Message,
				"cause", appErr.Log,
				"path", c.Path(),
			)
		}

		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"error":   appErr.Message,
		})
	}

	h.logger.ErrorContext(ctx, "unhandled system error",
		"error", err.Error(),
		"method", c.Method(),
		"path", c.Path(),
	)

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "internal server error",
	})
}
