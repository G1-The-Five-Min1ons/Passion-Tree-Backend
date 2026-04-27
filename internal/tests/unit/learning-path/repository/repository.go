package repository

import (
	"context"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
)

// MockRepo for LearningPath
type Repopository struct {
	GetAllLearningPathFunc              func(ctx context.Context) ([]model.LearningPath, error)
	GetLearningPathByIDFunc             func(ctx context.Context, path_id string) (*model.LearningPath, error)
	GetLearningPathsByIDsFunc           func(ctx context.Context, pathIDs []string) ([]model.LearningPath, error)
	CreateLearningPathFunc              func(ctx context.Context, req model.CreatePathRequest) (string, error)
	UpdateLearningPathFunc              func(ctx context.Context, path_id string, req model.UpdatePathRequest) error
	DeleteLearningPathFunc              func(ctx context.Context, path_id string) error
	EnrollLearningPathUserFunc          func(ctx context.Context, pathID string, userID string) error
	GetLearningPathEnrollmentStatusFunc func(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error)
	GetUserPathProgressFunc             func(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error)
	UpdateLearningPathImageFunc         func(ctx context.Context, pathID string, coverImgURL string) error
	GetUserEnrolledPathsFunc            func(ctx context.Context, userID string) ([]model.EnrolledPathResponse, error)
	UpdatePathEnrollmentCompletionFunc  func(ctx context.Context, pathID string, userID string) error
	GetPathCreatorVerificationFunc      func(ctx context.Context, userID string) (string, bool, error)

	// Mock hooks for Node
	GetNodeByIDFunc                  func(ctx context.Context, nodeID string, userID string) (*model.Node, error)
	CreateNodeFunc                   func(ctx context.Context, req model.CreateNodeRequest) (string, error)
	GetNodesByPathIDFunc             func(ctx context.Context, pathID string, userID string) ([]model.Node, error)
	UpdateNodeFunc                   func(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error
	DeleteNodeFunc                   func(ctx context.Context, nodeID string) error
	CreateMaterialFunc               func(ctx context.Context, req model.CreateMaterialRequest) (string, error)
	GetMaterialsByNodeIDFunc         func(ctx context.Context, nodeID string) ([]model.NodeMaterial, error)
	DeleteMaterialFunc               func(ctx context.Context, materialID string) error
	UpdateNodeSequenceFunc           func(ctx context.Context, nodeIDs []string) error
	CreateNodeWithContentFunc        func(ctx context.Context, req model.CreateNodeRequest) (string, error)
	UpdateNodeProgressStatusFunc     func(ctx context.Context, nodeID string, userID string) error
	UpdateNodeProgressCompletionFunc func(ctx context.Context, nodeID string, userID string) error
	UpdateNodeProgressFunc           func(ctx context.Context, nodeID string, userID string, status string) error

	// Mock hooks for Comment
	CreateCommentFunc           func(ctx context.Context, req model.CreateCommentRequest) (string, error)
	GetCommentsByNodeIDFunc     func(ctx context.Context, nodeID string) ([]model.NodeComment, error)
	GetCommentsByPathIDFunc     func(ctx context.Context, pathID string) ([]model.NodeComment, error)
	DeleteCommentFunc           func(ctx context.Context, commentID, userID string) error
	ToggleReactionFunc          func(ctx context.Context, req model.CreateReactionRequest) (bool, error)
	GetReactionsByCommentIDFunc func(ctx context.Context, commentID string) ([]model.CommentReaction, error)
	CreateMentionFunc           func(ctx context.Context, req model.CreateMentionRequest) (string, error)
	GetCommentOwnerFunc         func(ctx context.Context, commentID string) (string, error)
	UpdateCommentFunc           func(ctx context.Context, userID, messageID, message string) (bool, error)

	// Mock hooks for Quiz
	CreateQuestionFunc         func(ctx context.Context, req model.CreateQuestionRequest) (string, error)
	GetQuestionsByNodeIDFunc   func(ctx context.Context, nodeID string) ([]model.NodeQuestion, error)
	DeleteQuestionFunc         func(ctx context.Context, questionID string) error
	CreateChoiceFunc           func(ctx context.Context, req model.CreateChoiceRequest) (string, error)
	GetChoicesByQuestionIDFunc func(ctx context.Context, questionID string) ([]model.QuestionChoice, error)
	DeleteChoiceFunc           func(ctx context.Context, choiceID string) error
	GetQuestionByIDFunc        func(ctx context.Context, questionID string) (*model.NodeQuestion, error)
	UpdateQuestionFunc         func(ctx context.Context, questionID string, req model.UpdateQuestionRequest) error
	GetChoiceByIDFunc          func(ctx context.Context, choiceID string) (*model.QuestionChoice, error)
	UpdateChoiceFunc           func(ctx context.Context, choiceID string, req model.UpdateChoiceRequest) error

	// Mock hooks for Rating
	UpsertLearningPathRatingFunc  func(ctx context.Context, rating *model.LearningPathRating) error
	GetRatingByUserFunc           func(ctx context.Context, pathID string, userID string) (*model.LearningPathRating, error)
	DeleteLearningPathRatingFunc  func(ctx context.Context, pathID string, userID string) error

	// Mock hooks for History & Resume
	GetHistoryByUserIDFunc func(ctx context.Context, userID string) ([]model.HistoryResponse, error)
	GetNextNodeIDFunc      func(ctx context.Context, userID string, pathID string) (string, error)

	// Mock hooks for XP
	AddXPAndRecalcLevelFunc func(ctx context.Context, userID string, xpAmount int) error
}

func (m *Repopository) GetAllLearningPath(ctx context.Context) ([]model.LearningPath, error) {
	if m.GetAllLearningPathFunc != nil {
		return m.GetAllLearningPathFunc(ctx)
	}
	return nil, nil
}
func (m *Repopository) GetLearningPathByID(ctx context.Context, path_id string) (*model.LearningPath, error) {
	if m.GetLearningPathByIDFunc != nil {
		return m.GetLearningPathByIDFunc(ctx, path_id)
	}
	return nil, nil
}
func (m *Repopository) GetLearningPathsByIDs(ctx context.Context, pathIDs []string) ([]model.LearningPath, error) {
	if m.GetLearningPathsByIDsFunc != nil {
		return m.GetLearningPathsByIDsFunc(ctx, pathIDs)
	}
	return nil, nil
}
func (m *Repopository) CreateLearningPath(ctx context.Context, req model.CreatePathRequest) (string, error) {
	if m.CreateLearningPathFunc != nil {
		return m.CreateLearningPathFunc(ctx, req)
	}
	return "", nil
}
func (m *Repopository) UpdateLearningPath(ctx context.Context, path_id string, req model.UpdatePathRequest) error {
	if m.UpdateLearningPathFunc != nil {
		return m.UpdateLearningPathFunc(ctx, path_id, req)
	}
	return nil
}
func (m *Repopository) DeleteLearningPath(ctx context.Context, path_id string) error {
	if m.DeleteLearningPathFunc != nil {
		return m.DeleteLearningPathFunc(ctx, path_id)
	}
	return nil
}
func (m *Repopository) EnrollLearningPathUser(ctx context.Context, pathID string, userID string) error {
	if m.EnrollLearningPathUserFunc != nil {
		return m.EnrollLearningPathUserFunc(ctx, pathID, userID)
	}
	return nil
}
func (m *Repopository) GetLearningPathEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error) {
	if m.GetLearningPathEnrollmentStatusFunc != nil {
		return m.GetLearningPathEnrollmentStatusFunc(ctx, pathID, userID)
	}
	return nil, nil
}
func (m *Repopository) GetUserPathProgress(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error) {
	if m.GetUserPathProgressFunc != nil {
		return m.GetUserPathProgressFunc(ctx, pathID, userID)
	}
	return nil, nil
}
func (m *Repopository) UpdateLearningPathImage(ctx context.Context, pathID string, coverImgURL string) error {
	if m.UpdateLearningPathImageFunc != nil {
		return m.UpdateLearningPathImageFunc(ctx, pathID, coverImgURL)
	}
	return nil
}
func (m *Repopository) GetUserEnrolledPaths(ctx context.Context, userID string) ([]model.EnrolledPathResponse, error) {
	if m.GetUserEnrolledPathsFunc != nil {
		return m.GetUserEnrolledPathsFunc(ctx, userID)
	}
	return nil, nil
}
func (m *Repopository) UpdatePathEnrollmentCompletion(ctx context.Context, pathID string, userID string) error {
	if m.UpdatePathEnrollmentCompletionFunc != nil {
		return m.UpdatePathEnrollmentCompletionFunc(ctx, pathID, userID)
	}
	return nil
}

