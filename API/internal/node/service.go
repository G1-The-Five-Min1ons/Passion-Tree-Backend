package learningpath

import "errors"

type Service interface {
	// ... (interface เดิม)
	CreateNode(pathID int, req CreateNodeRequest) (int, error)
	GetNodeDetails(nodeID int) (*Node, error)
	UpdateNode(nodeID int, req UpdateNodeRequest) error
	DeleteNode(nodeID int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateNode(pathID int, req CreateNodeRequest) (int, error) {
	if req.Title == "" {
		return 0, errors.New("node title is required")
	}
	return s.repo.CreateNode(pathID, req)
}

func (s *service) GetNodeDetails(nodeID int) (*Node, error) {
	return s.repo.GetNodeByID(nodeID)
}

func (s *service) UpdateNode(nodeID int, req UpdateNodeRequest) error {
	return s.repo.UpdateNode(nodeID, req)
}

func (s *service) DeleteNode(nodeID int) error {
	return s.repo.DeleteNode(nodeID)
}