package repository

import (
	"database/sql"
	"passiontree/internal/database"
	"passiontree/internal/learning-path/model"
)

type RepositoryLearningPath interface {
	GetAllLearnningPath() ([]model.LearningPath, error)
	GetLearnningPathByID(id string) (*model.LearningPath, error)
	CreateLearnningPath(req model.CreatePathRequest) (string, error)
	UpdateLearnningPath(id string, req model.UpdatePathRequest) error
	DeleteLearnningPath(id string) error
	EnrollLearnningPathUser(pathID string, userID string) error
	GetLearnningPathEnrollmentStatus(pathID string, userID string) (*model.PathEnroll, error)
}

type RepositoryNode interface {
	CreateNode(req model.CreateNodeRequest) (string, error)
	GetNodesByPathID(pathID string) ([]model.Node, error)
	UpdateNode(nodeID string, req model.UpdateNodeRequest) error
	DeleteNode(nodeID string) error
	CreateMaterial(req model.CreateMaterialRequest) (string, error)
	GetMaterialsByNodeID(nodeID string) ([]model.NodeMaterial, error)
	DeleteMaterial(materialID string) error
}

type RepositoryComment interface {
	CreateComment(req model.CreateCommentRequest) (string, error)
	GetCommentsByNodeID(nodeID string) ([]model.NodeComment, error)
	DeleteComment(commentID string) error
	CreateReaction(req model.CreateReactionRequest) error
	GetReactionsByCommentID(commentID string) ([]model.CommentReaction, error)
	CreateMention(req model.CreateMentionRequest) (string, error)
}

type RepositoryQuiz interface {
	CreateQuestion(req model.CreateQuestionRequest) (string, error)
	GetQuestionsByNodeID(nodeID string) ([]model.NodeQuestion, error)
	DeleteQuestion(questionID string) error
	CreateChoice(req model.CreateChoiceRequest) (string, error)
	GetChoicesByQuestionID(questionID string) ([]model.QuestionChoice, error)
	DeleteChoice(choiceID string) error
}

type Repository interface {
	RepositoryLearningPath
	RepositoryNode
	RepositoryComment
	RepositoryQuiz
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(ds database.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}