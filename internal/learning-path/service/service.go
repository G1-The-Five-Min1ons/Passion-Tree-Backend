package service

import (
	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
)

type ServiceLearningPath interface {
	GetPaths() ([]model.LearningPath, error)
	GetPathDetails(id string) (*model.LearningPath, error)
	CreatePath(req model.CreatePathRequest) (string, error)
	UpdatePath(id string, req model.UpdatePathRequest) error
	DeletePath(id string) error
	StartPath(pathID string, userID string) error
	GetEnrollmentStatus(pathID string, userID string) (*model.PathEnroll, error)
}

type ServiceNode interface {
	AddNode(req model.CreateNodeRequest) (string, error)
	EditNode(nodeID string, req model.UpdateNodeRequest) error
	RemoveNode(nodeID string) error
	AddMaterial(req model.CreateMaterialRequest) (string, error)
	RemoveMaterial(materialID string) error
}

type ServiceComment interface {
	AddComment(req model.CreateCommentRequest) (string, error)
	GetNodeComments(nodeID string) ([]model.NodeComment, error)
	RemoveComment(commentID string) error
	AddReaction(req model.CreateReactionRequest) error
	AddMention(req model.CreateMentionRequest) (string, error)
}

type ServiceQuiz interface {
	AddQuestion(req model.CreateQuestionRequest) (string, error)
	GetQuestions(nodeID string) ([]model.NodeQuestion, error)
	RemoveQuestion(questionID string) error
	AddChoice(req model.CreateChoiceRequest) (string, error)
	RemoveChoice(choiceID string) error
}

type Service interface {
	ServiceLearningPath
	ServiceNode
	ServiceComment
	ServiceQuiz
}

type service struct {
	pathRepo    repository.RepositoryLearningPath
	nodeRepo    repository.RepositoryNode
	commentRepo repository.RepositoryComment
	quizRepo    repository.RepositoryQuiz
}

func NewService(repo repository.Repository) Service {
	return &service{
		pathRepo:    repo,
		nodeRepo:    repo,
		commentRepo: repo,
		quizRepo:    repo,
	}
}