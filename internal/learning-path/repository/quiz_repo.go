package repository

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"passiontree/internal/learning-path/model"
)

func (r *repositoryImpl) CreateQuestion(req model.CreateQuestionRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node_question (question_id, question_text, type, node_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.QuestionText, req.Type, req.NodeID)
	if err != nil {
		return "", fmt.Errorf("repo.CreateQuestion exec failed: %w", err)
	}
	return id, nil
}

func (r *repositoryImpl) GetQuestionsByNodeID(nodeID string) ([]model.NodeQuestion, error) {
	query := `SELECT question_id, question_text, type, node_id FROM node_question WHERE node_id = ?`
	rows, err := r.db.Query(query, nodeID)
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
		
		choices, err := r.GetChoicesByQuestionID(q.QuestionID)
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

func (r *repositoryImpl) DeleteQuestion(questionID string) error {
	res, err := r.db.Exec(`DELETE FROM node_question WHERE question_id = ?`, questionID)
	if err != nil {
		return fmt.Errorf("repo.DeleteQuestion exec failed [id=%s]: %w", questionID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) CreateChoice(req model.CreateChoiceRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO question_choice (choice_id, choice_text, is_correct, reasoning, node_id) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.ChoiceText, req.IsCorrect, req.Reasoning, req.QuestionID)
	if err != nil {
		return "", fmt.Errorf("repo.CreateChoice exec failed: %w", err)
	}
	return id, nil
}

func (r *repositoryImpl) GetChoicesByQuestionID(questionID string) ([]model.QuestionChoice, error) {
	query := `SELECT choice_id, choice_text, is_correct, reasoning, node_id FROM question_choice WHERE node_id = ?`
	rows, err := r.db.Query(query, questionID)
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

func (r *repositoryImpl) DeleteChoice(choiceID string) error {
	res, err := r.db.Exec(`DELETE FROM question_choice WHERE choice_id = ?`, choiceID)
	if err != nil {
		return fmt.Errorf("repo.DeleteChoice exec failed [id=%s]: %w", choiceID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}