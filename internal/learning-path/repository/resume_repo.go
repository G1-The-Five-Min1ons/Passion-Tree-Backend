package repository

import (
	"context"
	"database/sql"
	"fmt"
)

func (r *repositoryImpl) GetNextNodeID(ctx context.Context, userID string, pathID string) (string, error) {
	query := `
		SELECT n.node_id
		FROM node n
		LEFT JOIN node_progress np ON n.node_id = np.node_id AND np.user_id = ?
		WHERE n.path_id = ?
		AND (np.status IS NULL OR np.status != 'completed')
		ORDER BY n.created_at ASC
		LIMIT 1
	`

	var nodeID string
	err := r.db.QueryRowContext(ctx, query, userID, pathID).Scan(&nodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("resumeRepo.GetNextNodeID failed: %w", err)
	}

	return nodeID, nil
}