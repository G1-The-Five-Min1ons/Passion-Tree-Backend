package service

import (
	"passiontree/internal/learning-path/model"
)

func (s *serviceImpl) AddQuestion(req model.CreateQuestionRequest) (string, error) {
	return s.quizRepo.CreateQuestion(req)
}

func (s *serviceImpl) GetQuestions(nodeID string) ([]model.NodeQuestion, error) {
	return s.quizRepo.GetQuestionsByNodeID(nodeID)
}

func (s *serviceImpl) RemoveQuestion(questionID string) error {
	return s.quizRepo.DeleteQuestion(questionID)
}

func (s *serviceImpl) AddChoice(req model.CreateChoiceRequest) (string, error) {
	return s.quizRepo.CreateChoice(req)
}

func (s *serviceImpl) RemoveChoice(choiceID string) error {
	return s.quizRepo.DeleteChoice(choiceID)
}