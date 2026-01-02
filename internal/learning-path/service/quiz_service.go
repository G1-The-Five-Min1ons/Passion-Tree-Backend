package service

import (
	"passiontree/internal/learning-path/model"
)

func (s *service) AddQuestion(req model.CreateQuestionRequest) (string, error) {
	return s.quizRepo.CreateQuestion(req)
}

func (s *service) GetQuestions(nodeID string) ([]model.NodeQuestion, error) {
	return s.quizRepo.GetQuestionsByNodeID(nodeID)
}

func (s *service) RemoveQuestion(questionID string) error {
	return s.quizRepo.DeleteQuestion(questionID)
}

func (s *service) AddChoice(req model.CreateChoiceRequest) (string, error) {
	return s.quizRepo.CreateChoice(req)
}

func (s *service) RemoveChoice(choiceID string) error {
	return s.quizRepo.DeleteChoice(choiceID)
}