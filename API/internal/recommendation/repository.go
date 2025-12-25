package recommendation

import "database/sql"

type Repository interface {
	GetPopularItems() ([]RecommendedItem, error)
	GetPersonalizedItems(userID int) ([]RecommendedItem, error)
	GetNextPathInTree(treeID int, userID int) ([]RecommendedItem, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetPopularItems() ([]RecommendedItem, error) {
	// return []RecommendedItem{
	// 	{ID: 101, Name: "Popular: Go Basics", Score: 5.0},
	// 	{ID: 102, Name: "Popular: SQL Fundamentals", Score: 4.8},
	// }, nil

	return nil, nil
}

func (r *repository) GetPersonalizedItems(userID int) ([]RecommendedItem, error) {
	// if userID == 999 {
	// 	return []RecommendedItem{}, nil
	// }
	
	// return []RecommendedItem{
	// 	{ID: 201, Name: "Personal: Advanced Go", Score: 4.9},
	// }, nil

	return nil, nil
}

func (r *repository) GetNextPathInTree(treeID int, userID int) ([]RecommendedItem, error) {
	// Logic: Query ดูว่าใน Tree นี้ User เรียนถึง Node ไหนแล้ว แล้ว return Node ถัดไป
	// return []RecommendedItem{
	// 	{ID: 305, Name: "Tree Step: Next Node in Tree " + strconv.Itoa(treeID), Score: 1.0},
	// }, nil
	
	return nil, nil
}