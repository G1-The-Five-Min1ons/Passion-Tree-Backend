package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"passiontree/internal/learning-path/model"

	"github.com/google/uuid"
)

func (r *repositoryImpl) GetAllLearningPath(ctx context.Context) ([]model.LearningPath, error) {
	query := `
		SELECT 
   			CONVERT(VARCHAR(36), lp.path_id) as id, 
    		lp.title, 
    		ISNULL(lp.description, 'null') as description,
    		u.first_name as instructor,
    		u.first_name as creator_name,
    		u.username as creator_username,
    		ISNULL(pe_count.total_students, 0) as student,
			ISNULL(n_count.total_nodes, 0) as modules,
			ISNULL(lp.avg_rating, 0) as avg_rating,
    		ISNULL(lp.cover_img_url, 'null') as cover_img_url, 
    		ISNULL(lp.objective, 'null') as objective,
    		ISNULL(lp.publish_status, 'null') as publish_status, 
    		ISNULL(lp.create_at, GETDATE()) as create_at, 
    		ISNULL(lp.update_at, GETDATE()) as update_at,
			CONVERT(VARCHAR(36), lp.creator_id) creator_id
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

	var paths []model.LearningPath
	for rows.Next() {
		var p model.LearningPath
		if err := rows.Scan(
			&p.PathID,
			&p.Title,
			&p.Description,
			&p.Instructor,
			&p.CreatorName,
			&p.CreatorUsername,
			&p.Students,
			&p.Modules,
			&p.Rating,
			&p.CoverImgURL,
			&p.Objective,
			&p.Publish_status,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.CreatorID,
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
            u.first_name as creator_name,
            u.username as creator_username,
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
		&p.CreatorName,
		&p.CreatorUsername,
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

func (r *repositoryImpl) GetPathCreatorVerification(ctx context.Context, userID string) (string, bool, error) {
	query := `
		SELECT 
    		u.role,
    		CASE 
        		WHEN EXISTS (
            		SELECT 1 
            		FROM teacher_verification_requests tvr 
            		WHERE tvr.user_id = u.user_id 
            		AND tvr.status = 'approved'
            		AND ISNULL(tvr.phone_number, '') <> '' 
        		) 
        		THEN 1 
        		ELSE 0 
    		END AS is_teacher_verified
		FROM users u
		WHERE u.user_id = @p1`

	var (
		role            string
		isVerifiedAsInt int
	)

	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&role, &isVerifiedAsInt); err != nil {
		if err == sql.ErrNoRows {
			return "", false, err
		}
		return "", false, fmt.Errorf("repo.GetPathCreatorVerification failed: %w", err)
	}

	return role, isVerifiedAsInt == 1, nil
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
	// Check if user already enrolled in this learning path
	checkQuery := `SELECT COUNT(*) FROM path_enroll WHERE user_id = @p1 AND path_id = @p2`
	var count int
	err := r.db.QueryRowContext(ctx, checkQuery, userID, pathID).Scan(&count)
	if err != nil {
		return fmt.Errorf("repo.EnrollLearningPathUser check enrollment failed: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("user already enrolled in this learning path")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo.EnrollLearningPathUser begin tx failed: %w", err)
	}
	defer tx.Rollback()

	enrollID := uuid.New().String()
	enrollQuery := `INSERT INTO path_enroll (enroll_id, enrollment_status, enroll_at, user_id, path_id) VALUES (@p1, 'active', GETDATE(), @p2, @p3)`

	_, err = tx.ExecContext(ctx, enrollQuery, enrollID, userID, pathID)
	if err != nil {
		return fmt.Errorf("repo.EnrollLearningPathUser insert path_enroll failed: %w", err)
	}

	progressQuery := `
		INSERT INTO node_progress (progress_id, user_id, node_id, status, updated_at, complete)
		SELECT 
			NEWID(),
			@p1,
			node_id,
			CASE WHEN sequence = 1 THEN 'active' ELSE 'locked' END,
			GETDATE(),
			'false'
		FROM node 
		WHERE path_id = @p2
	`

	_, err = tx.ExecContext(ctx, progressQuery, userID, pathID)
	if err != nil {
		return fmt.Errorf("repo.EnrollLearningPathUser insert node_progress failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repo.EnrollLearningPathUser commit failed: %w", err)
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
		WITH LatestEnroll AS (
			SELECT
				CONVERT(VARCHAR(36), enroll_id) as enroll_id,
				enrollment_status,
				enroll_at,
				path_id,
				ROW_NUMBER() OVER (PARTITION BY path_id ORDER BY enroll_at DESC) AS rn
			FROM path_enroll
			WHERE user_id = @p1
		)
		SELECT 
			le.enroll_id,
			le.enrollment_status,
			le.enroll_at,
			
			CONVERT(VARCHAR(36), lp.path_id) as path_id,
			lp.title,
			lp.description,
			lp.cover_img_url,
			lp.avg_rating,
			u.first_name as instructor,
			
			ISNULL(n_count.total_nodes, 0) as modules,
			ISNULL(progress_count.completed, 0) as completed_nodes,
			
			CASE 
				WHEN ISNULL(n_count.total_nodes, 0) = 0 THEN 0.0
				ELSE ROUND((CAST(ISNULL(progress_count.completed, 0) AS FLOAT) / CAST(n_count.total_nodes AS FLOAT)) * 100, 2)
			END as progress_percent,

			CASE
				WHEN ISNULL(n_count.total_nodes, 0) > 0 AND ISNULL(progress_count.completed, 0) >= n_count.total_nodes THEN 'Completed'
				ELSE 'In Progress'
			END as progress_status

		FROM LatestEnroll le
		JOIN learning_path lp ON le.path_id = lp.path_id
		JOIN users u ON lp.creator_id = u.user_id
		
		LEFT JOIN (
			SELECT path_id, COUNT(node_id) as total_nodes 
			FROM node 
			GROUP BY path_id
		) AS n_count ON lp.path_id = n_count.path_id

		LEFT JOIN (
			SELECT n.path_id, COUNT(DISTINCT np.node_id) as completed
			FROM node_progress np
			JOIN node n ON np.node_id = n.node_id
			WHERE np.user_id = @p2 AND np.complete = 'true'
			GROUP BY n.path_id
		) AS progress_count ON lp.path_id = progress_count.path_id

		WHERE le.rn = 1
		ORDER BY le.enroll_at DESC
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
			&p.ProgressPercent,
			&p.ProgressStatus,
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

func (r *repositoryImpl) UpdatePathEnrollmentCompletion(ctx context.Context, pathID string, userID string) error {
	query := `UPDATE path_enroll SET enrollment_status = 'completed', complete_at = GETDATE() WHERE path_id = @p1 AND user_id = @p2`

	res, err := r.db.ExecContext(ctx, query, pathID, userID)
	if err != nil {
		return fmt.Errorf("repo.UpdatePathEnrollmentCompletion exec failed: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) GetLearningPathsByIDs(ctx context.Context, pathIDs []string) ([]model.LearningPath, error) {
	if len(pathIDs) == 0 {
		return []model.LearningPath{}, nil
	}

	placeholders := make([]string, len(pathIDs))
	args := make([]interface{}, len(pathIDs))
	for i, id := range pathIDs {
		placeholders[i] = fmt.Sprintf("@p%d", i+1)
		args[i] = id
	}

	inClause := strings.Join(placeholders, ", ")

	query := fmt.Sprintf(`
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
        WHERE lp.path_id IN (%s)
	`, inClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repo.GetLearningPathsByIDs query failed: %w", err)
	}
	defer rows.Close()

	var paths []model.LearningPath
	for rows.Next() {
		var p model.LearningPath
		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("repo.GetLearningPathsByIDs scan failed: %w", err)
		}
		paths = append(paths, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetLearningPathsByIDs row iteration failed: %w", err)
	}

	return paths, nil
}

func (r *repositoryImpl) UpsertLearningPathRating(ctx context.Context, rating *model.LearningPathRating) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo.UpsertRating begin tx failed: %w", err)
	}
	defer tx.Rollback()

	var existingID string
	var oldRatingOverall float64

	checkQuery := `SELECT CONVERT(VARCHAR(36), rating_id), rating_overall FROM Learning_Path_Rating WHERE path_id = @p1 AND user_id = @p2`
	err = tx.QueryRowContext(ctx, checkQuery, rating.PathID, rating.UserID).Scan(&existingID, &oldRatingOverall)

	var ratingDiff float64
	var countDiff int

	switch err {
	case sql.ErrNoRows:
		rating.RatingID = uuid.New().String()
		insertQuery := `INSERT INTO Learning_Path_Rating (rating_id, rating_content, rating_instruct, rating_overall, user_id, path_id) VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`
		_, err = tx.ExecContext(ctx, insertQuery, rating.RatingID, rating.RatingContent, rating.RatingInstruct, rating.RatingOverall, rating.UserID, rating.PathID)

		ratingDiff = rating.RatingOverall
		countDiff = 1

	case nil:
		updateQuery := `UPDATE Learning_Path_Rating SET rating_content = @p1, rating_instruct = @p2, rating_overall = @p3 WHERE rating_id = @p4`
		_, err = tx.ExecContext(ctx, updateQuery, rating.RatingContent, rating.RatingInstruct, rating.RatingOverall, existingID)

		ratingDiff = rating.RatingOverall - oldRatingOverall
		countDiff = 0

	default:
		return fmt.Errorf("repo.UpsertRating check existing rating failed: %w", err)
	}

	if err != nil {
		return fmt.Errorf("repo.UpsertRating modify rating failed: %w", err)
	}

	if err := r.updatePathRatingStats(ctx, tx, rating.PathID, ratingDiff, countDiff); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repositoryImpl) GetRatingByUser(ctx context.Context, pathID string, userID string) (*model.LearningPathRating, error) {
	query := `SELECT CONVERT(VARCHAR(36), rating_id), rating_content, rating_instruct, rating_overall FROM Learning_Path_Rating WHERE path_id = @p1 AND user_id = @p2`
	var rating model.LearningPathRating
	err := r.db.QueryRowContext(ctx, query, pathID, userID).Scan(&rating.RatingID, &rating.RatingContent, &rating.RatingInstruct, &rating.RatingOverall)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("repo.GetRatingByUser failed: %w", err)
	}
	rating.PathID = pathID
	rating.UserID = userID
	return &rating, nil
}

func (r *repositoryImpl) DeleteLearningPathRating(ctx context.Context, pathID string, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo.DeleteRating begin tx failed: %w", err)
	}
	defer tx.Rollback()

	var oldRatingOverall float64
	getRatingQuery := `SELECT rating_overall FROM Learning_Path_Rating WHERE path_id = @p1 AND user_id = @p2`
	if err := tx.QueryRowContext(ctx, getRatingQuery, pathID, userID).Scan(&oldRatingOverall); err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return fmt.Errorf("repo.DeleteRating fetch old rating failed: %w", err)
	}

	deleteQuery := `DELETE FROM Learning_Path_Rating WHERE path_id = @p1 AND user_id = @p2`
	res, err := tx.ExecContext(ctx, deleteQuery, pathID, userID)
	if err != nil {
		return fmt.Errorf("repo.DeleteRating exec failed: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	if err := r.updatePathRatingStats(ctx, tx, pathID, -oldRatingOverall, -1); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repositoryImpl) updatePathRatingStats(ctx context.Context, tx *sql.Tx, pathID string, ratingDiff float64, countDiff int) error {
	query := `
		UPDATE learning_path 
		SET 
			total_rating = ISNULL(total_rating, 0) + @p1,
			rating_count = ISNULL(rating_count, 0) + @p2,
			avg_rating = CASE 
				WHEN (ISNULL(rating_count, 0) + @p2) > 0 
				THEN (ISNULL(total_rating, 0) + @p1) / CAST((ISNULL(rating_count, 0) + @p2) AS FLOAT)
				ELSE 0 
			END,
			update_at = GETDATE()
		WHERE path_id = @p3
	`
	_, err := tx.ExecContext(ctx, query, ratingDiff, countDiff, pathID)
	if err != nil {
		return fmt.Errorf("repo.updatePathRatingStats update failed: %w", err)
	}
	return nil
}
