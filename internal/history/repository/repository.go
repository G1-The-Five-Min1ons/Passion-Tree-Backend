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
			pe.enroll_id, 
			pe.status, 
			pe.enroll_at, 
			pe.complete_at,
			lp.path_id, 
			lp.title, 
			lp.cover_img_url, 
			lp.objective, 
			IFNULL(lp.creator_ID, '')
		FROM path_enroll pe
		JOIN learning_path lp ON pe.path_id = lp.path_id
		WHERE pe.user_id = ?
		ORDER BY pe.enroll_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("historyRepo.GetHistoryByUserID query failed: %w", err)
	}
	defer rows.Close()

	var historyList []model.HistoryResponse
	for rows.Next() {
		var h model.HistoryResponse
		if err := rows.Scan(
			&h.EnrollID, 
			&h.Status, 
			&h.EnrollAt, 
			&h.CompleteAt, 
			&h.PathID, 
			&h.Title, 
			&h.CoverImgURL, 
			&h.Objective, 
			&h.CreatorID,
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