func (m *Repopository) GetPathCreatorVerification(ctx context.Context, userID string) (string, bool, error) {
	if m.GetPathCreatorVerificationFunc != nil {
		return m.GetPathCreatorVerificationFunc(ctx, userID)
	}
	return "", false, nil
}

// Implement Database interface
func (m *Repopository) GetDB() repository.Database { return nil }

// Implement RepositoryNode
func (m *Repopository) GetNodeByID(ctx context.Context, nodeID string, userID string) (*model.Node, error) {
	if m.GetNodeByIDFunc != nil {
		return m.GetNodeByIDFunc(ctx, nodeID, userID)
	}
	return nil, nil
}
func (m *Repopository) CreateNode(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	if m.CreateNodeFunc != nil {
		return m.CreateNodeFunc(ctx, req)
	}
	return "", nil
}
func (m *Repopository) GetNodesByPathID(ctx context.Context, pathID string, userID string) ([]model.Node, error) {
	if m.GetNodesByPathIDFunc != nil {
		return m.GetNodesByPathIDFunc(ctx, pathID, userID)
	}
	return nil, nil
}
func (m *Repopository) UpdateNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error {
	if m.UpdateNodeFunc != nil {
		return m.UpdateNodeFunc(ctx, nodeID, req)
	}
	return nil
}
func (m *Repopository) DeleteNode(ctx context.Context, nodeID string) error {
	if m.DeleteNodeFunc != nil {
		return m.DeleteNodeFunc(ctx, nodeID)
	}
	return nil
}
func (m *Repopository) CreateMaterial(ctx context.Context, req model.CreateMaterialRequest) (string, error) {
	if m.CreateMaterialFunc != nil {
		return m.CreateMaterialFunc(ctx, req)
	}
	return "", nil
}
func (m *Repopository) GetMaterialsByNodeID(ctx context.Context, nodeID string) ([]model.NodeMaterial, error) {
	if m.GetMaterialsByNodeIDFunc != nil {
		return m.GetMaterialsByNodeIDFunc(ctx, nodeID)
	}
	return nil, nil
}
func (m *Repopository) DeleteMaterial(ctx context.Context, materialID string) error {
	if m.DeleteMaterialFunc != nil {
		return m.DeleteMaterialFunc(ctx, materialID)
	}
	return nil
}
func (m *Repopository) UpdateNodeSequence(ctx context.Context, nodeIDs []string) error {
	if m.UpdateNodeSequenceFunc != nil {
		return m.UpdateNodeSequenceFunc(ctx, nodeIDs)
	}
	return nil
}
func (m *Repopository) CreateNodeWithContent(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	if m.CreateNodeWithContentFunc != nil {
		return m.CreateNodeWithContentFunc(ctx, req)
	}
	return "", nil
}
func (m *Repopository) UpdateNodeProgressStatus(ctx context.Context, nodeID string, userID string) error {
	if m.UpdateNodeProgressStatusFunc != nil {
		return m.UpdateNodeProgressStatusFunc(ctx, nodeID, userID)
	}
	return nil
}
func (m *Repopository) UpdateNodeProgressCompletion(ctx context.Context, nodeID string, userID string) error {
	if m.UpdateNodeProgressCompletionFunc != nil {
		return m.UpdateNodeProgressCompletionFunc(ctx, nodeID, userID)
	}
	return nil
}

