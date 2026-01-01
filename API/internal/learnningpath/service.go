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
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetPaths() ([]LearningPath, error) {
	return s.repo.GetAll()
}

func (s *service) GetPathDetails(id string) (*LearningPath, error) {
	return s.repo.GetByID(id)
}

func (s *service) CreatePath(req CreatePathRequest) (string, error) {
	if req.Title == "" {
		return "", errors.New("title is required")
	}
	return s.repo.Create(req)
}

func (s *service) UpdatePath(id string, req UpdatePathRequest) error {
	return s.repo.Update(id, req)
}

func (s *service) DeletePath(id string) error {
	return s.repo.Delete(id)
}

func (s *service) StartPath(pathID string, userID string) error {
	if userID == "" {
		return errors.New("invalid user id")
	}
	return s.repo.EnrollUser(pathID, userID)
}

func (s *service) GetEnrollmentStatus(pathID string, userID string) (*PathEnroll, error) {
	return s.repo.GetEnrollmentStatus(pathID, userID)
}