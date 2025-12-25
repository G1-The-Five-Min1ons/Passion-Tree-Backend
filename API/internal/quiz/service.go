package quiz

import "errors"

type Service interface {
	CreateQuiz(req CreateQuizRequest) (int, error)
	GetQuiz(quizID int) (*Quiz, error)
	UpdateQuiz(quizID int, req UpdateQuizRequest) error
	DeleteQuiz(quizID int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateQuiz(req CreateQuizRequest) (int, error) {
	if req.CorrectAnswer >= len(req.Options) {
		return 0, errors.New("correct_answer index is out of bounds")
	}
	return s.repo.CreateQuiz(req)
}

func (s *service) GetQuiz(quizID int) (*Quiz, error) {
	return s.repo.GetQuizByID(quizID)
}

func (s *service) UpdateQuiz(quizID int, req UpdateQuizRequest) error {
	if len(req.Options) > 0 && req.CorrectAnswer >= len(req.Options) {
		return errors.New("correct_answer index is out of bounds")
	}
	return s.repo.UpdateQuiz(quizID, req)
}

func (s *service) DeleteQuiz(quizID int) error {
	return s.repo.DeleteQuiz(quizID)
}