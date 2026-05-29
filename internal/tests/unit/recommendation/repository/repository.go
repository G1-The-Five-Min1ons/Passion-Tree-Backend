package repository_test

import (
	"context"

	lpmodel "passiontree/internal/learning-path/model"
	pathrepo "passiontree/internal/learning-path/repository"
	"passiontree/internal/recommendation/model"
)

// MockRecRepository satisfies recrepo.Repository for unit tests.
type MockRecRepository struct {
	GetUserReflectionsByTreeFunc    func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error)
	GetUserEnrolledPathsForRecFunc  func(ctx context.Context, userID string) ([]model.RecommendedPath, error)
	GetTopPopularPathsFunc          func(ctx context.Context) ([]model.RecommendedPath, error)
	GetBatchInteractionsFunc        func(ctx context.Context) ([]model.UserInteraction, error)
	GetBatchProfilesFunc            func(ctx context.Context) ([]model.UserProfile, error)
	GetUserInteractionsFunc         func(ctx context.Context, userID string) ([]model.UserInteraction, error)
	GetUserProfileFunc              func(ctx context.Context, userID string) (*model.UserProfile, error)
	SaveBatchRecommendationsFunc    func(ctx context.Context, results []model.BatchRecommendationResult) error
	GetSavedHomeRecommendationsFunc func(ctx context.Context, userID string) ([]model.RecommendedPath, error)
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

func (m *MockRecRepository) GetBatchInteractions(ctx context.Context) ([]model.UserInteraction, error) {
	if m.GetBatchInteractionsFunc != nil {
		return m.GetBatchInteractionsFunc(ctx)
	}
	return nil, nil
}

func (m *MockRecRepository) GetBatchProfiles(ctx context.Context) ([]model.UserProfile, error) {
	if m.GetBatchProfilesFunc != nil {
		return m.GetBatchProfilesFunc(ctx)
	}
	return nil, nil
}

func (m *MockRecRepository) GetUserInteractions(ctx context.Context, userID string) ([]model.UserInteraction, error) {
	if m.GetUserInteractionsFunc != nil {
		return m.GetUserInteractionsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRecRepository) GetUserProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRecRepository) SaveBatchRecommendations(ctx context.Context, results []model.BatchRecommendationResult) error {
	if m.SaveBatchRecommendationsFunc != nil {
		return m.SaveBatchRecommendationsFunc(ctx, results)
	}
	return nil
}

func (m *MockRecRepository) GetSavedHomeRecommendations(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
	if m.GetSavedHomeRecommendationsFunc != nil {
		return m.GetSavedHomeRecommendationsFunc(ctx, userID)
	}
	return nil, nil
}

// MockPathRepository satisfies pathrepo.RepositoryLearningPath for unit tests.
// Embeds the interface so only overridden methods need explicit stubs.
type MockPathRepository struct {
	pathrepo.RepositoryLearningPath
	GetLearningPathByIDFunc   func(ctx context.Context, pathID string) (*lpmodel.LearningPath, error)
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
