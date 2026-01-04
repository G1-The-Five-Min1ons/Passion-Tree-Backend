package service

import (
	"database/sql"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) AddNode(req model.CreateNodeRequest) (string, error) {
	if req.Title == "" {
		return "", apperror.NewBadRequest("node title is required")
	}
	
	id, err := s.nodeRepo.CreateNode(req)
	if err != nil {
		return "", apperror.NewInternal(err)
	}
	return id, nil
}

func (s *serviceImpl) EditNode(nodeID string, req model.UpdateNodeRequest) error {
	if nodeID == "" {
		return apperror.NewBadRequest("node_id is required")
	}
	if req.Title == "" &&
		req.Description == ""{
		return apperror.NewBadRequest("request is required")
	}
	if err := s.nodeRepo.UpdateNode(nodeID, req); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot update: node id '%s' not found", nodeID)
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) RemoveNode(nodeID string) error {
	if nodeID == "" {
		return apperror.NewBadRequest("node_id is required")
	}
	if err := s.nodeRepo.DeleteNode(nodeID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot delete: node id '%s' not found", nodeID)
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) AddMaterial(req model.CreateMaterialRequest) (string, error) {
	if req.Type == "" || req.URL == "" {
		return "", apperror.NewBadRequest("material type and url are required")
	}

	id, err := s.nodeRepo.CreateMaterial(req)
	if err != nil {
		return "", apperror.NewInternal(err)
	}
	return id, nil
}

func (s *serviceImpl) RemoveMaterial(materialID string) error {
	if materialID == "" {
		return apperror.NewBadRequest("material_id is required")
	}
	if err := s.nodeRepo.DeleteMaterial(materialID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot delete: material id '%s' not found", materialID)
		}
		return apperror.NewInternal(err)
	}
	return nil
}