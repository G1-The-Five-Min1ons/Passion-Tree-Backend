package tree

import "errors"

type Service interface {
	CreateAlbum(req CreateAlbumRequest) (int, error)
	GetAlbums() ([]TreeAlbum, error)
	UpdateAlbum(id int, req UpdateAlbumRequest) error
	DeleteAlbum(id int) error

	PlantTree(albumID int, req CreateTreeRequest) (int, error)
	GetAlbumTrees(albumID int) ([]Tree, error)
	GetTreeDetails(treeID int) (*Tree, error)

	AddNode(treeID int, req CreateNodeRequest) (int, error)
	GetNodeDetails(nodeID int) (*TreeNode, error)
	UpdateNodeReflection(nodeID int, req UpdateNodeRequest) error
	RemoveNode(nodeID int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateAlbum(req CreateAlbumRequest) (int, error) {
	if req.Title == "" {
		return 0, errors.New("album title is required")
	}
	return s.repo.CreateAlbum(req)
}

func (s *service) GetAlbums() ([]TreeAlbum, error) {
	return s.repo.GetAllAlbums()
}

func (s *service) UpdateAlbum(id int, req UpdateAlbumRequest) error {
	return s.repo.UpdateAlbum(id, req)
}

func (s *service) DeleteAlbum(id int) error {
	return s.repo.DeleteAlbum(id)
}

func (s *service) PlantTree(albumID int, req CreateTreeRequest) (int, error) {
	if req.Name == "" {
		return 0, errors.New("tree name is required")
	}
	return s.repo.CreateTree(albumID, req)
}

func (s *service) GetAlbumTrees(albumID int) ([]Tree, error) {
	return s.repo.GetTreesByAlbumID(albumID)
}

func (s *service) GetTreeDetails(treeID int) (*Tree, error) {
	return s.repo.GetTreeByID(treeID)
}

func (s *service) AddNode(treeID int, req CreateNodeRequest) (int, error) {
	return s.repo.CreateNode(treeID, req)
}

func (s *service) GetNodeDetails(nodeID int) (*TreeNode, error) {
	return s.repo.GetNodeByID(nodeID)
}

func (s *service) UpdateNodeReflection(nodeID int, req UpdateNodeRequest) error {
	return s.repo.UpdateNode(nodeID, req)
}

func (s *service) RemoveNode(nodeID int) error {
	return s.repo.DeleteNode(nodeID)
}