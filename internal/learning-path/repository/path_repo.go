package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"passiontree/internal/learning-path/model"

	"github.com/google/uuid"
)

func (r *repositoryImpl) GetAllLearningPath(ctx context.Context) ([]model.LearningPaths, error) {
	query := `
		SELECT 
   			CONVERT(VARCHAR(36), lp.path_id) as id, 
    		lp.title, 
    		lp.description,
    		u.first_name as instructor,
    		ISNULL(pe_count.total_students, 0) as student,
			ISNULL(n_count.total_nodes, 0) as modules,
			lp.avg_rating,
    		lp.cover_img_url, 
    		lp.objective,
    		lp.publish_status, 
    		lp.create_at, 
    		lp.update_at
		FROM learning_path AS lp 
		JOIN users AS u ON lp.creator_id = u.user_id
		LEFT JOIN (
    		SELECT path_id, COUNT(node_id) as total_nodes 
    		FROM node 
    		GROUP BY path_id
		) AS n_count ON lp.path_id = n_count.path_id
		LEFT JOIN (
    		SELECT path_id, COUNT(enroll_id) as total_students 
    		FROM path_enroll 
    		GROUP BY path_id
		) AS pe_count ON lp.path_id = pe_count.path_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repo.GetAllLearningPath query failed: %w", err)
	}
	defer rows.Close()

	var paths []model.LearningPaths
	for rows.Next() {
		var p model.LearningPaths
		if err := rows.Scan(
			&p.PathID,
			&p.Title,
			&p.Description,
			&p.Instructor,
			&p.Students,
			&p.Modules,
			&p.Rating,
			&p.CoverImgURL,
			&p.Objective,
			&p.PublishStatus,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repo.GetAllLearningPath scan failed: %w", err)
		}
		paths = append(paths, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetAllLearnningPath row iteration failed: %w", err)
	}

	return paths, nil
}

func (r *repositoryImpl) GetLearningPathByID(ctx context.Context, path_id string) (*model.LearningPath, error) {
	pathQuery := `
        SELECT 
            CONVERT(VARCHAR(36), lp.path_id) as path_id, 
            lp.title, 
            lp.cover_img_url, 
            lp.objective, 
            lp.description, 
            lp.avg_rating, 
            lp.publish_status, 
            lp.create_at, 
            lp.update_at, 
            u.first_name as instructor,
            CONVERT(VARCHAR(36), lp.creator_id) as creator_id,
            ISNULL(n_count.total_nodes, 0) as modules,
            ISNULL(pe_count.total_students, 0) as student
        FROM learning_path AS lp 
        JOIN users AS u ON lp.creator_id = u.user_id
        LEFT JOIN (
            SELECT path_id, COUNT(node_id) as total_nodes 
            FROM node 
            GROUP BY path_id
        ) AS n_count ON lp.path_id = n_count.path_id
        LEFT JOIN (
            SELECT path_id, COUNT(enroll_id) as total_students 
            FROM path_enroll 
            GROUP BY path_id
        ) AS pe_count ON lp.path_id = pe_count.path_id
        WHERE lp.path_id = @p1`

    var p model.LearningPath

    err := r.db.QueryRowContext(ctx, pathQuery, path_id).Scan(
        &p.PathID,
        &p.Title,
        &p.CoverImgURL,
        &p.Objective,
        &p.Description,
        &p.Rating,
        &p.Publish_status,
        &p.CreatedAt,
        &p.UpdatedAt,
        &p.Instructor,
        &p.CreatorID,
        &p.Modules,
        &p.Students,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, err
        }
        return nil, fmt.Errorf("repo.GetLearningPathByID scan failed: %w", err)
    }
    return &p, nil
}

func (r *repositoryImpl) CreateLearningPath(ctx context.Context, req model.CreatePathRequest) (string, error) {
	newID := uuid.New().String()
	query := `INSERT INTO learning_path (path_id, title, objective, description, cover_img_url, avg_rating, publish_status, create_at, update_at, creator_ID) VALUES (@p1, @p2, @p3, @p4, @p5, 0.0, @p6, GETDATE(), GETDATE(), @p7)`

	_, err := r.db.ExecContext(ctx, query, newID, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Publish_status, req.CreatorID)
	if err != nil {
		return "", fmt.Errorf("repo.CreateLearningPath exec failed: %w", err)
	}
	return newID, nil
}

func (r *repositoryImpl) UpdateLearningPath(ctx context.Context, path_id string, req model.UpdatePathRequest) error {
	query := `UPDATE learning_path SET title=@p1, objective=@p2, description=@p3, cover_img_url=@p4, publish_status=@p5, update_at=GETDATE() WHERE path_id=@p6`
	res, err := r.db.ExecContext(ctx, query, req.Title, req.Objective, req.Description, req.CoverImgURL, req.Publish_status, path_id)
	if err != nil {
		return fmt.Errorf("repo.UpdateLearningPath failed [id=%s]: %w", path_id, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) DeleteLearningPath(ctx context.Context, path_id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM learning_path WHERE path_id = @p1", path_id)
	if err != nil {
		return fmt.Errorf("repo.DeleteLearningPath failed [id=%s]: %w", path_id, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) EnrollLearningPathUser(ctx context.Context, pathID string, userID string) error {
	enrollID := uuid.New().String()
	query := `INSERT INTO path_enroll (enroll_id, enrollment_status, enroll_at, user_id, path_id) VALUES (@p1, 'active', GETDATE(), @p2, @p3)`
	_, err := r.db.ExecContext(ctx, query, enrollID, userID, pathID)
	if err != nil {
		return fmt.Errorf("repo.EnrollLearningPathUser failed: %w", err)
	}
	return nil
}

func (r *repositoryImpl) GetLearningPathEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error) {
	query := `SELECT CONVERT(VARCHAR(36), enroll_id) as enroll_id, enrollment_status, enroll_at, complete_at FROM path_enroll WHERE user_id = @p1 AND path_id = @p2`
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

func (r *repositoryImpl) UpdateLearningPathImage(ctx context.Context, pathID string, coverImgURL string) error {
	query := `UPDATE learning_path SET cover_img_url = @p1, update_at = GETDATE() WHERE path_id = @p2`

	res, err := r.db.ExecContext(ctx, query, coverImgURL, pathID)
	if err != nil {
		return fmt.Errorf("repo.UpdateLearningPathImage failed [id=%s]: %w", pathID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) GetUserEnrolledPaths(ctx context.Context, userID string) ([]model.EnrolledPathResponse, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), pe.enroll_id) as enroll_id,
			pe.enrollment_status,
			pe.enroll_at,
			
			CONVERT(VARCHAR(36), lp.path_id) as path_id,
			lp.title,
			lp.description,
			lp.cover_img_url,
			lp.avg_rating,
			u.first_name as instructor,
			
			ISNULL(n_count.total_nodes, 0) as modules,

			ISNULL(progress_count.completed, 0) as completed_nodes

		FROM path_enroll pe
		JOIN learning_path lp ON pe.path_id = lp.path_id
		JOIN users u ON lp.creator_id = u.user_id
		
		LEFT JOIN (
			SELECT path_id, COUNT(node_id) as total_nodes 
			FROM node 
			GROUP BY path_id
		) AS n_count ON lp.path_id = n_count.path_id

		LEFT JOIN (
			SELECT n.path_id, COUNT(np.node_id) as completed
			FROM node_progress np
			JOIN node n ON np.node_id = n.node_id
			WHERE np.user_id = @p1 AND np.status = 'completed'
			GROUP BY n.path_id
		) AS progress_count ON lp.path_id = progress_count.path_id

		WHERE pe.user_id = @p2
		ORDER BY pe.enroll_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetUserEnrolledPaths query failed: %w", err)
	}
	defer rows.Close()

	var paths []model.EnrolledPathResponse
	for rows.Next() {
		var p model.EnrolledPathResponse
		var lastAccess time.Time
		if err := rows.Scan(
			&p.EnrollID,
			&p.EnrollmentStatus,
			&lastAccess,
			&p.PathID,
			&p.Title,
			&p.Description,
			&p.CoverImgURL,
			&p.Rating,
			&p.Instructor,
			&p.Modules,
			&p.CompletedNodes,
		); err != nil {
			if err == sql.ErrNoRows {
				return nil, err
			}
			return nil, fmt.Errorf("repo.GetUserEnrolledPaths failed: %w", err)
		}
		p.LastAccessedAt = &lastAccess
		paths = append(paths, p)
	}

	return paths, nil
}
