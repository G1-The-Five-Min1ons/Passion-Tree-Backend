package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/database"
	"passiontree/internal/learning-path/model"
)

type RepositoryLearningPath interface {
	GetAllLearnningPath(ctx context.Context) ([]model.LearningPath, error)
	GetLearnningPathByID(ctx context.Context, id string) (*model.LearningPath, error)
	CreateLearnningPath(ctx context.Context, req model.CreatePathRequest) (string, error)
	UpdateLearnningPath(ctx context.Context, id string, req model.UpdatePathRequest) error
	DeleteLearnningPath(ctx context.Context, id string) error
	EnrollLearnningPathUser(ctx context.Context, pathID string, userID string) error
	GetLearnningPathEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error)
}

type RepositoryNode interface {
	CreateNode(ctx context.Context, req model.CreateNodeRequest) (string, error)
	GetNodesByPathID(ctx context.Context, pathID string) ([]model.Node, error)
	UpdateNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error
	DeleteNode(ctx context.Context, nodeID string) error
	CreateMaterial(ctx context.Context, req model.CreateMaterialRequest) (string, error)
	GetMaterialsByNodeID(ctx context.Context, nodeID string) ([]model.NodeMaterial, error)
	DeleteMaterial(ctx context.Context, materialID string) error
}

type RepositoryComment interface {
	CreateComment(ctx context.Context, req model.CreateCommentRequest) (string, error)
	GetCommentsByNodeID(ctx context.Context, nodeID string) ([]model.NodeComment, error)
	DeleteComment(ctx context.Context, commentID string) error
	CreateReaction(ctx context.Context, req model.CreateReactionRequest) error
	GetReactionsByCommentID(ctx context.Context, commentID string) ([]model.CommentReaction, error)
	CreateMention(ctx context.Context, req model.CreateMentionRequest) (string, error)
}

type RepositoryQuiz interface {
	CreateQuestion(ctx context.Context, req model.CreateQuestionRequest) (string, error)
	GetQuestionsByNodeID(ctx context.Context, nodeID string) ([]model.NodeQuestion, error)
	DeleteQuestion(ctx context.Context, questionID string) error
	CreateChoice(ctx context.Context, req model.CreateChoiceRequest) (string, error)
	GetChoicesByQuestionID(ctx context.Context, questionID string) ([]model.QuestionChoice, error)
	DeleteChoice(ctx context.Context, choiceID string) error
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
