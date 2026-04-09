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
			s.logger.WarnContext(ctx, "foreign key violation: node not found", "node_id", req.NodeID)
			return "", apperror.NewBadRequest("invalid node_id: node does not exist")
		}
		s.logger.ErrorContext(ctx, "database error: failed to create question", "error", err, "node_id", req.NodeID)
		return "", apperror.NewInternal("failed to create question: %w", err)
	}

	s.logger.InfoContext(ctx, "quiz question added successfully", "question_id", id, "node_id", req.NodeID)
	return id, nil
}

func (s *serviceImpl) GetQuestions(ctx context.Context, nodeID string) ([]model.NodeQuestion, error) {
	 	
	if nodeID == "" {
		return nil, apperror.NewBadRequest("node_id is required")
	}

	questions, err := s.quizRepo.GetQuestionsByNodeID(ctx, nodeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "database error: failed to get questions", "error", err, "node_id", nodeID)
		return nil, apperror.NewInternal("failed to get questions for node '%s': %w", nodeID, err)
	}

	s.logger.InfoContext(ctx, "successfully retrieved questions", "node_id", nodeID, "count", len(questions))
	return questions, nil
}

func (s *serviceImpl) RemoveQuestion(ctx context.Context, questionID string) error {

	if questionID == "" {
		return apperror.NewBadRequest("question_id is required")
	}
	if err := s.quizRepo.DeleteQuestion(ctx, questionID); err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "question not found for deletion", "question_id", questionID)
			return apperror.NewNotFound("cannot delete: question id '%s' not found", questionID)
		}

		if apperror.IsForeignKeyError(err) {
			s.logger.WarnContext(ctx, "deletion blocked: question has associated choices", "question_id", questionID)
			return apperror.NewConflict("cannot delete question: there are existing choices associated with this question")
		}

		s.logger.ErrorContext(ctx, "database error: failed to delete question", "error", err, "question_id", questionID)
		return apperror.NewInternal("failed to delete question '%s': %w", questionID, err)
	}

	s.logger.InfoContext(ctx, "quiz question removed successfully", "question_id", questionID)
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
			s.logger.WarnContext(ctx, "foreign key violation: question not found", "question_id", req.QuestionID)
			return "", apperror.NewBadRequest("invalid question_id: question does not exist")
		}
		s.logger.ErrorContext(ctx, "database error: failed to create choice", "error", err, "question_id", req.QuestionID)
		return "", apperror.NewInternal("failed to create choice: %w", err)
	}
	
	s.logger.InfoContext(ctx, "choice added successfully", "choice_id", id, "question_id", req.QuestionID)
	return id, nil
}

func (s *serviceImpl) RemoveChoice(ctx context.Context, choiceID string) error {
	
	if choiceID == "" {
		return apperror.NewBadRequest("choice_id is required")
	}
	if err := s.quizRepo.DeleteChoice(ctx, choiceID); err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "choice not found for deletion", "choice_id", choiceID)
			return apperror.NewNotFound("cannot delete: choice id '%s' not found", choiceID)
		}

		s.logger.ErrorContext(ctx, "database error: failed to delete choice", "error", err, "choice_id", choiceID)
		return apperror.NewInternal("failed to delete choice '%s': %w", choiceID, err)
	}

	s.logger.InfoContext(ctx, "choice removed successfully", "choice_id", choiceID)
	return nil
}

func (s *serviceImpl) EditQuestion(ctx context.Context, questionID string, req model.UpdateQuestionRequest) error {
	if questionID == "" {
		return apperror.NewBadRequest("question_id is required")
	}
	if req.QuestionText == "" && req.Type == "" {
		return apperror.NewBadRequest("at least one field (question_text or type) is required for update")
	}

	existing, err := s.quizRepo.GetQuestionByID(ctx, questionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("question id '%s' not found", questionID)
		}
		return apperror.NewInternal("failed to retrieve question: %w", err)
	}

	if req.QuestionText == "" {
		req.QuestionText = existing.QuestionText
	}
	if req.Type == "" {
		req.Type = existing.Type
	}

	if err := s.quizRepo.UpdateQuestion(ctx, questionID, req); err != nil {
		return apperror.NewInternal("failed to update question: %w", err)
	}

	s.logger.InfoContext(ctx, "question updated successfully", "question_id", questionID)
	return nil
}

func (s *serviceImpl) EditChoice(ctx context.Context, choiceID string, req model.UpdateChoiceRequest) error {
	if choiceID == "" {
		return apperror.NewBadRequest("choice_id is required")
	}
	if req.ChoiceText == nil && req.IsCorrect == nil && req.Reasoning == nil {
		return apperror.NewBadRequest("at least one field (choice_text, is_correct, reasoning) is required for update")
	}

	existing, err := s.quizRepo.GetChoiceByID(ctx, choiceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("choice id '%s' not found", choiceID)
		}
		return apperror.NewInternal("failed to retrieve choice: %w", err)
	}

	if req.ChoiceText == nil {
		req.ChoiceText = &existing.ChoiceText
	}
	if req.IsCorrect == nil {
		req.IsCorrect = &existing.IsCorrect
	}
	if req.Reasoning == nil {
		req.Reasoning = &existing.Reasoning
	}

	if err := s.quizRepo.UpdateChoice(ctx, choiceID, req); err != nil {
		return apperror.NewInternal("failed to update choice: %w", err)
	}

	s.logger.InfoContext(ctx, "choice updated successfully", "choice_id", choiceID)
	return nil
}
