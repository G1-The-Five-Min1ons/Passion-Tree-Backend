package repository

import (
	"context"
	"database/sql"
	"fmt"
	"passiontree/internal/reflection/model"
)

// GetReflectionStats returns statistics about reflections (admin only)
func (r *repositoryImpl) GetReflectionStats(ctx context.Context) (*model.ReflectionStats, error) {
	stats := &model.ReflectionStats{}

	// Get total reflections
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Reflect").Scan(&stats.TotalReflections)
	if err != nil {
		return nil, fmt.Errorf("failed to get total reflections: %w", err)
	}

	// Get reflections this week
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM Reflect 
		WHERE create_at >= DATEADD(week, -1, GETDATE())
	`).Scan(&stats.ThisWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get this week reflections: %w", err)
	}

	// Get reflections this month
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM Reflect 
		WHERE create_at >= DATEADD(month, -1, GETDATE())
	`).Scan(&stats.ThisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month reflections: %w", err)
	}

	// Get unique authors (via tree_node -> tree -> tree_album -> user)
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT ta.user_id) 
		FROM Reflect r
		INNER JOIN tree_node tn ON r.tree_node_id = tn.tree_node_id
		INNER JOIN tree t ON tn.tree_id = t.tree_id
		INNER JOIN tree_album ta ON t.album_id = ta.album_id
	`).Scan(&stats.UniqueAuthors)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique authors: %w", err)
	}

	// Get average reflection score
	var avgScore sql.NullFloat64
	err = r.db.QueryRowContext(ctx, `
		SELECT AVG(CAST(reflection_score AS FLOAT)) FROM Reflect
	`).Scan(&avgScore)
	if err != nil {
		return nil, fmt.Errorf("failed to get average reflection score: %w", err)
	}
	if avgScore.Valid {
		stats.AvgReflectionScore = avgScore.Float64
	}

	// Get total trees
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tree").Scan(&stats.TotalTrees)
	if err != nil {
		return nil, fmt.Errorf("failed to get total trees: %w", err)
	}

	// Get total albums
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tree_album").Scan(&stats.TotalAlbums)
	if err != nil {
		return nil, fmt.Errorf("failed to get total albums: %w", err)
	}

	return stats, nil
}
