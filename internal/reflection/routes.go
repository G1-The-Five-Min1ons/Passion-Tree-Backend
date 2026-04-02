package reflection

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"passiontree/internal/connection"
	"passiontree/internal/pkg/jwt"
	"passiontree/internal/pkg/middleware"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/reflection/handler"
	"passiontree/internal/reflection/repository"
	missionRepo "passiontree/internal/mission/repository"
	repoUser "passiontree/internal/auth/repository"
	missionService "passiontree/internal/mission/service"
	"passiontree/internal/reflection/service"
)

func RegisterRoutes(r fiber.Router, db connection.Database, aiClient *aiclient.AIClient, jwtService *jwt.Service, logger *slog.Logger) {
	missionRepo := missionRepo.NewRepository(db)
	repouser := repoUser.NewRepository(db)
	missionSvc := missionService.NewService(missionRepo, repouser, logger)	

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, missionSvc, aiClient, logger)
	h := handler.NewHandler(svc, logger)

	// All reflection routes require JWT authentication
	protected := r.Group("/", middleware.JWTMiddleware(jwtService, logger))

	reflections := protected.Group("/reflections")
	{
		reflections.Get("", h.GetAll)
		reflections.Get("/stats", h.GetStats)
		reflections.Post("", h.Create)
		reflections.Get("/:reflect_id", h.GetByID)
		reflections.Put("/:reflect_id", h.Update)
		reflections.Delete("/:reflect_id", h.Delete)
	}

	// Album routes
	albums := protected.Group("/albums")
	{
		albums.Post("", h.CreateAlbum)
		albums.Get("", h.GetAlbumsByUserID) // ?user_id=xxx
		albums.Get("/:album_id", h.GetAlbumByID)
		albums.Put("/:album_id", h.UpdateAlbum)
		albums.Delete("/:album_id", h.DeleteAlbum)
	}

	// Tree routes
	trees := protected.Group("/trees")
	{
		trees.Post("", h.CreateTree)
		trees.Get("", h.GetTreesByAlbumID) // ?album_id=xxx
		trees.Get("/:tree_id", h.GetTreeByID)
		trees.Put("/:tree_id", h.UpdateTree)
		trees.Patch("/:tree_id/pause", h.PauseTree)
		trees.Patch("/:tree_id/score", h.CalculateTreeScore)
		trees.Delete("/:tree_id", h.DeleteTree)
	}

	// Tree Node routes
	treeNodes := protected.Group("/tree-nodes")
	{
		treeNodes.Post("", h.CreateTreeNode)
		treeNodes.Get("", h.GetTreeNodesByTreeID) // ?tree_id=xxx
		treeNodes.Get("/:tree_node_id", h.GetTreeNodeByID)
		treeNodes.Put("/:tree_node_id", h.UpdateTreeNode)
		treeNodes.Delete("/:tree_node_id", h.DeleteTreeNode)
	}
}
