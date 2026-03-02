package repository

import (
	"context"
	"fmt"
	"passiontree/internal/learning-path/model"
)

func (r *repositoryImpl) GetHistoryByUserID(ctx context.Context, userID string) ([]model.HistoryResponse, error) {
	query := `
		WITH NextNodes AS (
			SELECT 
				n.path_id,
				n.node_id,
				ROW_NUMBER() OVER(PARTITION BY n.path_id ORDER BY n.sequence ASC) as rn
			FROM path_enroll pe
			JOIN node n ON pe.path_id = n.path_id
			LEFT JOIN node_progress np ON n.node_id = np.node_id AND np.user_id = @p2
			WHERE pe.user_id = @p1
			  AND (np.status IS NULL OR np.status != 'Completed')
		)
		SELECT path_id, node_id
		FROM NextNodes
		WHERE rn = 1;`

	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("historyRepo.GetHistoryByUserID query failed: %w", err)
	}
	defer rows.Close()

	var historyList []model.HistoryResponse
	for rows.Next() {
		var h model.HistoryResponse
		if err := rows.Scan(
			&h.Target_path,
			&h.Path_id,
		); err != nil {
			return nil, fmt.Errorf("historyRepo.GetHistoryByUserID scan failed: %w", err)
		}
		historyList = append(historyList, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("historyRepo.GetHistoryByUserID row iteration failed: %w", err)
	}

	return historyList, nil
}
