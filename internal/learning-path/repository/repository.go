package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/connection"
	"passiontree/internal/learning-path/model"
)

type RepositoryLearningPath interface {
	GetAllLearningPath(ctx context.Context) ([]model.LearningPaths, error)
	GetLearningPathByID(ctx context.Context, path_id string) (*model.LearningPath, error)
	CreateLearningPath(ctx context.Context, req model.CreatePathRequest) (string, error)
	UpdateLearningPath(ctx context.Context, path_id string, req model.UpdatePathRequest) error
	DeleteLearningPath(ctx context.Context, path_id string) error
	EnrollLearningPathUser(ctx context.Context, pathID string, userID string) error
	GetLearningPathEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error)
	GetUserPathProgress(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error)
	UpdateLearningPathImage(ctx context.Context, pathID string, coverImgURL string) error
}

type RepositoryNode interface {
	GetNodeByID(ctx context.Context, nodeID string) (*model.Node, error)
	CreateNode(ctx context.Context, req model.CreateNodeRequest) (string, error)
	GetNodesByPathID(ctx context.Context, pathID string) ([]model.Node, error)
	UpdateNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error
	DeleteNode(ctx context.Context, nodeID string) error
	CreateMaterial(ctx context.Context, req model.CreateMaterialRequest) (string, error)
	GetMaterialsByNodeID(ctx context.Context, nodeID string) ([]model.NodeMaterial, error)
	DeleteMaterial(ctx context.Context, materialID string) error
	UpdateNodeSequence(ctx context.Context, nodeIDs []string) error
	CreateNodeWithContent(ctx context.Context, req model.CreateNodeRequest) (string, error)
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

type RepositoryHistory interface {
	GetHistoryByUserID(ctx context.Context, userID string) ([]model.HistoryResponse, error)
}

type RepositoryResume interface {
	GetNextNodeID(ctx context.Context, userID string, pathID string) (string, error)
}

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Database interface {
	DBTX
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
}

type Repository interface {
	RepositoryLearningPath
	RepositoryNode
	RepositoryComment
	RepositoryQuiz
	RepositoryHistory
	RepositoryResume
}

type repositoryImpl struct {
	db Database
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
