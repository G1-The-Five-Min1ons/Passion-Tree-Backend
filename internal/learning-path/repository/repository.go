package repository

import (
	"context"
	"database/sql"
	
	"passiontree/internal/connection"
	"passiontree/internal/learning-path/model"
)

type RepositoryLearningPath interface {
	GetAllLearningPath(ctx context.Context) ([]model.LearningPath, error)
	GetLearningPathByID(ctx context.Context, path_id string) (*model.LearningPath, error)
	GetLearningPathsByIDs(ctx context.Context, pathIDs []string) ([]model.LearningPath, error)
	CreateLearningPath(ctx context.Context, req model.CreatePathRequest) (string, error)
	UpdateLearningPath(ctx context.Context, path_id string, req model.UpdatePathRequest) error
	DeleteLearningPath(ctx context.Context, path_id string) error
	EnrollLearningPathUser(ctx context.Context, pathID string, userID string) error
	GetLearningPathEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error)
	GetUserPathProgress(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error)
	UpdateLearningPathImage(ctx context.Context, pathID string, coverImgURL string) error
	GetUserEnrolledPaths(ctx context.Context, userID string) ([]model.EnrolledPathResponse, error)
	UpdatePathEnrollmentCompletion(ctx context.Context, pathID string, userID string) error
	GetPathCreatorVerification(ctx context.Context, userID string) (string, bool, error)
}

type RepositoryNode interface {
	GetNodeByID(ctx context.Context, nodeID string, userID string) (*model.Node, error)
	CreateNode(ctx context.Context, req model.CreateNodeRequest) (string, error)
	GetNodesByPathID(ctx context.Context, pathID string, userID string) ([]model.Node, error)
	UpdateNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error
	DeleteNode(ctx context.Context, nodeID string) error
	UpdateNodeProgress(ctx context.Context, nodeID string, userID string, status string) error
	CreateMaterial(ctx context.Context, req model.CreateMaterialRequest) (string, error)
	GetMaterialsByNodeID(ctx context.Context, nodeID string) ([]model.NodeMaterial, error)
	DeleteMaterial(ctx context.Context, materialID string) error
	UpdateNodeSequence(ctx context.Context, nodeIDs []string) error
	CreateNodeWithContent(ctx context.Context, req model.CreateNodeRequest) (string, error)
	UpdateNodeProgressStatus(ctx context.Context, nodeID string, userID string) error
	UpdateNodeProgressCompletion(ctx context.Context, nodeID string, userID string) error
}

type RepositoryComment interface {
	CreateComment(ctx context.Context, req model.CreateCommentRequest) (string, error)
	GetCommentsByNodeID(ctx context.Context, nodeID string) ([]model.NodeComment, error)
	GetCommentsByPathID(ctx context.Context, pathID string) ([]model.NodeComment, error)
	DeleteComment(ctx context.Context, commentID, userID string) error
	ToggleReaction(ctx context.Context, req model.CreateReactionRequest) (bool, error)
	GetReactionsByCommentID(ctx context.Context, commentID string) ([]model.CommentReaction, error)
	CreateMention(ctx context.Context, req model.CreateMentionRequest) (string, error)
	UpdateComment(ctx context.Context, userID, messageID, message string) (bool, error)
	GetCommentOwner(ctx context.Context, commentID string) (string, error)
}

type RepositoryQuiz interface {
	CreateQuestion(ctx context.Context, req model.CreateQuestionRequest) (string, error)
	GetQuestionsByNodeID(ctx context.Context, nodeID string) ([]model.NodeQuestion, error)
	DeleteQuestion(ctx context.Context, questionID string) error
	CreateChoice(ctx context.Context, req model.CreateChoiceRequest) (string, error)
	GetChoicesByQuestionID(ctx context.Context, questionID string) ([]model.QuestionChoice, error)
	DeleteChoice(ctx context.Context, choiceID string) error
}

type RepositoryRating interface {
	UpsertLearningPathRating(ctx context.Context, rating *model.LearningPathRating) error
	GetRatingByUser(ctx context.Context, pathID string, userID string) (*model.LearningPathRating, error)
	DeleteLearningPathRating(ctx context.Context, pathID string, userID string) error
}

type RepositoryProgress interface {
	GetUserPathProgress(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error)
}

type RepositoryHistory interface {
	GetHistoryByUserID(ctx context.Context, userID string) ([]model.HistoryResponse, error)
}

type RepositoryResume interface {
	GetNextNodeID(ctx context.Context, userID string, pathID string) (string, error)
}

type RepositoryXP interface {
	AddXPAndRecalcLevel(ctx context.Context, userID string, xpAmount int) error
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
	RepositoryRating
	RepositoryProgress
	RepositoryHistory
	RepositoryResume
	RepositoryXP
}

type repositoryImpl struct {
	db Database
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}