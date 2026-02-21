package repository_test

import (
	"context"

	"passiontree/internal/reflection/model"
)

// MockRefRepo implements repository.RepositoryReflection
type MockRefRepo struct {
	GetReflectionByIDFunc func(ctx context.Context, reflectID string) (*model.Reflection, error)
	UpdateReflectionFunc  func(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error
	DeleteReflectionFunc  func(ctx context.Context, reflectID string) error
	GetAllReflectionsFunc func(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error)
	CreateReflectionFunc  func(ctx context.Context, req model.CreateReflectionRequest, summary, sentimentAnalysis string, primaryEmotion *string, strugglePoint string, aiConfidentScore, reflectionScore, weightedReflectionScore float64) (string, error)
}

func (m *MockRefRepo) GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error) {
	if m.GetReflectionByIDFunc != nil {
		return m.GetReflectionByIDFunc(ctx, reflectID)
	}
	return nil, nil
}
func (m *MockRefRepo) UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
	if m.UpdateReflectionFunc != nil {
		return m.UpdateReflectionFunc(ctx, reflectID, req)
	}
	return nil
}
func (m *MockRefRepo) DeleteReflection(ctx context.Context, reflectID string) error {
	if m.DeleteReflectionFunc != nil {
		return m.DeleteReflectionFunc(ctx, reflectID)
	}
	return nil
}
func (m *MockRefRepo) GetAllReflections(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error) {
	if m.GetAllReflectionsFunc != nil {
		return m.GetAllReflectionsFunc(ctx, filter)
	}
	return nil, nil
}
func (m *MockRefRepo) CreateReflection(ctx context.Context, req model.CreateReflectionRequest, summary, sentimentAnalysis string, primaryEmotion *string, strugglePoint string, aiConfidentScore, reflectionScore, weightedReflectionScore float64) (string, error) {
	if m.CreateReflectionFunc != nil {
		return m.CreateReflectionFunc(ctx, req, summary, sentimentAnalysis, primaryEmotion, strugglePoint, aiConfidentScore, reflectionScore, weightedReflectionScore)
	}
	return "", nil
}

// Implement other interface methods as no-ops
func (m *MockRefRepo) CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (string, error) {
	return "", nil
}
func (m *MockRefRepo) GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error) {
	return nil, nil
}
func (m *MockRefRepo) GetAlbumsByUserID(ctx context.Context, userID string) ([]model.Album, error) {
	return nil, nil
}
func (m *MockRefRepo) UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error {
	return nil
}
func (m *MockRefRepo) DeleteAlbum(ctx context.Context, albumID string) error { return nil }
func (m *MockRefRepo) CreateTree(ctx context.Context, req model.CreateTreeRequest) (string, error) {
	return "", nil
}
func (m *MockRefRepo) GetTreeByID(ctx context.Context, treeID string) (*model.Tree, error) {
	return nil, nil
}
func (m *MockRefRepo) GetTreesByAlbumID(ctx context.Context, albumID string) ([]model.Tree, error) {
	return nil, nil
}
func (m *MockRefRepo) GetTreesWithNodesByAlbumID(ctx context.Context, albumID string) ([]model.TreeResponse, error) {
	return nil, nil
}
func (m *MockRefRepo) UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
	return nil
}
func (m *MockRefRepo) DeleteTree(ctx context.Context, treeID string) error              { return nil }
func (m *MockRefRepo) PauseTree(ctx context.Context, treeID string, isPause bool) error { return nil }
func (m *MockRefRepo) AddSingleTreeNode(ctx context.Context, req model.CreateTreeNodeRequest) (string, error) {
	return "", nil
}
func (m *MockRefRepo) GetTreeNodesByTreeID(ctx context.Context, treeID string) ([]model.TreeNode, error) {
	return nil, nil
}
func (m *MockRefRepo) GetTreeNodeByID(ctx context.Context, treeNodeID string) (*model.TreeNode, error) {
	return nil, nil
}
func (m *MockRefRepo) UpdateTreeNode(ctx context.Context, treeNodeID string, req model.UpdateTreeNodeRequest) error {
	return nil
}
func (m *MockRefRepo) DeleteTreeNode(ctx context.Context, treeNodeID string) error { return nil }
func (m *MockRefRepo) CreateTreeNodes(ctx context.Context, treeID string, pathID string) error {
	return nil
}
func (m *MockRefRepo) GetNodesByPathID(ctx context.Context, pathID string) ([]model.TreeNode, error) {
	return nil, nil
}
