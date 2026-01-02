package repository

import (
    "github.com/google/uuid"
    "passiontree/internal/learning-path/model"
)

func (r *repository) CreateNode(req model.CreateNodeRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node (node_id, title, description, path_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.Title, req.Description, req.PathID)
	return id, err
}

func (r *repository) GetNodesByPathID(pathID string) ([]model.Node, error) {
	query := `SELECT node_id, title, description, path_id FROM node WHERE path_id = ?`
	rows, err := r.db.Query(query, pathID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.NodeID, &n.Title, &n.Description, &n.PathID); err != nil {
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (r *repository) UpdateNode(nodeID string, req model.UpdateNodeRequest) error {
	query := `UPDATE node SET title=?, description=? WHERE node_id=?`
	_, err := r.db.Exec(query, req.Title, req.Description, nodeID)
	return err
}

func (r *repository) DeleteNode(nodeID string) error {
	_, err := r.db.Exec(`DELETE FROM node WHERE node_id = ?`, nodeID)
	return err
}

func (r *repository) CreateMaterial(req model.CreateMaterialRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node_material (material_id, type, url, node_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.Type, req.URL, req.NodeID)
	return id, err
}

func (r *repository) GetMaterialsByNodeID(nodeID string) ([]model.NodeMaterial, error) {
	query := `SELECT material_id, type, url, node_id FROM node_material WHERE node_id = ?`
	rows, err := r.db.Query(query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []model.NodeMaterial
	for rows.Next() {
		var m model.NodeMaterial
		if err := rows.Scan(&m.MaterialID, &m.Type, &m.URL, &m.NodeID); err != nil {
			continue
		}
		materials = append(materials, m)
	}
	return materials, nil
}

func (r *repository) DeleteMaterial(materialID string) error {
	_, err := r.db.Exec(`DELETE FROM node_material WHERE material_id = ?`, materialID)
	return err
}