package repository

import (
	"database/sql"
	"passiontree/internal/database"
	"passiontree/internal/learning-path/model"
)

type Repository interface {
	GetAllLearnningPath() ([]model.LearningPath, error)
	GetLearnningPathByID(id string) (*model.LearningPath, error)
	CreateLearnningPath(req model.CreatePathRequest) (string, error)
	UpdateLearnningPath(id string, req model.UpdatePathRequest) error
	DeleteLearnningPath(id string) error
	EnrollLearnningPathUser(pathID string, userID string) error
	GetLearnningPathEnrollmentStatus(pathID string, userID string) (*model.PathEnroll, error)

	CreateNode(req model.CreateNodeRequest) (string, error)
	GetNodesByPathID(pathID string) ([]model.Node, error)
	UpdateNode(nodeID string, req model.UpdateNodeRequest) error
	DeleteNode(nodeID string) error
	CreateMaterial(req model.CreateMaterialRequest) (string, error)
	GetMaterialsByNodeID(nodeID string) ([]model.NodeMaterial, error)
	DeleteMaterial(materialID string) error

	CreateComment(req model.CreateCommentRequest) (string, error)
	GetCommentsByNodeID(nodeID string) ([]model.NodeComment, error)
	DeleteComment(commentID string) error
	CreateReaction(req model.CreateReactionRequest) error
	GetReactionsByCommentID(commentID string) ([]model.CommentReaction, error)
	CreateMention(req model.CreateMentionRequest) (string, error)

	CreateQuestion(req model.CreateQuestionRequest) (string, error)
	GetQuestionsByNodeID(nodeID string) ([]model.NodeQuestion, error)
	DeleteQuestion(questionID string) error
	CreateChoice(req model.CreateChoiceRequest) (string, error)
	GetChoicesByQuestionID(questionID string) ([]model.QuestionChoice, error)
	DeleteChoice(choiceID string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(ds database.Database) Repository {
	return &repository{
		db: ds.GetDB(),
	}
}