package repository

import (
    "github.com/google/uuid"
    "passiontree/internal/learning-path/model"
)

func (r *repositoryImpl) CreateQuestion(req model.CreateQuestionRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node_question (question_id, question_text, type, node_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.QuestionText, req.Type, req.NodeID)
	return id, err
}

func (r *repositoryImpl) GetQuestionsByNodeID(nodeID string) ([]model.NodeQuestion, error) {
	query := `SELECT question_id, question_text, type, node_id FROM node_question WHERE node_id = ?`
	rows, err := r.db.Query(query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []model.NodeQuestion
	for rows.Next() {
		var q model.NodeQuestion
		if err := rows.Scan(&q.QuestionID, &q.QuestionText, &q.Type, &q.NodeID); err != nil {
			continue
		}
		choices, _ := r.GetChoicesByQuestionID(q.QuestionID)
		q.Choices = choices
		questions = append(questions, q)
	}
	return questions, nil
}

func (r *repositoryImpl) DeleteQuestion(questionID string) error {
	_, err := r.db.Exec(`DELETE FROM node_question WHERE question_id = ?`, questionID)
	return err
}

func (r *repositoryImpl) CreateChoice(req model.CreateChoiceRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO question_choice (choice_id, choice_text, is_correct, reasoning, node_id) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.ChoiceText, req.IsCorrect, req.Reasoning, req.QuestionID)
	return id, err
}

func (r *repositoryImpl) GetChoicesByQuestionID(questionID string) ([]model.QuestionChoice, error) {
	query := `SELECT choice_id, choice_text, is_correct, reasoning, node_id FROM question_choice WHERE node_id = ?`
	rows, err := r.db.Query(query, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var choices []model.QuestionChoice
	for rows.Next() {
		var c model.QuestionChoice
		if err := rows.Scan(&c.ChoiceID, &c.ChoiceText, &c.IsCorrect, &c.Reasoning, &c.QuestionID); err != nil {
			continue
		}
		choices = append(choices, c)
	}
	return choices, nil
}

func (r *repositoryImpl) DeleteChoice(choiceID string) error {
	_, err := r.db.Exec(`DELETE FROM question_choice WHERE choice_id = ?`, choiceID)
	return err
}