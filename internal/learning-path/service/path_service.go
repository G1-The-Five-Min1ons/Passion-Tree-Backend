package service

import (
	"database/sql"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) GetPaths() ([]model.LearningPath, error) {
	paths, err := s.pathRepo.GetAllLearnningPath()
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return paths, nil
}

func (s *serviceImpl) GetPathDetails(id string) (*model.LearningPath, error) {
	path, err := s.pathRepo.GetLearnningPathByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("learning path with id '%s' not found", id)
		}
		return nil, apperror.NewInternal(err)
	}
	return path, nil
}

func (s *serviceImpl) CreatePath(req model.CreatePathRequest) (string, error) {
	if req.Title == "" {
		return "", apperror.NewBadRequest("title cannot be empty")
	}
	id, err := s.pathRepo.CreateLearnningPath(req)
	if err != nil {
		return "", apperror.NewInternal(err)
	}
	return id, nil
}

func (s *serviceImpl) UpdatePath(id string, req model.UpdatePathRequest) error {
	if _, err := s.pathRepo.GetLearnningPathByID(id); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot update: path id '%s' not found", id)
		}
		return apperror.NewInternal(err)
	}

	if err := s.pathRepo.UpdateLearnningPath(id, req); err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) DeletePath(id string) error {
	if err := s.pathRepo.DeleteLearnningPath(id); err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) StartPath(pathID string, userID string) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}
	if err := s.pathRepo.EnrollLearnningPathUser(pathID, userID); err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) GetEnrollmentStatus(pathID string, userID string) (*model.PathEnroll, error) {
	enroll, err := s.pathRepo.GetLearnningPathEnrollmentStatus(pathID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("enrollment not found for user '%s'", userID)
		}
		return nil, apperror.NewInternal(err)
	}
	return enroll, nil
}