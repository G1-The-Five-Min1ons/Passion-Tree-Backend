package service

import (
	"database/sql"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) AddQuestion(req model.CreateQuestionRequest) (string, error) {
	if req.QuestionText == "" {
		return "", apperror.NewBadRequest("question text is required")
	}
	if req.Type == "" {
		return "", apperror.NewBadRequest("question type is required")
	}

	id, err := s.quizRepo.CreateQuestion(req)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return "", apperror.NewConflict("question with this ID already exists")
		}
		if apperror.IsForeignKeyError(err) {
			return "", apperror.NewBadRequest("invalid node_id: node does not exist")
		}
		return "", apperror.NewInternal(err)
	}
	return id, nil
}

func (s *serviceImpl) GetQuestions(nodeID string) ([]model.NodeQuestion, error) {
	if nodeID == "" {
		return nil, apperror.NewBadRequest("node_id is required")
	}
	questions, err := s.quizRepo.GetQuestionsByNodeID(nodeID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return questions, nil
}

func (s *serviceImpl) RemoveQuestion(questionID string) error {
	if questionID == "" {
		return apperror.NewBadRequest("question_id is required")
	}
	if err := s.quizRepo.DeleteQuestion(questionID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot delete: question id '%s' not found", questionID)
		}
		if apperror.IsForeignKeyError(err) {
			return apperror.NewConflict("cannot delete question: there are existing choices associated with this question")
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) AddChoice(req model.CreateChoiceRequest) (string, error) {
	if req.ChoiceText == "" {
		return "", apperror.NewBadRequest("choice text is required")
	}

	id, err := s.quizRepo.CreateChoice(req)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return "", apperror.NewConflict("choice with this ID already exists")
		}
		if apperror.IsForeignKeyError(err) {
			return "", apperror.NewBadRequest("invalid question_id: question does not exist")
		}
		return "", apperror.NewInternal(err)
	}
	return id, nil
}

func (s *serviceImpl) RemoveChoice(choiceID string) error {
	if choiceID == "" {
		return apperror.NewBadRequest("choice_id is required")
	}
	if err := s.quizRepo.DeleteChoice(choiceID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot delete: choice id '%s' not found", choiceID)
		}
		return apperror.NewInternal(err)
	}
	return nil
}
