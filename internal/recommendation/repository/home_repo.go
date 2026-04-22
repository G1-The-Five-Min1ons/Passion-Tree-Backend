package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"passiontree/internal/recommendation/model"
)

func (r *repositoryImpl) GetTopPopularPaths(ctx context.Context) ([]model.RecommendedPath, error) {
	query := `
		SELECT TOP 5
			CONVERT(VARCHAR(36), lp.path_id) as path_id,
			ISNULL(lp.title, 'null') as title,
			ISNULL(lp.cover_img_url, 'null') as cover_img_url,
			ISNULL(lp.objective, 'null') as objective,
			ISNULL(lp.description, 'null') as description,
			ISNULL(lp.avg_rating, 0) as avg_rating,
			ISNULL(lp.publish_status, 'null') as publish_status,
			lp.create_at,
			lp.update_at,
			CONVERT(VARCHAR(36), lp.creator_id) as creator_id,
			ISNULL(u.first_name, '') as creator_name,
			ISNULL(u.username, '') as creator_username,
			ISNULL(u.first_name, '') as instructor,
			ISNULL(n_count.total_nodes, 0) as modules,
			ISNULL(pe_count.total_students, 0) as student,
			ISNULL(lp.avg_rating, 0) as recommendation_score
		FROM dbo.Learning_Path lp
		JOIN dbo.users u ON lp.creator_id = u.user_id
		LEFT JOIN (
			SELECT path_id, COUNT(node_id) as total_nodes
			FROM dbo.Node
			GROUP BY path_id
		) AS n_count ON lp.path_id = n_count.path_id
		LEFT JOIN (
			SELECT path_id, COUNT(enroll_id) as total_students
			FROM dbo.Path_Enroll
			GROUP BY path_id
		) AS pe_count ON lp.path_id = pe_count.path_id
		WHERE lp.publish_status = 'published'
		ORDER BY lp.avg_rating DESC, pe_count.total_students DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repo.GetTopPopularPaths query failed: %w", err)
	}
	defer rows.Close()

	var paths []model.RecommendedPath
	for rows.Next() {
		var p model.RecommendedPath
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
			&p.CreatorID,
			&p.CreatorName,
			&p.CreatorUsername,
			&p.Instructor,
			&p.Modules,
			&p.Students,
			&p.RecommendationScore,
		); err != nil {
			return nil, fmt.Errorf("repo.GetTopPopularPaths scan failed: %w", err)
		}
		p.Reason = "Popular learning path based on ratings and enrollments."
		paths = append(paths, p)
	}
	return paths, nil
}

func (r *repositoryImpl) GetBatchInteractions(ctx context.Context) ([]model.UserInteraction, error) {
	//Enroll=2, Complete Path=5, Active Node=+1, Complete Node=+2
	query := `
		SELECT 
			CONVERT(VARCHAR(36), pe.user_id) AS user_id,
			CONVERT(VARCHAR(36), pe.path_id) AS path_id,
			(
				CASE WHEN pe.enrollment_status = 'completed' THEN 5.0 ELSE 2.0 END
				+
				ISNULL((
					SELECT SUM(CASE WHEN np.complete = 'true' THEN 2.0 ELSE 1.0 END)
					FROM dbo.Node_progress np
					JOIN dbo.Node n ON np.node_id = n.node_id
					WHERE np.user_id = pe.user_id AND n.path_id = pe.path_id
				), 0)
			) AS score
		FROM dbo.Path_Enroll pe
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query batch interactions: %w", err)
	}
	defer rows.Close()

	var interactions []model.UserInteraction
	for rows.Next() {
		var i model.UserInteraction
		if err := rows.Scan(&i.UserID, &i.PathID, &i.Score); err != nil {
			return nil, fmt.Errorf("failed to scan interaction: %w", err)
		}
		interactions = append(interactions, i)
	}
	return interactions, nil
}

