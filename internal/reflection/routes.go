package reflection

import (
	"log/slog"
	"github.com/gofiber/fiber/v2"
	
	"passiontree/internal/reflection/handler"
	"passiontree/internal/reflection/repository"
	"passiontree/internal/reflection/service"
	"passiontree/internal/connection"
	"passiontree/internal/platform/aiclient"
)

func RegisterRoutes(r fiber.Router, db connection.Database, aiClient *aiclient.AIClient, logger *slog.Logger) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, aiClient, logger)
	h := handler.NewHandler(svc, logger)

	reflections := r.Group("/reflections")
	{
		reflections.Get("", h.GetAll)
		reflections.Post("", h.Create)
		reflections.Get("/:reflect_id", h.GetByID)
		reflections.Put("/:reflect_id", h.Update)
		reflections.Delete("/:reflect_id", h.Delete)
	}
	
	// Album routes
	albums := r.Group("/albums")
	{
		albums.Post("", h.CreateAlbum)
		albums.Get("", h.GetAlbumsByUserID)  // ?user_id=xxx
		albums.Get("/:album_id", h.GetAlbumByID)
		albums.Put("/:album_id", h.UpdateAlbum)
		albums.Delete("/:album_id", h.DeleteAlbum)
	}
	
	// Tree routes
	trees := r.Group("/trees")
	{
		trees.Post("", h.CreateTree)
		trees.Get("", h.GetTreesByAlbumID)  // ?album_id=xxx
		trees.Get("/:tree_id", h.GetTreeByID)
		trees.Put("/:tree_id", h.UpdateTree)
		trees.Delete("/:tree_id", h.DeleteTree)
	}
}
