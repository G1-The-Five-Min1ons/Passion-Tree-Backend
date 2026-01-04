package learningpath

import (
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/learning-path/handler"
	"passiontree/internal/learning-path/repository"
	"passiontree/internal/learning-path/service"
	"passiontree/internal/database"
)

func RegisterRoutes(r fiber.Router, db database.Database) {
    repo := repository.NewRepository(db)
    svc := service.NewService(repo)
    h := handler.NewHandler(svc)

	paths := r.Group("/learningpaths")
	{
		paths.Get("", h.GetAll)
		paths.Post("", h.Create)
		paths.Get("/:path_id", h.GetOne)
		paths.Put("/:path_id", h.Update)
		paths.Delete("/:path_id", h.Delete)
		paths.Post("/:path_id/start", h.Start)
		paths.Post("/:path_id/nodes", h.CreateNode)
	}

	nodes := r.Group("/learningpaths/nodes")
	{
		nodes.Put("/:node_id", h.UpdateNode)
		nodes.Delete("/:node_id", h.DeleteNode)
		nodes.Post("/:node_id/materials", h.CreateMaterial)
		nodes.Get("/:node_id/comments", h.GetComments)
		nodes.Post("/:node_id/comments", h.CreateComment)
		nodes.Get("/:node_id/questions", h.GetQuestions)
		nodes.Post("/:node_id/questions", h.CreateQuestion)
	}

	questions := r.Group("/learningpaths/questions")
	{
		questions.Delete("/:question_id", h.DeleteQuestion)
		questions.Post("/:question_id/choices", h.CreateChoice)
	}

	userPaths := r.Group("/user/learningpaths")
	{
		userPaths.Get("/:path_id/status", h.GetEnrollmentStatus)
	}

	r.Post("/learningpaths/comments/:comment_id/mentions", h.CreateMention)
	r.Post("/learningpaths/comments/:comment_id/reactions", h.CreateReaction)
	r.Delete("/learningpaths/comments/:comment_id", h.DeleteComment)
	r.Delete("/learningpaths/choices/:choice_id", h.DeleteChoice)
	r.Delete("/learningpaths/materials/:material_id", h.DeleteMaterial)
}