func (r *repositoryImpl) GetBatchProfiles(ctx context.Context) ([]model.UserProfile, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), user_id) AS user_id,
			CONCAT(
				'Interested in: ', ISNULL(subjects, ''), '. ',
				'Learning style: ', ISNULL(learning_styles, ''), '. ',
				'Motivation: ', ISNULL(motivation, ''), '. ',
				'Daily Goal: ', ISNULL(daily_goal, '')
			) AS interests
		FROM dbo.onboarding_answers
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query batch profiles: %w", err)
	}
	defer rows.Close()

	var profiles []model.UserProfile
	for rows.Next() {
		var p model.UserProfile
		if err := rows.Scan(&p.UserID, &p.Interests); err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (r *repositoryImpl) SaveBatchRecommendations(ctx context.Context, results []model.BatchRecommendationResult) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	deleteQuery := `DELETE FROM dbo.Recommendation WHERE user_id = @p1`

	insertQuery := `
		INSERT INTO dbo.Recommendation (recommend_id, user_id, path_id, rank_order, score, updated_at)
		VALUES (@p1, @p2, @p3, @p4, @p5, GETUTCDATE())
	`

	for _, userResult := range results {
		_, err = tx.ExecContext(ctx, deleteQuery, userResult.UserID)
		if err != nil {
			return fmt.Errorf("failed to delete old recommendations for user %s: %w", userResult.UserID, err)
		}

		for index, rec := range userResult.RecommendedPaths {
			rankOrder := index + 1
			newID := uuid.New().String()

			_, err = tx.ExecContext(ctx, insertQuery, newID, userResult.UserID, rec.PathID, rankOrder, rec.Score)
			if err != nil {
				return fmt.Errorf("failed to insert recommendation (user: %s, path: %s): %w", userResult.UserID, rec.PathID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit recommendations: %w", err)
	}

	return nil
}

func (r *repositoryImpl) GetSavedHomeRecommendations(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
	query := `
		SELECT
			CONVERT(VARCHAR(36), lp.path_id) as path_id,
			ISNULL(lp.title, 'null') as title,
			ISNULL(lp.cover_img_url, 'null') as cover_img_url,
			ISNULL(lp.objective, 'null') as objective,
			ISNULL(lp.description, 'null') as description,
			ISNULL(lp.avg_rating, 0) as avg_rating,
			ISNULL(lp.publish_status, 'null') as publish_status,
			lp.create_at,
			lp.update_at,
			CONVERT(VARCHAR(36), lp.creator_id) as creator_id,
			ISNULL(u.first_name, '') as creator_name,
			ISNULL(u.username, '') as creator_username,
			ISNULL(u.first_name, '') as instructor,
			ISNULL(n_count.total_nodes, 0) as modules,
			ISNULL(pe_count.total_students, 0) as student,
			ISNULL(r.score, 0) as recommendation_score
		FROM dbo.Recommendation r
		JOIN dbo.Learning_Path lp ON r.path_id = lp.path_id
		JOIN dbo.users u ON lp.creator_id = u.user_id
		LEFT JOIN (
			SELECT path_id, COUNT(node_id) as total_nodes
			FROM dbo.Node
			GROUP BY path_id
		) AS n_count ON lp.path_id = n_count.path_id
		LEFT JOIN (
			SELECT path_id, COUNT(enroll_id) as total_students
			FROM dbo.Path_Enroll
			GROUP BY path_id
		) AS pe_count ON lp.path_id = pe_count.path_id
		WHERE r.user_id = @p1 AND lp.publish_status = 'published'
		ORDER BY r.rank_order ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetSavedHomeRecommendations query failed: %w", err)
	}
	defer rows.Close()

	var paths []model.RecommendedPath
	for rows.Next() {
		var p model.RecommendedPath
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
			&p.CreatorID,
			&p.CreatorName,
			&p.CreatorUsername,
			&p.Instructor,
			&p.Modules,
			&p.Students,
			&p.RecommendationScore,
		); err != nil {
			return nil, fmt.Errorf("repo.GetSavedHomeRecommendations scan failed: %w", err)
		}
		p.Reason = "Recommended for you based on your interests and learning history."
		paths = append(paths, p)
	}
	return paths, nil
}
