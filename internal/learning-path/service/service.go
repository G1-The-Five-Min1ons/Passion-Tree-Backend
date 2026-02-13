package service

import (
	"context"
	"log/slog"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/platform/aiclient"
)

type ServiceLearningPath interface {
	GetPaths(ctx context.Context) ([]model.LearningPaths, error)
	GetPathDetails(ctx context.Context, id string) (*model.LearningPath, error)
	CreatePath(ctx context.Context, req model.CreatePathRequest) (string, error)
	UpdatePath(ctx context.Context, path_id string, req model.UpdatePathRequest) error
	DeletePath(ctx context.Context, id string) error
	StartPath(ctx context.Context, pathID string, userID string) error
	GetEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error)
	GetPathProgress(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error)
	GeneratePathWithAI(ctx context.Context, topic string) (*model.GeneratedPathResponse, error)
	UpdatePathCoverImage(ctx context.Context, pathID string, coverImgURL string) error
}

type ServiceSearch interface {
	SearchLearningPaths(ctx context.Context, req model.SearchPathRequest) (*model.SearchPathResponse, error)
	GetCollectionInfo(collectionName string) (*aiclient.CollectionInfoResponse, error)
	SyncLearningPath(ctx context.Context, pathID string) (*model.SyncPathResponse, error)
}

type ServiceNode interface {
	GetNodeDetails(ctx context.Context, nodeID string) (*model.Node, error)
	AddNode(ctx context.Context, req model.CreateNodeRequest) (string, error)
	EditNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error
	RemoveNode(ctx context.Context, nodeID string) error
	AddMaterial(ctx context.Context, req model.CreateMaterialRequest) (string, error)
	RemoveMaterial(ctx context.Context, materialID string) error
	ReorderNodes(ctx context.Context, pathID string, req model.ReorderNodesRequest) error
	GetNodesByPathID(ctx context.Context, pathID string) ([]model.Node, error)
}

type ServiceComment interface {
	AddComment(ctx context.Context, req model.CreateCommentRequest) (string, error)
	GetNodeComments(ctx context.Context, nodeID string) ([]model.NodeComment, error)
	RemoveComment(ctx context.Context, commentID string) error
	AddReaction(ctx context.Context, req model.CreateReactionRequest) error
	AddMention(ctx context.Context, req model.CreateMentionRequest) (string, error)
}

type ServiceQuiz interface {
	AddQuestion(ctx context.Context, req model.CreateQuestionRequest) (string, error)
	GetQuestions(ctx context.Context, nodeID string) ([]model.NodeQuestion, error)
	RemoveQuestion(ctx context.Context, questionID string) error
	AddChoice(ctx context.Context, req model.CreateChoiceRequest) (string, error)
	RemoveChoice(ctx context.Context, choiceID string) error
}

type ServiceHistory interface {
	GetUserHistory(ctx context.Context, userID string) ([]model.HistoryResponse, error)
}

type ServiceResume interface {
	GetResumeNode(ctx context.Context, userID string, pathID string) (*model.ResumeResponse, error)
}

type Service interface {
	ServiceLearningPath
	ServiceSearch
	ServiceNode
	ServiceComment
	ServiceQuiz
	ServiceHistory
	ServiceResume
}

type serviceImpl struct {
	pathRepo    repository.RepositoryLearningPath
	nodeRepo    repository.RepositoryNode
	commentRepo repository.RepositoryComment
	quizRepo    repository.RepositoryQuiz
	historyRepo repository.RepositoryHistory
	resumeRepo  repository.RepositoryResume
	logger      *slog.Logger
	aiClient    *aiclient.AIClient
	storage     *storage.BlobService
}

func NewService(repo repository.Repository, aiClient *aiclient.AIClient, logger *slog.Logger) Service {
	if aiClient == nil {
		slog.Warn("[DEBUG] Warning: aiClient passed to NewService is NIL!")
	} else {
		slog.Info("[DEBUG] aiClient successfully passed to NewService", "aiClient", aiClient)
	}
	return &serviceImpl{
		pathRepo:    repo,
		nodeRepo:    repo,
		commentRepo: repo,
		quizRepo:    repo,
		logger:      logger,
		aiClient:    aiClient,
		historyRepo: repo,
		resumeRepo:  repo,
		storage:     nil,
	}
}
