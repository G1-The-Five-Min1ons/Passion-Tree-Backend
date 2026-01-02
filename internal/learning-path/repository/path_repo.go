package repository

import (
    "time"
    "github.com/google/uuid"
    "passiontree/internal/learning-path/model"
)

func (r *repository) GetAllLearnningPath() ([]model.LearningPath, error) {
	query := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, status, create_at, update_at, IFNULL(creator_ID, '')
		FROM learning_path`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []model.LearningPath
	for rows.Next() {
		var p model.LearningPath
		if err := rows.Scan(&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.CreatorID); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func (r *repository) GetLearnningPathByID(id string) (*model.LearningPath, error) {
	pathQuery := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, status, create_at, update_at, IFNULL(creator_ID, '')
		FROM learning_path 
		WHERE path_id = ?`

	var p model.LearningPath
	err := r.db.QueryRow(pathQuery, id).Scan(
		&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.CreatorID,
	)
	if err != nil {
		return nil, err
	}

	nodes, err := r.GetNodesByPathID(id)
	if err != nil {
		return &p, nil
	}

	for i := range nodes {
		materials, err := r.GetMaterialsByNodeID(nodes[i].NodeID)
		if err == nil {
			nodes[i].Materials = materials
		}
	}

	p.Nodes = nodes
	return &p, nil
}

func (r *repository) CreateLearnningPath(req model.CreatePathRequest) (string, error) {
	newID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO learning_path (path_id, title, objective, description, cover_img_url, avg_rating, status, creator_ID, create_at, update_at)
		VALUES (?, ?, ?, ?, ?, 0.0, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, newID, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Status, req.CreatorID, now, now)
	if err != nil {
		return "", err
	}
	return newID, nil
}

func (r *repository) UpdateLearnningPath(id string, req model.UpdatePathRequest) error {
	query := `
		UPDATE learning_path 
		SET title=?, objective=?, description=?, cover_img_url=?, status=?, update_at=? 
		WHERE path_id=?`

	_, err := r.db.Exec(query, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Status, time.Now(), id)
	return err
}

func (r *repository) DeleteLearnningPath(id string) error {
	_, err := r.db.Exec("DELETE FROM learning_path WHERE path_id = ?", id)
	return err
}

func (r *repository) EnrollLearnningPathUser(pathID string, userID string) error {
	enrollID := uuid.New().String()
	now := time.Now()
	query := `
		INSERT INTO path_enroll (enroll_id, user_id, path_id, status, enroll_at)
		VALUES (?, ?, ?, 'active', ?)`

	_, err := r.db.Exec(query, enrollID, userID, pathID, now)
	return err
}

func (r *repository) GetLearnningPathEnrollmentStatus(pathID string, userID string) (*model.PathEnroll, error) {
	query := `
		SELECT enroll_id, status, enroll_at, complete_at 
		FROM path_enroll 
		WHERE user_id = ? AND path_id = ?`

	var pe model.PathEnroll
	err := r.db.QueryRow(query, userID, pathID).Scan(&pe.EnrollID, &pe.Status, &pe.EnrollAt, &pe.CompleteAt)
	if err != nil {
		return nil, err
	}
	return &pe, nil
}