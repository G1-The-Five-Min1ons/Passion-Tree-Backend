package learningpath

import (
	"errors"
)

type Service interface {
	GetPaths() ([]LearningPath, error)
	GetPathDetails(id string) (*LearningPath, error)
	CreatePath(req CreatePathRequest) (string, error)
	UpdatePath(id string, req UpdatePathRequest) error
	DeletePath(id string) error

	StartPath(pathID string, userID string) error
	GetEnrollmentStatus(pathID string, userID string) (*PathEnroll, error)

	AddNode(req CreateNodeRequest) (string, error)
	EditNode(nodeID string, req UpdateNodeRequest) error
	RemoveNode(nodeID string) error

	AddMaterial(req CreateMaterialRequest) (string, error)
	RemoveMaterial(materialID string) error

	AddComment(req CreateCommentRequest) (string, error)
	GetNodeComments(nodeID string) ([]NodeComment, error)
	RemoveComment(commentID string) error
	AddReaction(req CreateReactionRequest) error
	AddMention(req CreateMentionRequest) (string, error)

	AddQuestion(req CreateQuestionRequest) (string, error)
	GetQuestions(nodeID string) ([]NodeQuestion, error)
	RemoveQuestion(questionID string) error
	AddChoice(req CreateChoiceRequest) (string, error)
	RemoveChoice(choiceID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetPaths() ([]LearningPath, error) {
	return s.repo.GetAllLearnningPath()
}

func (s *service) GetPathDetails(id string) (*LearningPath, error) {
	return s.repo.GetLearnningPathByID(id)
}

func (s *service) CreatePath(req CreatePathRequest) (string, error) {
	if req.Title == "" {
		return "", errors.New("title is required")
	}
	return s.repo.CreateLearnningPath(req)
}

func (s *service) UpdatePath(id string, req UpdatePathRequest) error {
	return s.repo.UpdateLearnningPath(id, req)
}

func (s *service) DeletePath(id string) error {
	return s.repo.DeleteLearnningPath(id)
}

func (s *service) StartPath(pathID string, userID string) error {
	if userID == "" {
		return errors.New("invalid user id")
	}
	return s.repo.EnrollLearnningPathUser(pathID, userID)
}

func (s *service) GetEnrollmentStatus(pathID string, userID string) (*PathEnroll, error) {
	return s.repo.GetLearnningPathEnrollmentStatus(pathID, userID)
}

func (s *service) AddNode(req CreateNodeRequest) (string, error) {
	return s.repo.CreateNode(req)
}

func (s *service) EditNode(nodeID string, req UpdateNodeRequest) error {
	return s.repo.UpdateNode(nodeID, req)
}

func (s *service) RemoveNode(nodeID string) error {
	return s.repo.DeleteNode(nodeID)
}

func (s *service) AddMaterial(req CreateMaterialRequest) (string, error) {
	return s.repo.CreateMaterial(req)
}

func (s *service) RemoveMaterial(materialID string) error {
	return s.repo.DeleteMaterial(materialID)
}

func (s *service) AddComment(req CreateCommentRequest) (string, error) {
	return s.repo.CreateComment(req)
}

func (s *service) GetNodeComments(nodeID string) ([]NodeComment, error) {
	return s.repo.GetCommentsByNodeID(nodeID)
}

func (s *service) RemoveComment(commentID string) error {
	return s.repo.DeleteComment(commentID)
}

func (s *service) AddReaction(req CreateReactionRequest) error {
	return s.repo.CreateReaction(req)
}

func (s *service) AddMention(req CreateMentionRequest) (string, error) {
	return s.repo.CreateMention(req)
}

func (s *service) AddQuestion(req CreateQuestionRequest) (string, error) {
	return s.repo.CreateQuestion(req)
}

func (s *service) GetQuestions(nodeID string) ([]NodeQuestion, error) {
	return s.repo.GetQuestionsByNodeID(nodeID)
}

func (s *service) RemoveQuestion(questionID string) error {
	return s.repo.DeleteQuestion(questionID)
}

func (s *service) AddChoice(req CreateChoiceRequest) (string, error) {
	return s.repo.CreateChoice(req)
}

func (s *service) RemoveChoice(choiceID string) error {
	return s.repo.DeleteChoice(choiceID)
}