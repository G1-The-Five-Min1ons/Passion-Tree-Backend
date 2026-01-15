package history

import (
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/database"
	"passiontree/internal/history/handler"
	"passiontree/internal/history/repository"
	"passiontree/internal/history/service"
)

func RegisterRoutes(r fiber.Router, db database.Database) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	historyGroup := r.Group("/history")
	{
		historyGroup.Get("", h.GetUserHistory)
	}
}