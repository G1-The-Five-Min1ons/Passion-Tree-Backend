package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"passiontree/internal/learning-path/model"
)

func (r *repositoryImpl) GetAllLearnningPath() ([]model.LearningPath, error) {
	query := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, status, create_at, update_at, IFNULL(creator_ID, '')
		FROM learning_path`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("repo.GetAllLearnningPath query failed: %w", err)
	}
	defer rows.Close()

	var paths []model.LearningPath
	for rows.Next() {
		var p model.LearningPath
		if err := rows.Scan(&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.CreatorID); err != nil {
			return nil, fmt.Errorf("repo.GetAllLearnningPath scan failed: %w", err)
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func (r *repositoryImpl) GetLearnningPathByID(id string) (*model.LearningPath, error) {
	pathQuery := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, status, create_at, update_at, IFNULL(creator_ID, '')
		FROM learning_path 
		WHERE path_id = ?`

	var p model.LearningPath
	err := r.db.QueryRow(pathQuery, id).Scan(
		&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.CreatorID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetLearnningPathByID scan failed: %w", err)
	}

	nodes, err := r.GetNodesByPathID(id)
	if err != nil {
		return nil, fmt.Errorf("repo.GetLearnningPathByID fetch nodes failed: %w", err)
	}

	for i := range nodes {
		materials, err := r.GetMaterialsByNodeID(nodes[i].NodeID)
		if err != nil {
			return nil, fmt.Errorf("repo.GetLearnningPathByID fetch materials failed (node=%s): %w", nodes[i].NodeID, err)
		}
		nodes[i].Materials = materials
	}

	p.Nodes = nodes
	return &p, nil
}

func (r *repositoryImpl) CreateLearnningPath(req model.CreatePathRequest) (string, error) {
	newID := uuid.New().String()
	now := time.Now()
	query := `INSERT INTO learning_path (path_id, title, objective, description, cover_img_url, avg_rating, status, creator_ID, create_at, update_at) VALUES (?, ?, ?, ?, ?, 0.0, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, newID, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Status, req.CreatorID, now, now)
	if err != nil {
		return "", fmt.Errorf("repo.CreateLearnningPath exec failed: %w", err)
	}
	return newID, nil
}

func (r *repositoryImpl) UpdateLearnningPath(id string, req model.UpdatePathRequest) error {
	query := `UPDATE learning_path SET title=?, objective=?, description=?, cover_img_url=?, status=?, update_at=? WHERE path_id=?`
	_, err := r.db.Exec(query, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("repo.UpdateLearnningPath failed [id=%s]: %w", id, err)
	}
	return nil
}

func (r *repositoryImpl) DeleteLearnningPath(id string) error {
	_, err := r.db.Exec("DELETE FROM learning_path WHERE path_id = ?", id)
	if err != nil {
		return fmt.Errorf("repo.DeleteLearnningPath failed [id=%s]: %w", id, err)
	}
	return nil
}

func (r *repositoryImpl) EnrollLearnningPathUser(pathID string, userID string) error {
	enrollID := uuid.New().String()
	now := time.Now()
	query := `INSERT INTO path_enroll (enroll_id, user_id, path_id, status, enroll_at) VALUES (?, ?, ?, 'active', ?)`
	_, err := r.db.Exec(query, enrollID, userID, pathID, now)
	if err != nil {
		return fmt.Errorf("repo.EnrollLearnningPathUser failed: %w", err)
	}
	return nil
}

func (r *repositoryImpl) GetLearnningPathEnrollmentStatus(pathID string, userID string) (*model.PathEnroll, error) {
	query := `SELECT enroll_id, status, enroll_at, complete_at FROM path_enroll WHERE user_id = ? AND path_id = ?`
	var pe model.PathEnroll
	err := r.db.QueryRow(query, userID, pathID).Scan(&pe.EnrollID, &pe.Status, &pe.EnrollAt, &pe.CompleteAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetEnrollmentStatus failed: %w", err)
	}
	return &pe, nil
}