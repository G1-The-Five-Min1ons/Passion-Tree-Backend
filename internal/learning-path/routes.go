package learningpath

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"passiontree/internal/connection"
	"passiontree/internal/learning-path/handler"
	"passiontree/internal/learning-path/repository"
	missionRepo "passiontree/internal/mission/repository"
	repoUser "passiontree/internal/auth/repository"
	missionService "passiontree/internal/mission/service"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/pkg/jwt"
	"passiontree/internal/pkg/middleware"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/platform/aiclient"
)

func RegisterRoutes(r fiber.Router, db connection.Database, aiClient *aiclient.AIClient, jwtService *jwt.Service, logger *slog.Logger, storageClient *storage.BlobService) {
	missionRepo := missionRepo.NewRepository(db)
	repouser := repoUser.NewRepository(db)
	missionSvc := missionService.NewService(missionRepo, repouser, logger)

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, missionSvc, aiClient, logger)
	h := handler.NewHandler(svc, logger, storageClient)

	// All reflection routes require JWT authentication
	// protected := r.Group("/") //for dev
	protected := r.Group("/", middleware.JWTMiddleware(jwtService, logger))

	paths := protected.Group("/learningpaths")
	{
		paths.Get("", h.GetAll)
		paths.Post("", h.Create)
		paths.Get("/user/enroll", h.GetMyPaths)
		paths.Put("/uploadimg", h.UpdateCoverImage)
		paths.Post("/search", h.Search)
		paths.Get("/debug/collection/:collection_name", h.DebugCollection)
		paths.Post("/sync/:path_id", h.SyncLearningPath)
		paths.Get("/:path_id", h.GetOne)
		paths.Put("/:path_id", h.Update)
		paths.Delete("/:path_id", h.Delete)
		paths.Post("/:path_id/start", h.Start)
		paths.Get("/:path_id/nodes", h.GetNodesByPathID)
		paths.Post("/:path_id/nodes", h.CreateNode)
		paths.Post("/generate", h.Generate)
		paths.Put("/:path_id/nodes/reorder", h.ReorderNodes)
		paths.Get("/:path_id/comments", h.GetPathComments)
		paths.Post("/:path_id/comments", h.CreatePathComment)
		paths.Post("/:path_id/ratings", h.SubmitRating)
		paths.Get("/:path_id/ratings", h.GetMyRating)
		paths.Delete("/:path_id/ratings", h.DeleteRating)
	}

	nodes := protected.Group("/learningpaths/nodes")
	{
		nodes.Get("/:node_id", h.GetOneNode)
		nodes.Put("/:node_id", h.UpdateNode)
		nodes.Delete("/:node_id", h.DeleteNode)
		nodes.Put("/:node_id/start", h.StartNodeStatus)
		nodes.Put("/:node_id/complete", h.CompleteNodeStatus)
		nodes.Post("/:node_id/materials", h.CreateMaterial)
		nodes.Get("/:node_id/comments", h.GetComments)
		nodes.Post("/:node_id/comments", h.CreateComment)
		nodes.Get("/:node_id/questions", h.GetQuestions)
		nodes.Post("/:node_id/questions", h.CreateQuestion)
		nodes.Delete("/materials/:material_id", h.DeleteMaterial)
	}

	questions := protected.Group("/learningpaths/questions")
	{
		questions.Put("/:question_id", h.UpdateQuestion)
		questions.Delete("/:question_id", h.DeleteQuestion)
		questions.Post("/:question_id/choices", h.CreateChoice)
		questions.Put("/choices/:choice_id", h.UpdateChoice)
		questions.Delete("/choices/:choice_id", h.DeleteChoice)
	}

	userPaths := protected.Group("/user/learningpaths")
	{
		userPaths.Get("/:path_id/status", h.GetEnrollmentStatus)
		userPaths.Get("/:path_id/progress", h.GetPathProgress)
	}

	historyGroup := protected.Group("/user/learningpaths/history")
	{
		historyGroup.Get("", h.GetUserHistory)
	}

	resumeGroup := protected.Group("/user/learningpaths/resume")
	{
		resumeGroup.Get("", h.GetResume)
	}

	comments := protected.Group("/learningpaths/comments")
	{
		comments.Put("/:comment_id", h.UpdateComment)
		comments.Post("/:comment_id/reactions", h.CreateReaction)
		comments.Delete("/:comment_id", h.DeleteComment)
		// Manual mention — triggered when user types @ to tag someone explicitly
		comments.Post("/:comment_id/mentions", h.CreateMention)
	}
}
