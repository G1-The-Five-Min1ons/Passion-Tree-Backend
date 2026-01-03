package handler

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/pkg/apperror"
)
type Handler struct {
	pathSvc    service.ServiceLearningPath
	nodeSvc    service.ServiceNode
	commentSvc service.ServiceComment
	quizSvc    service.ServiceQuiz
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		pathSvc:    svc,
		nodeSvc:    svc,
		commentSvc: svc,
		quizSvc:    svc,
	}
}

func handleError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperror.AppError); ok {
		if appErr.Log != nil {
			log.Printf("[APP ERROR] Code: %d, Msg: %s, Cause: %v", appErr.Code, appErr.Message, appErr.Log)
		}
		return c.Status(appErr.Code).JSON(fiber.Map{
			"error": appErr.Message,
		})
	}

	log.Printf("[UNKNOWN ERROR] %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "internal server error",
	})
}