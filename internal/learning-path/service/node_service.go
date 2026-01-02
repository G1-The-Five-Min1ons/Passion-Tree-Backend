package service

import (
	"passiontree/internal/learning-path/model"
)

func (s *service) AddNode(req model.CreateNodeRequest) (string, error) {
	return s.nodeRepo.CreateNode(req)
}

func (s *service) EditNode(nodeID string, req model.UpdateNodeRequest) error {
	return s.nodeRepo.UpdateNode(nodeID, req)
}

func (s *service) RemoveNode(nodeID string) error {
	return s.nodeRepo.DeleteNode(nodeID)
}

func (s *service) AddMaterial(req model.CreateMaterialRequest) (string, error) {
	return s.nodeRepo.CreateMaterial(req)
}

func (s *service) RemoveMaterial(materialID string) error {
	return s.nodeRepo.DeleteMaterial(materialID)
}