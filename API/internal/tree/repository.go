package tree

import "database/sql"

type Repository interface {
	CreateAlbum(req CreateAlbumRequest) (int, error)
	GetAllAlbums() ([]TreeAlbum, error)
	UpdateAlbum(id int, req UpdateAlbumRequest) error
	DeleteAlbum(id int) error

	CreateTree(albumID int, req CreateTreeRequest) (int, error)
	GetTreesByAlbumID(albumID int) ([]Tree, error)
	GetTreeByID(treeID int) (*Tree, error)

	CreateNode(treeID int, req CreateNodeRequest) (int, error)
	GetNodeByID(nodeID int) (*TreeNode, error)
	UpdateNode(nodeID int, req UpdateNodeRequest) error
	DeleteNode(nodeID int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateAlbum(req CreateAlbumRequest) (int, error) {
	return 101, nil
}

func (r *repository) GetAllAlbums() ([]TreeAlbum, error) {
	return []TreeAlbum{
		{ID: 101, Title: "My Coding Journey", IconURL: "icon_code.png"},
		{ID: 102, Title: "Art & Design", IconURL: "icon_paint.png"},
	}, nil
}

func (r *repository) UpdateAlbum(id int, req UpdateAlbumRequest) error {
	return nil
}

func (r *repository) DeleteAlbum(id int) error {
	return nil
}


func (r *repository) CreateTree(albumID int, req CreateTreeRequest) (int, error) {
	return 501, nil
}

func (r *repository) GetTreesByAlbumID(albumID int) ([]Tree, error) {
	return []Tree{
		{ID: 501, AlbumID: albumID, Name: "Golang Mastery", Variety: "Banyan", PlantedAt: "2023-01-01"},
		{ID: 502, AlbumID: albumID, Name: "Algorithm Basics", Variety: "Oak", PlantedAt: "2023-02-15"},
	}, nil
}

func (r *repository) GetTreeByID(treeID int) (*Tree, error) {
	return &Tree{
		ID: treeID, AlbumID: 101, Name: "Golang Mastery", Variety: "Banyan", PlantedAt: "2023-01-01",
	}, nil
}

func (r *repository) CreateNode(treeID int, req CreateNodeRequest) (int, error) {
	return 901, nil
}

func (r *repository) GetNodeByID(nodeID int) (*TreeNode, error) {
	return &TreeNode{
		ID:         nodeID,
		TreeID:     501,
		Title:      "Learned Structs",
		Reflection: "Today I learned how to define structs in Go. It's like a Class but lighter.",
		ImageURL:   "struct_diagram.png",
	}, nil
}

func (r *repository) UpdateNode(nodeID int, req UpdateNodeRequest) error {
	return nil
}

func (r *repository) DeleteNode(nodeID int) error {
	return nil
}