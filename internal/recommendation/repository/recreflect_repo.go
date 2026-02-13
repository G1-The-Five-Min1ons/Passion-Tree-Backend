package repository

import (
	"context"
	"fmt"
	"passiontree/internal/recommendation/model"
)

func (r *repositoryImpl) GetUserReflectionsAllNodes(ctx context.Context, user_id string, path_id string) ([]model.UserReflection, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), reflect_id) as reflect_id,
			ISNULL(reflect_description, '') as reflect_description,
			ISNULL(mood, '') as mood,
			ISNULL(tag, '') as tag,
			ISNULL(challenge_score, 0) as challenge_score,
			ISNULL(progress_score, 0) as progress_score,
			CONVERT(VARCHAR(36), tree_node_id) as tree_node_id
		FROM dbo.Reflect
		WHERE user_id = @p1 AND path_id = @p2
		ORDER BY create_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, user_id, path_id)
	if err != nil {
		return nil, fmt.Errorf("repo.GetUserReflectionsAllNodes query failed: %w", err)
	}
	defer rows.Close()

	var reflections []model.UserReflection
	for rows.Next() {
		var ref model.UserReflection
		if err := rows.Scan(
			&ref.ReflectID, &ref.ReflectDescription, &ref.Mood, &ref.Tag, 
			&ref.ChallengeScore, &ref.ProgressScore, &ref.TreeNodeID,
		); err != nil {
			return nil, fmt.Errorf("repo.GetUserReflectionsAllNodes scan failed: %w", err)
		}
		reflections = append(reflections, ref)
	}

	return reflections, nil
}