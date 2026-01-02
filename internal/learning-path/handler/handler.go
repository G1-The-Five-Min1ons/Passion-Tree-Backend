package handler

import "passiontree/internal/learning-path/service"

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