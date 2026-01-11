package repository

import (
	"database/sql"
	"fmt"
	"passiontree/internal/database"
	"passiontree/internal/history/model"
)

type RepositoryHistory interface {
	GetHistoryByUserID(userID string) ([]model.HistoryResponse, error)
}

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(ds database.Database) RepositoryHistory {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}

func (r *repositoryImpl) GetHistoryByUserID(userID string) ([]model.HistoryResponse, error) {
	query := `
		SELECT 
    		target_path.path_id,
    		n.node_id
		FROM (
    		SELECT path_id
    		FROM path_enroll
    		WHERE user_id = ?
    		ORDER BY update_at DESC
    		LIMIT 1
		) AS target_path
		JOIN node n ON target_path.path_id = n.path_id
		LEFT JOIN node_progress np ON n.node_id = np.node_id AND np.user_id = ? 
		WHERE 
    		(np.status IS NULL OR np.status != 'completed')
		ORDER BY 
    		n.created_at ASC 
		LIMIT 1;`

	rows, err := r.db.Query(query, userID)
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