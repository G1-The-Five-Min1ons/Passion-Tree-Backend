package resume

import (
	"github.com/gofiber/fiber/v2"
	"passiontree/internal/database"
	
	nodeRepoPkg "passiontree/internal/learning-path/repository" 
	resumeRepo "passiontree/internal/resume/repository"
	"passiontree/internal/resume/handler"
	"passiontree/internal/resume/service"
)

func RegisterRoutes(r fiber.Router, db database.Database) {
	rRepo := resumeRepo.NewRepository(db)
	nRepo := nodeRepoPkg.NewRepository(db) 
	svc := service.NewService(rRepo, nRepo)
	h := handler.NewHandler(svc)

	resumeGroup := r.Group("/resume")
	{
		resumeGroup.Get("", h.GetResume)
	}
}