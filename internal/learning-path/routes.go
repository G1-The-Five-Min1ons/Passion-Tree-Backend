package learningpath

import (
	"log/slog"
	"github.com/gofiber/fiber/v2"

	"passiontree/internal/database"
	"passiontree/internal/learning-path/handler"
	"passiontree/internal/learning-path/repository"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/platform/aiclient"
)

func RegisterRoutes(r fiber.Router, db database.Database, aiClient *aiclient.AIClient, logger *slog.Logger, storageClient *database.StorageClient) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, aiClient, logger)
	h := handler.NewHandler(svc, logger, storageClient)

	paths := r.Group("/learningpaths")
	{
		paths.Get("", h.GetAll)
		paths.Post("", h.Create)
		paths.Post("/uploadimg", h.GetUploadURL)
		paths.Post("/search", h.Search)
		paths.Get("/debug/collection/:collection_name", h.DebugCollection)
		paths.Post("/sync/:path_id", h.SyncLearningPath)
		paths.Get("/:path_id", h.GetOne)
		paths.Put("/:path_id", h.Update)
		paths.Delete("/:path_id", h.Delete)
		paths.Post("/:path_id/start", h.Start)
		paths.Post("/:path_id/nodes", h.CreateNode)
		paths.Post("/generate", h.Generate)
		paths.Put("/:path_id/nodes/reorder", h.ReorderNodes)
	}

	nodes := r.Group("/learningpaths/nodes")
	{
		nodes.Get("/:node_id", h.GetOneNode)
		nodes.Put("/:node_id", h.UpdateNode)
		nodes.Delete("/:node_id", h.DeleteNode)
		nodes.Post("/:node_id/materials", h.CreateMaterial)
		nodes.Get("/:node_id/comments", h.GetComments)
		nodes.Post("/:node_id/comments", h.CreateComment)
		nodes.Get("/:node_id/questions", h.GetQuestions)
		nodes.Post("/:node_id/questions", h.CreateQuestion)
		nodes.Delete("/materials/:material_id", h.DeleteMaterial)
	}

	questions := r.Group("/learningpaths/questions")
	{
		questions.Delete("/:question_id", h.DeleteQuestion)
		questions.Post("/:question_id/choices", h.CreateChoice)
		questions.Delete("/choices/:choice_id", h.DeleteChoice)
	}

	userPaths := r.Group("/user/learningpaths")
	{
		userPaths.Get("/:path_id/status", h.GetEnrollmentStatus)
		userPaths.Get("/:path_id/progress", h.GetPathProgress)
	}

	comments := r.Group("/learningpaths/comments")
	{
		comments.Post("/:comment_id/mentions", h.CreateMention)
		comments.Post("/:comment_id/reactions", h.CreateReaction)
		comments.Delete("/:comment_id", h.DeleteComment)
	}
}
