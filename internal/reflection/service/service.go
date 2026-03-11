package service

import (
	"context"
	"log/slog"
	"passiontree/internal/reflection/model"
	"passiontree/internal/reflection/repository"
	"passiontree/internal/platform/aiclient"
)

type ReflectionService interface {
	CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (*model.ReflectionResponse, error)
	GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error)
	GetAllReflections(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error)
	UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflection(ctx context.Context, reflectID string) error
	
	// Album methods
	CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (*model.AlbumResponse, error)
	GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error)
	GetAlbumsByUserID(ctx context.Context, userID string) ([]model.Album, error)
	UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error
	DeleteAlbum(ctx context.Context, albumID string) error
	
	// Tree methods
	CreateTree(ctx context.Context, req model.CreateTreeRequest) (*model.TreeResponse, error)
	GetTreeByID(ctx context.Context, treeID string) (*model.Tree, error)
	GetTreesByAlbumID(ctx context.Context, albumID string, includeNodes bool, userID string) (interface{}, error)
	UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error
	DeleteTree(ctx context.Context, treeID string) error
	PauseTree(ctx context.Context, treeID string, req model.PauseTreeRequest) (bool, error)
	
	// Tree Node methods
	CreateTreeNode(ctx context.Context, req model.CreateTreeNodeRequest) (*model.TreeNodeResponse, error)
	GetTreeNodesByTreeID(ctx context.Context, treeID string) ([]model.TreeNode, error)
	GetTreeNodeByID(ctx context.Context, treeNodeID string) (*model.TreeNode, error)
	UpdateTreeNode(ctx context.Context, treeNodeID string, req model.UpdateTreeNodeRequest) error
	DeleteTreeNode(ctx context.Context, treeNodeID string) error
}

type serviceImpl struct {
	refRepo  repository.RepositoryReflection
	aiClient *aiclient.AIClient
	logger   *slog.Logger
}

func NewService(repo repository.RepositoryReflection, aiClient *aiclient.AIClient, logger *slog.Logger) ReflectionService {
	return &serviceImpl{
		refRepo:  repo,
		aiClient: aiClient,
		logger:   logger,
	}
}