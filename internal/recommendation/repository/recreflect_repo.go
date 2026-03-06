package repository

import (
	"context"
	"fmt"
	"passiontree/internal/recommendation/model"
)

func (r *repositoryImpl) GetUserReflectionsByTree(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), r.reflect_id) as reflect_id,
			ISNULL(r.summary, '') as summary,
			ISNULL(r.primary_emotion, '') as primary_emotion,
			ISNULL(r.struggle_point, '') as struggle_point,
			ISNULL(r.weighted_reflection_score, 0) as weighted_score,
			CONVERT(VARCHAR(36), t.path_id) as current_path_id
		FROM dbo.Reflect r
		JOIN dbo.Tree_Node tn ON r.tree_node_id = tn.tree_node_id
		JOIN dbo.Tree t ON tn.tree_id = t.tree_id
		JOIN dbo.Tree_Album ta ON t.tree_id = ta.tree_id
		WHERE tn.tree_id = @p1 AND ta.user_id = @p2
		ORDER BY r.create_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, treeID, userID)
	if err != nil {
		return nil, "", fmt.Errorf("repo.GetUserReflectionsByTree query context failed: %w", err)
	}
	defer rows.Close()

	var reflections []model.UserReflection
	var currentPathID string

	for rows.Next() {
		var ref model.UserReflection
		var pathID string
		if err := rows.Scan(
			&ref.ReflectID, &ref.Summary, &ref.PrimaryEmotion,
			&ref.StrugglePoint, &ref.WeightedScore, &pathID,
		); err != nil {
			return nil, "", fmt.Errorf("repo.GetUserReflectionsByTree row scanning failed: %w", err)
		}
		reflections = append(reflections, ref)
		currentPathID = pathID
	}

	return reflections, currentPathID, nil
}
