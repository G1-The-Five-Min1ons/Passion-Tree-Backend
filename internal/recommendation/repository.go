package recommendation

import (
	"database/sql"
	
	"passiontree/internal/database" 
)

type Repository interface {
	GetPopularItems() ([]RecommendedItem, error)
	GetPersonalizedItems(userID string) ([]RecommendedItem, error)
	GetNextPathInTree(treeID string, userID string) ([]RecommendedItem, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(ds database.Database) Repository {
	return &repository{
		db: ds.GetDB(), 
	}
}

func (r *repository) GetPopularItems() ([]RecommendedItem, error) {
	query := `
		SELECT path_id
		FROM learning_path lp
		ORDER BY avg_rating DESC 
		LIMIT 10
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RecommendedItem
	for rows.Next() {
		var item RecommendedItem
		if err := rows.Scan(&item.Path_id); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *repository) GetPersonalizedItems(userID string) ([]RecommendedItem, error) {
	query := `
		SELECT lp.path_id
		FROM learning_path lp
		JOIN path_enroll pe ON lp.path_id = pe.path_id
		WHERE pe.user_id = $1
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RecommendedItem
	for rows.Next() {
		var item RecommendedItem
		if err := rows.Scan(&item.Path_id); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *repository) GetNextPathInTree(treeID string, userID string) ([]RecommendedItem, error) {
	// query := `

	// `
	return nil, nil
}