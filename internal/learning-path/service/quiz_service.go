package service

import (
	"passiontree/internal/learning-path/model"
)

func (s *service) AddQuestion(req model.CreateQuestionRequest) (string, error) {
	return s.repo.CreateQuestion(req)
}

func (s *service) GetQuestions(nodeID string) ([]model.NodeQuestion, error) {
	return s.repo.GetQuestionsByNodeID(nodeID)
}

func (s *service) RemoveQuestion(questionID string) error {
	return s.repo.DeleteQuestion(questionID)
}

func (s *service) AddChoice(req model.CreateChoiceRequest) (string, error) {
	return s.repo.CreateChoice(req)
}

func (s *service) RemoveChoice(choiceID string) error {
	return s.repo.DeleteChoice(choiceID)
}