func (m *Repopository) UpdateNodeProgress(ctx context.Context, nodeID string, userID string, status string) error {
	if m.UpdateNodeProgressFunc != nil {
		return m.UpdateNodeProgressFunc(ctx, nodeID, userID, status)
	}
	return nil
}

// Implement RepositoryComment
func (m *Repopository) CreateComment(ctx context.Context, req model.CreateCommentRequest) (string, error) {
	if m.CreateCommentFunc != nil {
		return m.CreateCommentFunc(ctx, req)
	}
	return "", nil
}
func (m *Repopository) GetCommentsByNodeID(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
	if m.GetCommentsByNodeIDFunc != nil {
		return m.GetCommentsByNodeIDFunc(ctx, nodeID)
	}
	return nil, nil
}
func (m *Repopository) GetCommentsByPathID(ctx context.Context, pathID string) ([]model.NodeComment, error) {
	if m.GetCommentsByPathIDFunc != nil {
		return m.GetCommentsByPathIDFunc(ctx, pathID)
	}
	return nil, nil
}
func (m *Repopository) DeleteComment(ctx context.Context, commentID, userID string) error {
	if m.DeleteCommentFunc != nil {
		return m.DeleteCommentFunc(ctx, commentID, userID)
	}
	return nil
}
func (m *Repopository) ToggleReaction(ctx context.Context, req model.CreateReactionRequest) (bool, error) {
	if m.ToggleReactionFunc != nil {
		return m.ToggleReactionFunc(ctx, req)
	}
	return false, nil
}
func (m *Repopository) GetReactionsByCommentID(ctx context.Context, commentID string) ([]model.CommentReaction, error) {
	if m.GetReactionsByCommentIDFunc != nil {
		return m.GetReactionsByCommentIDFunc(ctx, commentID)
	}
	return nil, nil
}
func (m *Repopository) CreateMention(ctx context.Context, req model.CreateMentionRequest) (string, error) {
	if m.CreateMentionFunc != nil {
		return m.CreateMentionFunc(ctx, req)
	}
	return "", nil
}
func (m *Repopository) GetCommentOwner(ctx context.Context, commentID string) (string, error) {
	if m.GetCommentOwnerFunc != nil {
		return m.GetCommentOwnerFunc(ctx, commentID)
	}
	return "", nil
}
func (m *Repopository) UpdateComment(ctx context.Context, userID, messageID, message string) (bool, error) {
	if m.UpdateCommentFunc != nil {
		return m.UpdateCommentFunc(ctx, userID, messageID, message)
	}
	return false, nil
}

