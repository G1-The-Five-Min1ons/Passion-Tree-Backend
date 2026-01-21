package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"passiontree/internal/learning-path/model"
)

func (r *repositoryImpl) CreateQuestion(ctx context.Context, req model.CreateQuestionRequest) (string, error) {
	question_id := uuid.New().String()
	query := `INSERT INTO node_question (question_id, question_text, type, node_id) VALUES (@p1, @p2, @p3, @p4)`
	_, err := r.db.ExecContext(ctx, query, question_id, req.QuestionText, req.Type, req.NodeID)
	if err != nil {
		return "", fmt.Errorf("repo.CreateQuestion exec failed: %w", err)
	}
	return question_id, nil
}

func (r *repositoryImpl) GetQuestionsByNodeID(ctx context.Context, nodeID string) ([]model.NodeQuestion, error) {
	query := `SELECT question_id, question_text, type, node_id FROM node_question WHERE node_id = @p1`
	rows, err := r.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetQuestionsByNodeID query failed: %w", err)
	}
	defer rows.Close()

	var questions []model.NodeQuestion
	for rows.Next() {
		var q model.NodeQuestion
		if err := rows.Scan(&q.QuestionID, &q.QuestionText, &q.Type, &q.NodeID); err != nil {
			return nil, fmt.Errorf("repo.GetQuestionsByNodeID scan failed: %w", err)
		}
		
		choices, err := r.GetChoicesByQuestionID(ctx, q.QuestionID)
		if err != nil {
			return nil, fmt.Errorf("repo.GetQuestionsByNodeID fetch choices failed: %w", err)
		}
		q.Choices = choices
		
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetQuestionsByNodeID row iteration failed: %w", err)
	}

	return questions, nil
}

func (r *repositoryImpl) DeleteQuestion(ctx context.Context, questionID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM node_question WHERE question_id = @p1`, questionID)
	if err != nil {
		return fmt.Errorf("repo.DeleteQuestion exec failed [id=%s]: %w", questionID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) CreateChoice(ctx context.Context, req model.CreateChoiceRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO question_choice (choice_id, choice_text, is_correct, reasoning, question_id) VALUES (@p1, @p2, @p3, @p4, @p5)`
	_, err := r.db.ExecContext(ctx, query, id, req.ChoiceText, req.IsCorrect, req.Reasoning, req.QuestionID)
	if err != nil {
		return "", fmt.Errorf("repo.CreateChoice exec failed: %w", err)
	}
	return id, nil
}

func (r *repositoryImpl) GetChoicesByQuestionID(ctx context.Context, questionID string) ([]model.QuestionChoice, error) {
	query := `SELECT choice_id, choice_text, is_correct, reasoning, question_id FROM question_choice WHERE question_id = @p1`
	rows, err := r.db.QueryContext(ctx, query, questionID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetChoicesByQuestionID query failed: %w", err)
	}
	defer rows.Close()

	var choices []model.QuestionChoice
	for rows.Next() {
		var c model.QuestionChoice
		if err := rows.Scan(&c.ChoiceID, &c.ChoiceText, &c.IsCorrect, &c.Reasoning, &c.QuestionID); err != nil {
			return nil, fmt.Errorf("repo.GetChoicesByQuestionID scan failed: %w", err)
		}
		choices = append(choices, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetChoicesByQuestionID row iteration failed: %w", err)
	}

	return choices, nil
}

func (r *repositoryImpl) DeleteChoice(ctx context.Context, choiceID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM question_choice WHERE choice_id = @p1`, choiceID)
	if err != nil {
		return fmt.Errorf("repo.DeleteChoice exec failed [id=%s]: %w", choiceID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}