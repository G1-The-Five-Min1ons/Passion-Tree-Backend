package repository

import (
	"context"
	"fmt"
	"database/sql"
	"passiontree/internal/recommendation/model"
)

func (r *repositoryImpl) GetUserReflectionsByTree(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
	query := `
		SELECT 
			CONVERT(VARCHAR(36), t.path_id) as current_path_id,
			CONVERT(VARCHAR(36), r.reflect_id) as reflect_id,
			r.summary,
			r.primary_emotion,
			r.struggle_point,
			r.weighted_reflection_score as weighted_score
		FROM dbo.Tree t
		JOIN dbo.Tree_Album ta ON t.tree_id = ta.tree_id
		LEFT JOIN dbo.Tree_Node tn ON t.tree_id = tn.tree_id
		LEFT JOIN dbo.Reflect r ON tn.tree_node_id = r.tree_node_id
		WHERE t.tree_id = @p1 AND ta.user_id = @p2
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
		var reflectID, summary, primaryEmotion, strugglePoint sql.NullString
		var weightedScore sql.NullFloat64
		var pathID string

		if err := rows.Scan(
			&pathID, &reflectID, &summary, &primaryEmotion,
			&strugglePoint, &weightedScore,
		); err != nil {
			return nil, "", fmt.Errorf("repo.GetUserReflectionsByTree row scanning failed: %w", err)
		}

		currentPathID = pathID

		if reflectID.Valid {
			ref := model.UserReflection{
				ReflectID:      reflectID.String,
				Summary:        summary.String,
				PrimaryEmotion: primaryEmotion.String,
				StrugglePoint:  strugglePoint.String,
				WeightedScore:  weightedScore.Float64,
			}
			reflections = append(reflections, ref)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("repo.GetAllLearnningPath row iteration failed: %w", err)
	}

	return reflections, currentPathID, nil
}
