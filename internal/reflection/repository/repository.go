package repository

import (
	"context"
	"database/sql"
	"passiontree/internal/connection"
	"passiontree/internal/reflection/model"
)

type RepositoryReflection interface {
	CreateReflection(ctx context.Context, req model.CreateReflectionRequest, summary, sentimentAnalysis string, primaryEmotion *string, strugglePoint string, aiConfidentScore, reflectionScore, weightedReflectionScore float64) (string, error)
	GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error)
	GetAllReflections(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error)
	GetReflectionStats(ctx context.Context) (*model.ReflectionStats, error)
	UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflection(ctx context.Context, reflectID string) error

	// Album methods
	CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (string, error)
	GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error)
	GetAlbumsByUserID(ctx context.Context, userID string) ([]model.Album, error)
	UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error
	DeleteAlbum(ctx context.Context, albumID string) error

	// Tree methods
	CreateTree(ctx context.Context, req model.CreateTreeRequest) (string, error)
	GetTreeByID(ctx context.Context, treeID string) (*model.Tree, error)
	GetTreesByAlbumID(ctx context.Context, albumID string) ([]model.Tree, error)
	GetTreesWithNodesByAlbumID(ctx context.Context, albumID string, userID string) ([]model.TreeResponse, error)
	UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error
	DeleteTree(ctx context.Context, treeID string) error
	PauseTree(ctx context.Context, treeID string, isPause bool) error

	// Tree Node methods
	CreateStandaloneNode(ctx context.Context, title string, sequence int) (string, error)
	AddSingleTreeNode(ctx context.Context, req model.CreateTreeNodeRequest) (string, error)
	GetTreeNodesByTreeID(ctx context.Context, treeID string) ([]model.TreeNode, error)
	GetTreeNodeByID(ctx context.Context, treeNodeID string) (*model.TreeNode, error)
	UpdateTreeNode(ctx context.Context, treeNodeID string, req model.UpdateTreeNodeRequest) error
	DeleteTreeNode(ctx context.Context, treeNodeID string) error
	CreateTreeNodes(ctx context.Context, treeID string, pathID string) error
	GetNodesByPathID(ctx context.Context, pathID string) ([]model.TreeNode, error)
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(ds connection.Database) RepositoryReflection {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
