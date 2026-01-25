package service

import (
	"context"
	"database/sql"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) AddQuestion(ctx context.Context, req model.CreateQuestionRequest) (string, error) {
	if req.QuestionText == "" {
		return "", apperror.NewBadRequest("question text is required")
	}
	if req.Type == "" {
		return "", apperror.NewBadRequest("question type is required")
	}

	id, err := s.quizRepo.CreateQuestion(ctx, req)
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

func (s *serviceImpl) GetQuestions(ctx context.Context, nodeID string) ([]model.NodeQuestion, error) {
	if nodeID == "" {
		return nil, apperror.NewBadRequest("node_id is required")
	}
	questions, err := s.quizRepo.GetQuestionsByNodeID(ctx, nodeID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return questions, nil
}

func (s *serviceImpl) RemoveQuestion(ctx context.Context, questionID string) error {
	if questionID == "" {
		return apperror.NewBadRequest("question_id is required")
	}
	if err := s.quizRepo.DeleteQuestion(ctx, questionID); err != nil {
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

func (s *serviceImpl) AddChoice(ctx context.Context, req model.CreateChoiceRequest) (string, error) {
	if req.ChoiceText == "" {
		return "", apperror.NewBadRequest("choice text is required")
	}

	id, err := s.quizRepo.CreateChoice(ctx, req)
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

func (s *serviceImpl) RemoveChoice(ctx context.Context, choiceID string) error {
	if choiceID == "" {
		return apperror.NewBadRequest("choice_id is required")
	}
	if err := s.quizRepo.DeleteChoice(ctx, choiceID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot delete: choice id '%s' not found", choiceID)
		}
		return apperror.NewInternal(err)
	}
	return nil
}
