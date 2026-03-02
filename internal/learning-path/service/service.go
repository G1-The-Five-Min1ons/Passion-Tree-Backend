package service

import (
	"context"
	"fmt"
	"log/slog"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/platform/aiclient"
)

type ServiceLearningPath interface {
	GetPaths(ctx context.Context) ([]model.LearningPath, error)
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
	RemoveComment(ctx context.Context, userID, commentID string) error
	UpdateComment(ctx context.Context, userID, commentID, message string) (bool, error)
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

// ServiceComment implementation for serviceImpl
func (s *serviceImpl) AddComment(ctx context.Context, req model.CreateCommentRequest) (string, error) {
	// If this is a reply, validate the parent exists BEFORE inserting
	var parentOwnerID string
	if req.ParentID != nil && *req.ParentID != "" {
		ownerID, err := s.commentRepo.GetCommentOwner(ctx, *req.ParentID)
		if err != nil {
			return "", fmt.Errorf("parent comment not found: %w", err)
		}
		parentOwnerID = ownerID
	}

	commentID, err := s.commentRepo.CreateComment(ctx, req)
	if err != nil {
		return "", err
	}

	// Auto-mention the parent comment's author when replying to someone else
	if parentOwnerID != "" && parentOwnerID != req.UserID {
		mentionReq := model.CreateMentionRequest{
			MentionerUserID: req.UserID,
			MentionedUserID: parentOwnerID,
			CommentID:       commentID,
		}
		// Best-effort: reply is still successful even if mention insert fails
		_, _ = s.commentRepo.CreateMention(ctx, mentionReq)
	}

	return commentID, nil
}

func (s *serviceImpl) GetNodeComments(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
	return s.commentRepo.GetCommentsByNodeID(ctx, nodeID)
}

func (s *serviceImpl) RemoveComment(ctx context.Context, userID, commentID string) error {
	return s.commentRepo.DeleteComment(ctx, commentID, userID)
}

func (s *serviceImpl) UpdateComment(ctx context.Context, userID, commentID, message string) (bool, error) {
	return s.commentRepo.UpdateComment(ctx, userID, commentID, message)
}

func (s *serviceImpl) AddReaction(ctx context.Context, req model.CreateReactionRequest) error {
	return s.commentRepo.CreateReaction(ctx, req)
}

func (s *serviceImpl) AddMention(ctx context.Context, req model.CreateMentionRequest) (string, error) {
	return s.commentRepo.CreateMention(ctx, req)
}

func NewService(repo repository.Repository, aiClient *aiclient.AIClient, logger *slog.Logger) Service {
	return &serviceImpl{
		pathRepo:    repo,
		nodeRepo:    repo,
		commentRepo: repo,
		quizRepo:    repo,
		historyRepo: repo,
		resumeRepo:  repo,
		logger:      logger,
		aiClient:    aiClient,
		storage:     nil,
	}
}
