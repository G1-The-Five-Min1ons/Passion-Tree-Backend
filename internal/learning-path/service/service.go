package service

import (
	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
)

type Service interface {
	GetPaths() ([]model.LearningPath, error)
	GetPathDetails(id string) (*model.LearningPath, error)
	CreatePath(req model.CreatePathRequest) (string, error)
	UpdatePath(id string, req model.UpdatePathRequest) error
	DeletePath(id string) error
	StartPath(pathID string, userID string) error
	GetEnrollmentStatus(pathID string, userID string) (*model.PathEnroll, error)

	AddNode(req model.CreateNodeRequest) (string, error)
	EditNode(nodeID string, req model.UpdateNodeRequest) error
	RemoveNode(nodeID string) error
	AddMaterial(req model.CreateMaterialRequest) (string, error)
	RemoveMaterial(materialID string) error

	AddComment(req model.CreateCommentRequest) (string, error)
	GetNodeComments(nodeID string) ([]model.NodeComment, error)
	RemoveComment(commentID string) error
	AddReaction(req model.CreateReactionRequest) error
	AddMention(req model.CreateMentionRequest) (string, error)

	AddQuestion(req model.CreateQuestionRequest) (string, error)
	GetQuestions(nodeID string) ([]model.NodeQuestion, error)
	RemoveQuestion(questionID string) error
	AddChoice(req model.CreateChoiceRequest) (string, error)
	RemoveChoice(choiceID string) error
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}