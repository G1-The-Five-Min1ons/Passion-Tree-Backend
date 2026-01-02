package service

import (
	"errors"
	"passiontree/internal/learning-path/model"
)

func (s *service) GetPaths() ([]model.LearningPath, error) {
	return s.repo.GetAllLearnningPath()
}

func (s *service) GetPathDetails(id string) (*model.LearningPath, error) {
	return s.repo.GetLearnningPathByID(id)
}

func (s *service) CreatePath(req model.CreatePathRequest) (string, error) {
	if req.Title == "" {
		return "", errors.New("title is required")
	}
	return s.repo.CreateLearnningPath(req)
}

func (s *service) UpdatePath(id string, req model.UpdatePathRequest) error {
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

func (s *service) GetEnrollmentStatus(pathID string, userID string) (*model.PathEnroll, error) {
	return s.repo.GetLearnningPathEnrollmentStatus(pathID, userID)
}