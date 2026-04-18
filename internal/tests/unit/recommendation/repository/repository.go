package repository_test

import (
	"context"
	lpmodel "passiontree/internal/learning-path/model"
	pathrepo "passiontree/internal/learning-path/repository"
	"passiontree/internal/recommendation/model"
)

type MockRecRepository struct {
	GetUserReflectionsByTreeFunc   func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error)
	GetUserEnrolledPathsForRecFunc func(ctx context.Context, userID string) ([]model.RecommendedPath, error)
	GetTopPopularPathsFunc         func(ctx context.Context) ([]model.RecommendedPath, error)
}

func (m *MockRecRepository) GetUserReflectionsByTree(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
	if m.GetUserReflectionsByTreeFunc != nil {
		return m.GetUserReflectionsByTreeFunc(ctx, userID, treeID)
	}
	return nil, "", nil
}

func (m *MockRecRepository) GetUserEnrolledPathsForRec(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
	if m.GetUserEnrolledPathsForRecFunc != nil {
		return m.GetUserEnrolledPathsForRecFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRecRepository) GetTopPopularPaths(ctx context.Context) ([]model.RecommendedPath, error) {
	if m.GetTopPopularPathsFunc != nil {
		return m.GetTopPopularPathsFunc(ctx)
	}
	return nil, nil
}

type MockPathRepository struct {
	pathrepo.RepositoryLearningPath
	GetLearningPathByIDFunc  func(ctx context.Context, pathID string) (*lpmodel.LearningPath, error)
	GetLearningPathsByIDsFunc func(ctx context.Context, pathIDs []string) ([]lpmodel.LearningPath, error)
}

func (m *MockPathRepository) GetLearningPathByID(ctx context.Context, pathID string) (*lpmodel.LearningPath, error) {
	if m.GetLearningPathByIDFunc != nil {
		return m.GetLearningPathByIDFunc(ctx, pathID)
	}
	return nil, nil
}

func (m *MockPathRepository) GetLearningPathsByIDs(ctx context.Context, pathIDs []string) ([]lpmodel.LearningPath, error) {
	if m.GetLearningPathsByIDsFunc != nil {
		return m.GetLearningPathsByIDsFunc(ctx, pathIDs)
	}
	return nil, nil
}