// Implement RepositoryQuiz
func (m *Repopository) CreateQuestion(ctx context.Context, req model.CreateQuestionRequest) (string, error) {
	if m.CreateQuestionFunc != nil {
		return m.CreateQuestionFunc(ctx, req)
	}
	return "", nil
}
func (m *Repopository) GetQuestionsByNodeID(ctx context.Context, nodeID string) ([]model.NodeQuestion, error) {
	if m.GetQuestionsByNodeIDFunc != nil {
		return m.GetQuestionsByNodeIDFunc(ctx, nodeID)
	}
	return nil, nil
}
func (m *Repopository) DeleteQuestion(ctx context.Context, questionID string) error {
	if m.DeleteQuestionFunc != nil {
		return m.DeleteQuestionFunc(ctx, questionID)
	}
	return nil
}
func (m *Repopository) CreateChoice(ctx context.Context, req model.CreateChoiceRequest) (string, error) {
	if m.CreateChoiceFunc != nil {
		return m.CreateChoiceFunc(ctx, req)
	}
	return "", nil
}
func (m *Repopository) GetChoicesByQuestionID(ctx context.Context, questionID string) ([]model.QuestionChoice, error) {
	if m.GetChoicesByQuestionIDFunc != nil {
		return m.GetChoicesByQuestionIDFunc(ctx, questionID)
	}
	return nil, nil
}
func (m *Repopository) DeleteChoice(ctx context.Context, choiceID string) error {
	if m.DeleteChoiceFunc != nil {
		return m.DeleteChoiceFunc(ctx, choiceID)
	}
	return nil
}

func (m *Repopository) GetQuestionByID(ctx context.Context, questionID string) (*model.NodeQuestion, error) {
	if m.GetQuestionByIDFunc != nil {
		return m.GetQuestionByIDFunc(ctx, questionID)
	}
	return nil, nil
}
func (m *Repopository) UpdateQuestion(ctx context.Context, questionID string, req model.UpdateQuestionRequest) error {
	if m.UpdateQuestionFunc != nil {
		return m.UpdateQuestionFunc(ctx, questionID, req)
	}
	return nil
}
func (m *Repopository) GetChoiceByID(ctx context.Context, choiceID string) (*model.QuestionChoice, error) {
	if m.GetChoiceByIDFunc != nil {
		return m.GetChoiceByIDFunc(ctx, choiceID)
	}
	return nil, nil
}
func (m *Repopository) UpdateChoice(ctx context.Context, choiceID string, req model.UpdateChoiceRequest) error {
	if m.UpdateChoiceFunc != nil {
		return m.UpdateChoiceFunc(ctx, choiceID, req)
	}
	return nil
}

// Implement RepositoryRating
func (m *Repopository) UpsertLearningPathRating(ctx context.Context, rating *model.LearningPathRating) error {
	if m.UpsertLearningPathRatingFunc != nil {
		return m.UpsertLearningPathRatingFunc(ctx, rating)
	}
	return nil
}
func (m *Repopository) GetRatingByUser(ctx context.Context, pathID string, userID string) (*model.LearningPathRating, error) {
	if m.GetRatingByUserFunc != nil {
		return m.GetRatingByUserFunc(ctx, pathID, userID)
	}
	return nil, nil
}
func (m *Repopository) DeleteLearningPathRating(ctx context.Context, pathID string, userID string) error {
	if m.DeleteLearningPathRatingFunc != nil {
		return m.DeleteLearningPathRatingFunc(ctx, pathID, userID)
	}
	return nil
}

// Implement RepositoryHistory
func (m *Repopository) GetHistoryByUserID(ctx context.Context, userID string) ([]model.HistoryResponse, error) {
	if m.GetHistoryByUserIDFunc != nil {
		return m.GetHistoryByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

// Implement RepositoryResume
func (m *Repopository) GetNextNodeID(ctx context.Context, userID string, pathID string) (string, error) {
	if m.GetNextNodeIDFunc != nil {
		return m.GetNextNodeIDFunc(ctx, userID, pathID)
	}
	return "", nil
}

// Implement RepositoryXP
func (m *Repopository) AddXPAndRecalcLevel(ctx context.Context, userID string, xpAmount int) error {
	if m.AddXPAndRecalcLevelFunc != nil {
		return m.AddXPAndRecalcLevelFunc(ctx, userID, xpAmount)
	}
	return nil
}
