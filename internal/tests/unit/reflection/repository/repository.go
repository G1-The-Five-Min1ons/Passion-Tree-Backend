package repository_test

import (
	"context"

	"passiontree/internal/reflection/model"
)

// MockRefRepo implements repository.RepositoryReflection
type Repository struct {
	GetReflectionByIDFunc func(ctx context.Context, reflectID string) (*model.Reflection, error)
	UpdateReflectionFunc  func(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflectionFunc  func(ctx context.Context, reflectID string) error
	GetAllReflectionsFunc func(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error)
	CreateReflectionFunc  func(ctx context.Context, req model.CreateReflectionRequest, summary, sentimentAnalysis string, primaryEmotion *string, strugglePoint string, aiConfidentScore, reflectionScore, weightedReflectionScore float64) (string, error)

	// Album methods
	CreateAlbumFunc       func(ctx context.Context, req model.CreateAlbumRequest) (string, error)
	GetAlbumByIDFunc      func(ctx context.Context, albumID string) (*model.Album, error)
	GetAlbumsByUserIDFunc func(ctx context.Context, userID string) ([]model.Album, error)
	UpdateAlbumFunc       func(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error
	DeleteAlbumFunc       func(ctx context.Context, albumID string) error

	// Tree methods
	CreateTreeFunc                 func(ctx context.Context, req model.CreateTreeRequest) (string, error)
	GetTreeByIDFunc                func(ctx context.Context, treeID string) (*model.Tree, error)
	GetTreesByAlbumIDFunc          func(ctx context.Context, albumID string) ([]model.Tree, error)
	GetTreesWithNodesByAlbumIDFunc func(ctx context.Context, albumID string, userID string) ([]model.TreeResponse, error)
	UpdateTreeFunc                 func(ctx context.Context, treeID string, req model.UpdateTreeRequest) error
	DeleteTreeFunc                 func(ctx context.Context, treeID string) error
	PauseTreeFunc                  func(ctx context.Context, treeID string, isPause bool) error

	// Tree Node methods
	AddSingleTreeNodeFunc    func(ctx context.Context, req model.CreateTreeNodeRequest) (string, error)
	GetTreeNodesByTreeIDFunc func(ctx context.Context, treeID string) ([]model.TreeNode, error)
	GetTreeNodeByIDFunc      func(ctx context.Context, treeNodeID string) (*model.TreeNode, error)
	UpdateTreeNodeFunc       func(ctx context.Context, treeNodeID string, req model.UpdateTreeNodeRequest) error
	DeleteTreeNodeFunc       func(ctx context.Context, treeNodeID string) error
	CreateTreeNodesFunc      func(ctx context.Context, treeID string, pathID string) error
	GetNodesByPathIDFunc     func(ctx context.Context, pathID string) ([]model.TreeNode, error)
}

func (m *Repository) GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error) {
	if m.GetReflectionByIDFunc != nil {
		return m.GetReflectionByIDFunc(ctx, reflectID)
	}
	return nil, nil
}
func (m *Repository) UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
	if m.UpdateReflectionFunc != nil {
		return m.UpdateReflectionFunc(ctx, reflectID, req)
	}
	return nil
}
func (m *Repository) DeleteReflection(ctx context.Context, reflectID string) error {
	if m.DeleteReflectionFunc != nil {
		return m.DeleteReflectionFunc(ctx, reflectID)
	}
	return nil
}
func (m *Repository) GetAllReflections(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error) {
	if m.GetAllReflectionsFunc != nil {
		return m.GetAllReflectionsFunc(ctx, filter)
	}
	return nil, nil
}
func (m *Repository) CreateReflection(ctx context.Context, req model.CreateReflectionRequest, summary, sentimentAnalysis string, primaryEmotion *string, strugglePoint string, aiConfidentScore, reflectionScore, weightedReflectionScore float64) (string, error) {
	if m.CreateReflectionFunc != nil {
		return m.CreateReflectionFunc(ctx, req, summary, sentimentAnalysis, primaryEmotion, strugglePoint, aiConfidentScore, reflectionScore, weightedReflectionScore)
	}
	return "", nil
}

