package learningpath

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetAll() ([]LearningPath, error)
	GetByID(id string) (*LearningPath, error)
	Create(req CreatePathRequest) (string, error)
	Update(id string, req UpdatePathRequest) error
	Delete(id string) error
	EnrollUser(pathID string, userID string) error
	GetEnrollmentStatus(pathID string, userID string) (*PathEnroll, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll() ([]LearningPath, error) {
	query := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, status, create_at, update_at 
		FROM learning_path`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []LearningPath
	for rows.Next() {
		var p LearningPath
		if err := rows.Scan(&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func (r *repository) GetByID(id string) (*LearningPath, error) {
	pathQuery := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, status, create_at, update_at 
		FROM learning_path 
		WHERE path_id = ?`
	
	var p LearningPath
	err := r.db.QueryRow(pathQuery, id).Scan(
		&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	nodeQuery := `SELECT node_id, title, description FROM node WHERE path_id = ?`
	
	rows, err := r.db.Query(nodeQuery, id)
	if err != nil {
		return &p, nil 
	}
	defer rows.Close()

	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.NodeID, &n.Title, &n.Description); err != nil {
			continue
		}
		p.Nodes = append(p.Nodes, n)
	}

	return &p, nil
}

func (r *repository) Create(req CreatePathRequest) (string, error) {
	newID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO learning_path (path_id, title, objective, description, cover_img_url, avg_rating, status, create_at, update_at)
		VALUES (?, ?, ?, ?, ?, 0.0, ?, ?, ?)`

	_, err := r.db.Exec(query, newID, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Status, now, now)
	if err != nil {
		return "", err
	}
	return newID, nil
}

func (r *repository) Update(id string, req UpdatePathRequest) error {
	query := `
		UPDATE learning_path 
		SET title=?, objective=?, description=?, cover_img_url=?, status=?, update_at=? 
		WHERE path_id=?`
	
	_, err := r.db.Exec(query, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Status, time.Now(), id)
	return err
}

func (r *repository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM learning_path WHERE path_id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *repository) EnrollUser(pathID string, userID string) error {
	enrollID := uuid.New().String()
	now := time.Now()
	query := `
		INSERT INTO path_enroll (enroll_id, user_id, path_id, status, enroll_at)
		VALUES (?, ?, ?, 'active', ?)`

	_, err := r.db.Exec(query, enrollID, userID, pathID, now)
	return err
}

func (r *repository) GetEnrollmentStatus(pathID string, userID string) (*PathEnroll, error) {
	query := `
		SELECT enroll_id, status, enroll_at, complete_at 
		FROM path_enroll 
		WHERE user_id = ? AND path_id = ?`

	var pe PathEnroll
	err := r.db.QueryRow(query, userID, pathID).Scan(&pe.EnrollID, &pe.Status, &pe.EnrollAt, &pe.CompleteAt)
	if err != nil {
		return nil, err
	}
	return &pe, nil
}