package recommendation

import "database/sql"

type Repository interface {
	GetPopularItems() ([]RecommendedItem, error)
	GetPersonalizedItems(userID int) ([]RecommendedItem, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetPopularItems() ([]RecommendedItem, error) {
	// SELECT TOP 5 id, name, price FROM products ORDER BY sales DESC
	// return []RecommendedItem{
	// 	{ID: 1, Name: "Azure Mouse", Price: 500, Score: 4.8},
	// 	{ID: 2, Name: "Go Lang T-Shirt", Price: 300, Score: 4.5},
	// }, nil

	return nil, nil
}

func (r *repository) GetPersonalizedItems(userID int) ([]RecommendedItem, error) {
	// ตัวอย่าง Query จริง:
	// JOIN order_history ... WHERE user_id = @p1
	// return []RecommendedItem{
	// 	{ID: 99, Name: "Mechanical Keyboard (For You)", Price: 2500, Score: 4.9},
	// }, nil
	
	return nil, nil
}