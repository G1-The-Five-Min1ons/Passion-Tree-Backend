package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"passiontree/internal/learning-path/model"
)

func (r *repositoryImpl) CreateNode(req model.CreateNodeRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node (node_id, title, description, path_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.Title, req.Description, req.PathID)
	if err != nil {
		return "", fmt.Errorf("repo.CreateNode exec failed: %w", err)
	}
	return id, nil
}

func (r *repositoryImpl) GetNodesByPathID(pathID string) ([]model.Node, error) {
	query := `SELECT node_id, title, description, path_id FROM node WHERE path_id = ?`
	rows, err := r.db.Query(query, pathID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetNodesByPathID query failed: %w", err)
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.NodeID, &n.Title, &n.Description, &n.PathID); err != nil {
			return nil, fmt.Errorf("repo.GetNodesByPathID scan failed: %w", err)
		}
		nodes = append(nodes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetNodesByPathID row iteration failed: %w", err)
	}

	return nodes, nil
}

func (r *repositoryImpl) UpdateNode(nodeID string, req model.UpdateNodeRequest) error {
	query := `UPDATE node SET title=?, description=? WHERE node_id=?`
	res, err := r.db.Exec(query, req.Title, req.Description, nodeID)
	if err != nil {
		return fmt.Errorf("repo.UpdateNode exec failed [id=%s]: %w", nodeID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) DeleteNode(nodeID string) error {
	res, err := r.db.Exec(`DELETE FROM node WHERE node_id = ?`, nodeID)
	if err != nil {
		return fmt.Errorf("repo.DeleteNode exec failed [id=%s]: %w", nodeID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) CreateMaterial(req model.CreateMaterialRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node_material (material_id, type, url, node_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.Type, req.URL, req.NodeID)
	if err != nil {
		return "", fmt.Errorf("repo.CreateMaterial exec failed: %w", err)
	}
	return id, nil
}

func (r *repositoryImpl) GetMaterialsByNodeID(nodeID string) ([]model.NodeMaterial, error) {
	query := `SELECT material_id, type, url, node_id FROM node_material WHERE node_id = ?`
	rows, err := r.db.Query(query, nodeID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetMaterialsByNodeID query failed: %w", err)
	}
	defer rows.Close()

	var mats []model.NodeMaterial
	for rows.Next() {
		var m model.NodeMaterial
		if err := rows.Scan(&m.MaterialID, &m.Type, &m.URL, &m.NodeID); err == nil {
			mats = append(mats, m)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetMaterialsByNodeID row iteration failed: %w", err)
	}

	return mats, nil
}

func (r *repositoryImpl) DeleteMaterial(materialID string) error {
	res, err := r.db.Exec(`DELETE FROM node_material WHERE material_id = ?`, materialID)
	if err != nil {
		return fmt.Errorf("repo.DeleteMaterial exec failed [id=%s]: %w", materialID, err)
	}
	
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}