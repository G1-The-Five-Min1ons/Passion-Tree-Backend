package learningpath

import (
	"database/sql"
)

type Repository interface {
	GetAll(category string, search string) ([]LearningPath, error)
	GetByID(id int) (*LearningPath, error)
	Create(req CreatePathRequest) (int, error)
	Update(id int, req UpdatePathRequest) error
	Delete(id int) error
	EnrollUser(pathID int, userID int) error
	GetUserProgress(pathID int, userID int) ([]NodeProgress, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll(category string, search string) ([]LearningPath, error) {
	// return []LearningPath{
	// 	{ID: 1, Title: "Go Basics", Category: "Programming"},
	// 	{ID: 2, Title: "Data Science 101", Category: "Data"},
	// }, nil

	return nil, nil
}

func (r *repository) GetByID(id int) (*LearningPath, error) {
	// SQL: SELECT * FROM paths WHERE id = ?
	// SQL: SELECT * FROM nodes WHERE path_id = ? ORDER BY order_index
	// return &LearningPath{
	// 	ID: id, Title: "Go Basics", Description: "Deep dive into Go",
	// 	Nodes: []Node{{ID: 10, Title: "Intro", Order: 1}, {ID: 11, Title: "Variables", Order: 2}},
	// }, nil

	return nil, nil
}

func (r *repository) Create(req CreatePathRequest) (int, error) {
	// SQL: INSERT INTO paths (...) VALUES (...) RETURNING id
	return 101, nil // สมมติว่าสร้างได้ ID 101
}

func (r *repository) Update(id int, req UpdatePathRequest) error {
	// SQL: UPDATE paths SET title=?, ... WHERE id=?
	return nil
}

func (r *repository) Delete(id int) error {
	// SQL Transaction:
	// 1. DELETE FROM nodes WHERE path_id = ?
	// 2. DELETE FROM user_progress WHERE path_id = ?
	// 3. DELETE FROM paths WHERE id = ?
	return nil
}

func (r *repository) EnrollUser(pathID int, userID int) error {
	// SQL: INSERT INTO user_enrollments (user_id, path_id) VALUES (?, ?)
	return nil
}

func (r *repository) GetUserProgress(pathID int, userID int) ([]NodeProgress, error) {
	// SQL: JOIN nodes และ user_node_completion เพื่อเช็คสถานะ
	// return []NodeProgress{
	// 	{Node: Node{ID: 10, Title: "Intro"}, Status: "completed"},
	// 	{Node: Node{ID: 11, Title: "Variables"}, Status: "in_progress"},
	// }, nil
	return nil, nil
}