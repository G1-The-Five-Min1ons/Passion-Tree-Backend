package repository

import (
	"context"
	"database/sql"
	"fmt"

	"passiontree/internal/learning-path/model"

	"github.com/google/uuid"
)

func (r *repositoryImpl) CreateNode(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	return r.createNodeInternal(ctx, r.db, req)
}

func (r *repositoryImpl) createNodeInternal(ctx context.Context, db DBTX, req model.CreateNodeRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node (node_id, title, description, path_id, sequence) VALUES (@p1, @p2, @p3, @p5, @p6)`
	_, err := db.ExecContext(ctx, query, id, req.Title, req.Description, req.PathID, req.Sequence)
	if err != nil {
		return "", fmt.Errorf("repo.CreateNode exec failed: %w", err)
	}
	return id, nil
}

func (r *repositoryImpl) GetNodesByPathID(ctx context.Context, pathID string) ([]model.Node, error) {
	query := `SELECT node_id, title, description, path_id, sequence FROM node WHERE path_id = @p1 ORDER BY sequence ASC`
	rows, err := r.db.QueryContext(ctx, query, pathID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetNodesByPathID query failed: %w", err)
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.NodeID, &n.Title, &n.Description, &n.PathID, &n.Sequence); err != nil {
			return nil, fmt.Errorf("repo.GetNodesByPathID scan failed: %w", err)
		}
		nodes = append(nodes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetNodesByPathID row iteration failed: %w", err)
	}

	return nodes, nil
}

func (r *repositoryImpl) UpdateNode(ctx context.Context, nodeID string, req model.UpdateNodeRequest) error {
	query := `UPDATE node SET title=@p1, description=@p2 WHERE node_id=@p3`
	res, err := r.db.ExecContext(ctx, query, req.Title, req.Description, nodeID)
	if err != nil {
		return fmt.Errorf("repo.UpdateNode exec failed [id=%s]: %w", nodeID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) DeleteNode(ctx context.Context, nodeID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM node WHERE node_id = @p1`, nodeID)
	if err != nil {
		return fmt.Errorf("repo.DeleteNode exec failed [id=%s]: %w", nodeID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) CreateMaterial(ctx context.Context, req model.CreateMaterialRequest) (string, error) {
	return r.createMaterialInternal(ctx, r.db, req)
}

func (r *repositoryImpl) createMaterialInternal(ctx context.Context, db DBTX, req model.CreateMaterialRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO node_material (material_id, type, url, node_id) VALUES (@p1, @p2, @p3, @p4)`
	_, err := db.ExecContext(ctx, query, id, req.Type, req.URL, req.NodeID)
	if err != nil {
		return "", fmt.Errorf("repo.CreateMaterial exec failed: %w", err)
	}
	return id, nil
}

func (r *repositoryImpl) GetMaterialsByNodeID(ctx context.Context, nodeID string) ([]model.NodeMaterial, error) {
	query := `SELECT material_id, type, url, node_id FROM node_material WHERE node_id = @p1`
	rows, err := r.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetMaterialsByNodeID query failed: %w", err)
	}
	defer rows.Close()

	var mats []model.NodeMaterial
	for rows.Next() {
		var m model.NodeMaterial
		if err := rows.Scan(&m.MaterialID, &m.Type, &m.URL, &m.NodeID); err != nil {
			return nil, fmt.Errorf("repo.GetMaterialsByNodeID scan failed: %w", err)
		}
		mats = append(mats, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetMaterialsByNodeID row iteration failed: %w", err)
	}

	return mats, nil
}

func (r *repositoryImpl) DeleteMaterial(ctx context.Context, materialID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM node_material WHERE material_id = @p1`, materialID)
	if err != nil {
		return fmt.Errorf("repo.DeleteMaterial exec failed [id=%s]: %w", materialID, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repositoryImpl) GetNodeByID(ctx context.Context, nodeID string) (*model.Node, error) {
	query := `SELECT node_id, title, description, path_id FROM node WHERE node_id = @p1`

	var n model.Node
	err := r.db.QueryRowContext(ctx, query, nodeID).Scan(&n.NodeID, &n.Title, &n.Description, &n.PathID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetNodeByID scan failed: %w", err)
	}

	materials, err := r.GetMaterialsByNodeID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetNodeByID fetch materials failed: %w", err)
	}
	n.Materials = materials

	return &n, nil
}

func (r *repositoryImpl) UpdateNodeSequence(ctx context.Context, nodeIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo.ReorderNodesTx begin failed: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE node SET sequence = @p1 WHERE node_id = @p2`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("repo.ReorderNodesTx prepare failed: %w", err)
	}
	defer stmt.Close()

	for i, nodeID := range nodeIDs {
		_, err := stmt.ExecContext(ctx, i, nodeID)
		if err != nil {
			return fmt.Errorf("repo.ReorderNodesTx exec failed [id=%s, seq=%d]: %w", nodeID, i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repo.ReorderNodesTx commit failed: %w", err)
	}

	return nil
}

func (r *repositoryImpl) CreateNodeWithContent(ctx context.Context, req model.CreateNodeRequest) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	nodeID, err := r.createNodeInternal(ctx, tx, req)
	if err != nil {
		return "", err
	}

	for _, mat := range req.Materials {
		mat.NodeID = nodeID
		_, err := r.createMaterialInternal(ctx, tx, mat)
		if err != nil {
			return "", err
		}
	}

	for _, qWrapper := range req.Questions {
		qReq := qWrapper.CreateQuestionRequest
		qReq.NodeID = nodeID

		qID, err := r.createQuestionInternal(ctx, tx, qReq)
		if err != nil {
			return "", err
		}

		for _, c := range qWrapper.Choices {
			c.QuestionID = qID
			_, err := r.createChoiceInternal(ctx, tx, c)
			if err != nil {
				return "", err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit failed: %w", err)
	}

	return nodeID, nil
}
