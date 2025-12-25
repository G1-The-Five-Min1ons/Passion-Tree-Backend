package learningpath

import "database/sql"

type Repository interface {
	// ... (interface เดิม)
	CreateNode(pathID int, req CreateNodeRequest) (int, error)
	GetNodeByID(nodeID int) (*Node, error)
	UpdateNode(nodeID int, req UpdateNodeRequest) error
	DeleteNode(nodeID int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateNode(pathID int, req CreateNodeRequest) (int, error) {
	// SQL: INSERT INTO nodes (path_id, title, description, video_url, file_url, order_index)
	// VALUES (?, ?, ?, ?, ?, ?) RETURNING id
	return 201, nil // สมมติว่าได้ ID 201
}

func (r *repository) GetNodeByID(nodeID int) (*Node, error) {
	// SQL: SELECT * FROM nodes WHERE id = ?
	// return &Node{
	// 	ID:          nodeID,
	// 	Title:       "Introduction to Variables",
	// 	Description: "Learn about int, string, bool",
	// 	VideoURL:    "https://video.example.com/v/123",
	// 	Order:       1,
	// }, nil

	return nil, nil
}

func (r *repository) UpdateNode(nodeID int, req UpdateNodeRequest) error {
	// SQL: UPDATE nodes SET title=?, video_url=?, ... WHERE id=?
	return nil
}

func (r *repository) DeleteNode(nodeID int) error {
	// SQL: DELETE FROM nodes WHERE id = ?
	return nil
}