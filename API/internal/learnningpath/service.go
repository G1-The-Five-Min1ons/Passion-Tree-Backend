package learningpath

import "errors"

type Service interface {
	GetPaths(category, search string) ([]LearningPath, error)
	GetPathDetails(id int) (*LearningPath, error)
	CreatePath(req CreatePathRequest) (int, error)
	UpdatePath(id int, req UpdatePathRequest) error
	DeletePath(id int) error
	StartPath(pathID int, userID int) error
	GetPathProgress(pathID int, userID int) ([]NodeProgress, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetPaths(category, search string) ([]LearningPath, error) {
	return s.repo.GetAll(category, search)
}

func (s *service) GetPathDetails(id int) (*LearningPath, error) {
	return s.repo.GetByID(id)
}

func (s *service) CreatePath(req CreatePathRequest) (int, error) {
	if req.Title == "" {
		return 0, errors.New("title is required")
	}
	return s.repo.Create(req)
}

func (s *service) UpdatePath(id int, req UpdatePathRequest) error {
	return s.repo.Update(id, req)
}

func (s *service) DeletePath(id int) error {
	return s.repo.Delete(id)
}

func (s *service) StartPath(pathID int, userID int) error {
	if userID == 0 {
		return errors.New("invalid user")
	}
	return s.repo.EnrollUser(pathID, userID)
}

func (s *service) GetPathProgress(pathID int, userID int) ([]NodeProgress, error) {
	return s.repo.GetUserProgress(pathID, userID)
}