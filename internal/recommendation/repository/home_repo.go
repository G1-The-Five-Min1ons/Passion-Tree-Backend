package repository

import (
	"context"
	"errors"
	"passiontree/internal/recommendation/model"
)

func (r *repositoryImpl) GetUserEnrolledPathsForRec(ctx context.Context, userID string) ([]model.RecommendedPath, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), lp.path_id) as path_id,
			ISNULL(lp.title, 'null') as title,
			ISNULL(lp.description, 'null') as description
		FROM dbo.Path_Enroll pe
		JOIN dbo.Learning_Path lp ON pe.path_id = lp.path_id
		WHERE pe.user_id = @p1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.New("failed to query enrolled paths: " + err.Error())
	}
	defer rows.Close()

	var paths []model.RecommendedPath
	for rows.Next() {
		var p model.RecommendedPath
		if err := rows.Scan(&p.PathID, &p.Title, &p.Description); err != nil {
			return nil, errors.New("failed to scan enrolled path: " + err.Error())
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func (r *repositoryImpl) GetTopPopularPaths(ctx context.Context) ([]model.RecommendedPath, error) {
	query := `
		SELECT TOP 5 
			CONVERT(VARCHAR(36), lp.path_id) as path_id, 
			ISNULL(lp.title, 'null') as title, 
			ISNULL(lp.cover_img_url, 'null') as cover_img_url, 
			ISNULL(lp.objective, 'null') as objective, 
			ISNULL(lp.avg_rating, 0) as recommendation_score
		FROM dbo.Learning_Path lp
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
		return nil, errors.New("failed to query popular paths: " + err.Error())
	}
	defer rows.Close()

	var paths []model.RecommendedPath
	for rows.Next() {
		var p model.RecommendedPath
		if err := rows.Scan(&p.PathID, &p.Title, &p.CoverImgURL, &p.Objective, &p.RecommendationScore); err != nil {
			return nil, errors.New("failed to scan popular path: " + err.Error())
		}
		p.Reason = "Popular learning path based on ratings and enrollments."
		paths = append(paths, p)
	}
	return paths, nil
}
