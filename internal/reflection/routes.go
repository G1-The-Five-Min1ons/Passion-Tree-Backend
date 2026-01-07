package reflection

import (
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/reflection/handler"
	"passiontree/internal/reflection/repository"
	"passiontree/internal/reflection/service"
	"passiontree/internal/database"
)

func RegisterRoutes(r fiber.Router, db database.Database) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	reflections := r.Group("/reflections")
	{
		reflections.Get("", h.GetAll)
		reflections.Post("", h.Create)
		reflections.Get("/:reflect_id", h.GetByID)
		reflections.Put("/:reflect_id", h.Update)
		reflections.Delete("/:reflect_id", h.Delete)
	}
}
