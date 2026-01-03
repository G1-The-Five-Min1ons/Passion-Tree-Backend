package service

import (
	"passiontree/internal/learning-path/model"
)

func (s *serviceImpl) AddNode(req model.CreateNodeRequest) (string, error) {
	return s.nodeRepo.CreateNode(req)
}

func (s *serviceImpl) EditNode(nodeID string, req model.UpdateNodeRequest) error {
	return s.nodeRepo.UpdateNode(nodeID, req)
}

func (s *serviceImpl) RemoveNode(nodeID string) error {
	return s.nodeRepo.DeleteNode(nodeID)
}

func (s *serviceImpl) AddMaterial(req model.CreateMaterialRequest) (string, error) {
	return s.nodeRepo.CreateMaterial(req)
}

func (s *serviceImpl) RemoveMaterial(materialID string) error {
	return s.nodeRepo.DeleteMaterial(materialID)
}