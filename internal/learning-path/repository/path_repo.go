package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"passiontree/internal/learning-path/model"

	"github.com/google/uuid"
)

func (r *repositoryImpl) GetAllLearnningPath(ctx context.Context) ([]model.LearningPath, error) {
	query := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, publish_status, create_at, update_at, creator_id
		FROM learning_path`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repo.GetAllLearnningPath query failed: %w", err)
	}
	defer rows.Close()

	var paths []model.LearningPath
	for rows.Next() {
		var p model.LearningPath
		if err := rows.Scan(&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Publish_status, &p.CreatedAt, &p.UpdatedAt, &p.CreatorID); err != nil {
			return nil, fmt.Errorf("repo.GetAllLearnningPath scan failed: %w", err)
		}
		paths = append(paths, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetAllLearnningPath row iteration failed: %w", err)
	}

	return paths, nil
}

func (r *repositoryImpl) GetLearnningPathByID(ctx context.Context, path_id string) (*model.LearningPath, error) {
	pathQuery := `
		SELECT path_id, title, cover_img_url, objective, description, avg_rating, publish_status, create_at, update_at, creator_id
		FROM learning_path 
		WHERE path_id = @p1`

	var p model.LearningPath
	err := r.db.QueryRowContext(ctx, pathQuery, path_id).Scan(
		&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.Description, &p.AvgRating, &p.Publish_status, &p.CreatedAt, &p.UpdatedAt, &p.CreatorID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetLearnningPathByID scan failed: %w", err)
	}

	return &p, nil
}

func (r *repositoryImpl) CreateLearnningPath(ctx context.Context, req model.CreatePathRequest) (string, error) {
	newID := uuid.New().String()
	now := time.Now()
	query := `INSERT INTO learning_path (path_id, title, objective, description, cover_img_url, avg_rating, publish_status, creator_ID, create_at, update_at) VALUES (@p1, @p2, @p3, @p4, @p5, 0.0, @p6, @p7, @p8, @p9)`

	_, err := r.db.ExecContext(ctx, query, newID, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Publish_status, req.CreatorID, now, now)
	if err != nil {
		return "", fmt.Errorf("repo.CreateLearnningPath exec failed: %w", err)
	}
	return newID, nil
}

func (r *repositoryImpl) UpdateLearnningPath(ctx context.Context, path_id string, req model.UpdatePathRequest) error {
	query := `UPDATE learning_path SET title=@p1, objective=@p2, description=@p3, cover_img_url=@p4, publish_status=@p5, update_at=@p6 WHERE path_id=@p7`
	_, err := r.db.ExecContext(ctx, query, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Publish_status, time.Now(), path_id)
	if err != nil {
		return fmt.Errorf("repo.UpdateLearnningPath failed [id=%s]: %w", path_id, err)
	}
	return nil
}

func (r *repositoryImpl) DeleteLearnningPath(ctx context.Context, path_id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM learning_path WHERE path_id = @p1", path_id)
	if err != nil {
		return fmt.Errorf("repo.DeleteLearnningPath failed [id=%s]: %w", path_id, err)
	}
	return nil
}

func (r *repositoryImpl) EnrollLearnningPathUser(ctx context.Context, pathID string, userID string) error {
	enrollID := uuid.New().String()
	now := time.Now()
	query := `INSERT INTO path_enroll (enroll_id, user_id, path_id, enrollment_status, enroll_at) VALUES (@p1, @p2, @p3, 'active', @p4)`
	_, err := r.db.ExecContext(ctx, query, enrollID, userID, pathID, now)
	if err != nil {
		return fmt.Errorf("repo.EnrollLearnningPathUser failed: %w", err)
	}
	return nil
}

func (r *repositoryImpl) GetLearnningPathEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error) {
	query := `SELECT enroll_id, enrollment_status, enroll_at, complete_at FROM path_enroll WHERE user_id = @p1 AND path_id = @p2`
	var pe model.PathEnroll
	err := r.db.QueryRowContext(ctx, query, userID, pathID).Scan(&pe.EnrollID, &pe.Enrollment_status, &pe.EnrollAt, &pe.CompleteAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetEnrollmentStatus failed: %w", err)
	}
	return &pe, nil
}
