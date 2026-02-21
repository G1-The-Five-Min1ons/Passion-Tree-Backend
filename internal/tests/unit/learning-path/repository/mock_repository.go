package repository_test

import (
	"context"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
)

// MockRepo for LearningPath
type MockRepo struct {
	GetAllLearningPathFunc              func(ctx context.Context) ([]model.LearningPath, error)
	GetLearningPathByIDFunc             func(ctx context.Context, path_id string) (*model.LearningPath, error)
	CreateLearningPathFunc              func(ctx context.Context, req model.CreatePathRequest) (string, error)
	UpdateLearningPathFunc              func(ctx context.Context, path_id string, req model.UpdatePathRequest) error
	DeleteLearningPathFunc              func(ctx context.Context, path_id string) error
	EnrollLearningPathUserFunc          func(ctx context.Context, pathID string, userID string) error
	GetLearningPathEnrollmentStatusFunc func(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error)
	GetUserPathProgressFunc             func(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error)
	UpdateLearningPathImageFunc         func(ctx context.Context, pathID string, coverImgURL string) error
}

func (m *MockRepo) GetAllLearningPath(ctx context.Context) ([]model.LearningPath, error) {
	if m.GetAllLearningPathFunc != nil {
		return m.GetAllLearningPathFunc(ctx)
	}
	return nil, nil
}
func (m *MockRepo) GetLearningPathByID(ctx context.Context, path_id string) (*model.LearningPath, error) {
	if m.GetLearningPathByIDFunc != nil {
		return m.GetLearningPathByIDFunc(ctx, path_id)
	}
	return nil, nil
}
func (m *MockRepo) CreateLearningPath(ctx context.Context, req model.CreatePathRequest) (string, error) {
	if m.CreateLearningPathFunc != nil {
		return m.CreateLearningPathFunc(ctx, req)
	}
	return "", nil
}
func (m *MockRepo) UpdateLearningPath(ctx context.Context, path_id string, req model.UpdatePathRequest) error {
	if m.UpdateLearningPathFunc != nil {
		return m.UpdateLearningPathFunc(ctx, path_id, req)
	}
	return nil
}
func (m *MockRepo) DeleteLearningPath(ctx context.Context, path_id string) error {
	if m.DeleteLearningPathFunc != nil {
		return m.DeleteLearningPathFunc(ctx, path_id)
	}
	return nil
}
func (m *MockRepo) EnrollLearningPathUser(ctx context.Context, pathID string, userID string) error {
	if m.EnrollLearningPathUserFunc != nil {
		return m.EnrollLearningPathUserFunc(ctx, pathID, userID)
	}
	return nil
}
func (m *MockRepo) GetLearningPathEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error) {
	if m.GetLearningPathEnrollmentStatusFunc != nil {
		return m.GetLearningPathEnrollmentStatusFunc(ctx, pathID, userID)
	}
	return nil, nil
}
func (m *MockRepo) GetUserPathProgress(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error) {
	if m.GetUserPathProgressFunc != nil {
		return m.GetUserPathProgressFunc(ctx, pathID, userID)
	}
	return nil, nil
}
func (m *MockRepo) UpdateLearningPathImage(ctx context.Context, pathID string, coverImgURL string) error {
	if m.UpdateLearningPathImageFunc != nil {
		return m.UpdateLearningPathImageFunc(ctx, pathID, coverImgURL)
	}
	return nil
}

// Implement Database interface
func (m *MockRepo) GetDB() repository.Database { return nil }

// Implement RepositoryNode
func (m *MockRepo) GetNodeByID(ctx context.Context, nodeID string) (*model.Node, error) {
	return nil, nil
}
func (m *MockRepo) CreateNode(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	return "", nil
}
func (m *MockRepo) GetNodesByPathID(ctx context.Context, pathID string) ([]model.Node, error) {
	return nil, nil
}
func (m *MockRepo) UpdateNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error {
	return nil
}
func (m *MockRepo) DeleteNode(ctx context.Context, nodeID string) error { return nil }
func (m *MockRepo) CreateMaterial(ctx context.Context, req model.CreateMaterialRequest) (string, error) {
	return "", nil
}
func (m *MockRepo) GetMaterialsByNodeID(ctx context.Context, nodeID string) ([]model.NodeMaterial, error) {
	return nil, nil
}
func (m *MockRepo) DeleteMaterial(ctx context.Context, materialID string) error    { return nil }
func (m *MockRepo) UpdateNodeSequence(ctx context.Context, nodeIDs []string) error { return nil }
func (m *MockRepo) CreateNodeWithContent(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	return "", nil
}

// Implement RepositoryComment
func (m *MockRepo) CreateComment(ctx context.Context, req model.CreateCommentRequest) (string, error) {
	return "", nil
}
func (m *MockRepo) GetCommentsByNodeID(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
	return nil, nil
}
func (m *MockRepo) DeleteComment(ctx context.Context, commentID string) error { return nil }
func (m *MockRepo) CreateReaction(ctx context.Context, req model.CreateReactionRequest) error {
	return nil
}
func (m *MockRepo) GetReactionsByCommentID(ctx context.Context, commentID string) ([]model.CommentReaction, error) {
	return nil, nil
}
func (m *MockRepo) CreateMention(ctx context.Context, req model.CreateMentionRequest) (string, error) {
	return "", nil
}

// Implement RepositoryQuiz
func (m *MockRepo) CreateQuestion(ctx context.Context, req model.CreateQuestionRequest) (string, error) {
	return "", nil
}
func (m *MockRepo) GetQuestionsByNodeID(ctx context.Context, nodeID string) ([]model.NodeQuestion, error) {
	return nil, nil
}
func (m *MockRepo) DeleteQuestion(ctx context.Context, questionID string) error { return nil }
func (m *MockRepo) CreateChoice(ctx context.Context, req model.CreateChoiceRequest) (string, error) {
	return "", nil
}
func (m *MockRepo) GetChoicesByQuestionID(ctx context.Context, questionID string) ([]model.QuestionChoice, error) {
	return nil, nil
}
func (m *MockRepo) DeleteChoice(ctx context.Context, choiceID string) error { return nil }

// Implement RepositoryHistory
func (m *MockRepo) GetHistoryByUserID(ctx context.Context, userID string) ([]model.HistoryResponse, error) {
	return nil, nil
}

// Implement RepositoryResume
func (m *MockRepo) GetNextNodeID(ctx context.Context, userID string, pathID string) (string, error) {
	return "", nil
}