// Album methods
func (m *Repository) CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (string, error) {
	if m.CreateAlbumFunc != nil {
		return m.CreateAlbumFunc(ctx, req)
	}
	return "", nil
}
func (m *Repository) GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error) {
	if m.GetAlbumByIDFunc != nil {
		return m.GetAlbumByIDFunc(ctx, albumID)
	}
	return nil, nil
}
func (m *Repository) GetAlbumsByUserID(ctx context.Context, userID string) ([]model.Album, error) {
	if m.GetAlbumsByUserIDFunc != nil {
		return m.GetAlbumsByUserIDFunc(ctx, userID)
	}
	return nil, nil
}
func (m *Repository) UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error {
	if m.UpdateAlbumFunc != nil {
		return m.UpdateAlbumFunc(ctx, albumID, req)
	}
	return nil
}
func (m *Repository) DeleteAlbum(ctx context.Context, albumID string) error {
	if m.DeleteAlbumFunc != nil {
		return m.DeleteAlbumFunc(ctx, albumID)
	}
	return nil
}

// Tree methods
func (m *Repository) CreateTree(ctx context.Context, req model.CreateTreeRequest) (string, error) {
	if m.CreateTreeFunc != nil {
		return m.CreateTreeFunc(ctx, req)
	}
	return "", nil
}
func (m *Repository) GetTreeByID(ctx context.Context, treeID string) (*model.Tree, error) {
	if m.GetTreeByIDFunc != nil {
		return m.GetTreeByIDFunc(ctx, treeID)
	}
	return nil, nil
}
func (m *Repository) GetTreesByAlbumID(ctx context.Context, albumID string) ([]model.Tree, error) {
	if m.GetTreesByAlbumIDFunc != nil {
		return m.GetTreesByAlbumIDFunc(ctx, albumID)
	}
	return nil, nil
}
func (m *Repository) GetTreesWithNodesByAlbumID(ctx context.Context, albumID string, userID string) ([]model.TreeResponse, error) {
	if m.GetTreesWithNodesByAlbumIDFunc != nil {
		return m.GetTreesWithNodesByAlbumIDFunc(ctx, albumID, userID)
	}
	return nil, nil
}
func (m *Repository) UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
	if m.UpdateTreeFunc != nil {
		return m.UpdateTreeFunc(ctx, treeID, req)
	}
	return nil
}
func (m *Repository) DeleteTree(ctx context.Context, treeID string) error {
	if m.DeleteTreeFunc != nil {
		return m.DeleteTreeFunc(ctx, treeID)
	}
	return nil
}
func (m *Repository) PauseTree(ctx context.Context, treeID string, isPause bool) error {
	if m.PauseTreeFunc != nil {
		return m.PauseTreeFunc(ctx, treeID, isPause)
	}
	return nil
}

// Tree Node methods
func (m *Repository) AddSingleTreeNode(ctx context.Context, req model.CreateTreeNodeRequest) (string, error) {
	if m.AddSingleTreeNodeFunc != nil {
		return m.AddSingleTreeNodeFunc(ctx, req)
	}
	return "", nil
}
func (m *Repository) GetTreeNodesByTreeID(ctx context.Context, treeID string) ([]model.TreeNode, error) {
	if m.GetTreeNodesByTreeIDFunc != nil {
		return m.GetTreeNodesByTreeIDFunc(ctx, treeID)
	}
	return nil, nil
}
func (m *Repository) GetTreeNodeByID(ctx context.Context, treeNodeID string) (*model.TreeNode, error) {
	if m.GetTreeNodeByIDFunc != nil {
		return m.GetTreeNodeByIDFunc(ctx, treeNodeID)
	}
	return nil, nil
}
func (m *Repository) UpdateTreeNode(ctx context.Context, treeNodeID string, req model.UpdateTreeNodeRequest) error {
	if m.UpdateTreeNodeFunc != nil {
		return m.UpdateTreeNodeFunc(ctx, treeNodeID, req)
	}
	return nil
}
func (m *Repository) DeleteTreeNode(ctx context.Context, treeNodeID string) error {
	if m.DeleteTreeNodeFunc != nil {
		return m.DeleteTreeNodeFunc(ctx, treeNodeID)
	}
	return nil
}
func (m *Repository) CreateTreeNodes(ctx context.Context, treeID string, pathID string) error {
	if m.CreateTreeNodesFunc != nil {
		return m.CreateTreeNodesFunc(ctx, treeID, pathID)
	}
	return nil
}
func (m *Repository) GetNodesByPathID(ctx context.Context, pathID string) ([]model.TreeNode, error) {
	if m.GetNodesByPathIDFunc != nil {
		return m.GetNodesByPathIDFunc(ctx, pathID)
	}
	return nil, nil
}
