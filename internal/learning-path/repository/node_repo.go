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
	query := `INSERT INTO node (node_id, title, description, path_id, sequence, link_vdo) VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`
	_, err := db.ExecContext(ctx, query, id, req.Title, req.Description, req.PathID, req.Sequence, req.Link_vdo)
	if err != nil {
		return "", fmt.Errorf("repo.CreateNode exec failed: %w", err)
	}
	return id, nil
}

func (r *repositoryImpl) GetNodesByPathID(ctx context.Context, pathID string, userID string) ([]model.Node, error) {
	var dbUserID sql.NullString
	if userID != "" {
		dbUserID = sql.NullString{String: userID, Valid: true}
	}

	query := `
		SELECT 
			CONVERT(VARCHAR(36), n.node_id) as node_id, 
			ISNULL(n.title, '') as title, 
			ISNULL(n.description, '') as description, 
			CONVERT(VARCHAR(36), n.path_id) as path_id, 
			ISNULL(n.sequence, 0) as sequence,
			CASE WHEN @p2 IS NULL THEN 'locked' ELSE ISNULL(Progress.status, 'locked') END as status,
			CASE WHEN @p2 IS NULL THEN 'null' ELSE ISNULL(Progress.complete, 'null') END as complete,
			ISNULL(n.link_vdo, 'null') as link_vdo
		FROM node n
		LEFT JOIN (
			SELECT node_id, user_id, 
			       MAX(status) as status, 
			       MAX(complete) as complete 
			FROM node_progress 
			WHERE user_id = @p2
			GROUP BY node_id, user_id
		) Progress ON n.node_id = Progress.node_id
		WHERE n.path_id = @p1 
		ORDER BY n.sequence ASC`

	rows, err := r.db.QueryContext(ctx, query, pathID, dbUserID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetNodesByPathID query failed: %w", err)
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.NodeID, &n.Title, &n.Description, &n.PathID, &n.Sequence, &n.Status, &n.Complete, &n.Link_vdo); err != nil {
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	queryNode := `UPDATE node SET title=@p1, description=@p2, link_vdo=@p3 WHERE node_id=@p4`
	_, err = tx.ExecContext(ctx, queryNode, req.Title, req.Description, req.Link_vdo, nodeID)
	if err != nil {
		return fmt.Errorf("update node failed: %w", err)
	}

	if req.Materials != nil {
		_, err = tx.ExecContext(ctx, `DELETE FROM node_material WHERE node_id = @p1`, nodeID)
		if err != nil {
			return fmt.Errorf("delete old materials failed: %w", err)
		}

		for _, mat := range *req.Materials {
			newMatID := uuid.New().String()
			queryMat := `INSERT INTO node_material (material_id, type, url, node_id) VALUES (@p1, @p2, @p3, @p4)`
			_, err = tx.ExecContext(ctx, queryMat, newMatID, mat.Type, mat.URL, nodeID)
			if err != nil {
				return fmt.Errorf("insert new material failed: %w", err)
			}
		}
	}

	return tx.Commit()
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
	query := `SELECT CONVERT(VARCHAR(36), material_id) as material_id, type, url, CONVERT(VARCHAR(36), node_id) as node_id FROM node_material WHERE node_id = @p1`
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

func (r *repositoryImpl) GetNodeByID(ctx context.Context, nodeID string, userID string) (*model.Node, error) {
	var dbUserID sql.NullString
	if userID != "" {
		dbUserID = sql.NullString{String: userID, Valid: true}
	}

	query := `
		SELECT 
			CONVERT(VARCHAR(36), n.node_id) as node_id, 
			ISNULL(n.title, '') as title, 
			ISNULL(n.description, '') as description, 
			CONVERT(VARCHAR(36), n.path_id) as path_id,
			CASE WHEN @p2 IS NULL THEN 'null' ELSE ISNULL(Progress.status, 'locked') END as status,
			CASE WHEN @p2 IS NULL THEN 'null' ELSE ISNULL(Progress.complete, 'null') END as complete,
			ISNULL(n.link_vdo, 'null') as link_vdo
		FROM node n
		LEFT JOIN (
			SELECT node_id, user_id, 
			       MAX(status) as status, 
			       MAX(complete) as complete 
			FROM node_progress 
			WHERE user_id = @p2
			GROUP BY node_id, user_id
		) Progress ON n.node_id = Progress.node_id
		WHERE n.node_id = @p1`

	var n model.Node
	err := r.db.QueryRowContext(ctx, query, nodeID, dbUserID).Scan(
		&n.NodeID,
		&n.Title,
		&n.Description,
		&n.PathID,
		&n.Status,
		&n.Complete,
		&n.Link_vdo,
	)

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
	return r.CreateNodeWithContentInternal(ctx, r.db, req)
}

func (r *repositoryImpl) CreateNodeWithContentInternal(ctx context.Context, db Database, req model.CreateNodeRequest) (string, error) {
	tx, err := db.BeginTx(ctx, nil)
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

func (r *repositoryImpl) UpdateNodeProgressStatus(ctx context.Context, nodeID string, userID string) error {
	// UPSERT: update if row exists, insert with 'active' status if it doesn't
	// (can happen if node was added to the path after the user enrolled)
	query := `
		IF EXISTS (SELECT 1 FROM node_progress WHERE node_id = @p1 AND user_id = @p2)
			UPDATE node_progress
			SET status = 'active', updated_at = GETDATE()
			WHERE node_id = @p1 AND user_id = @p2
		ELSE
			INSERT INTO node_progress (progress_id, user_id, node_id, status, updated_at, complete)
			VALUES (NEWID(), @p2, @p1, 'active', GETDATE(), 'false')
	`
	_, err := r.db.ExecContext(ctx, query, nodeID, userID)
	if err != nil {
		return fmt.Errorf("repo.UpdateNodeProgressStatus exec failed: %w", err)
	}
	return nil
}

func (r *repositoryImpl) UpdateNodeProgressCompletion(ctx context.Context, nodeID string, userID string) error {
	// UPSERT: update if row exists, insert as completed if it doesn't
	query := `
		IF EXISTS (SELECT 1 FROM node_progress WHERE node_id = @p1 AND user_id = @p2)
			UPDATE node_progress
			SET complete = 'true', status = 'active', updated_at = GETDATE()
			WHERE node_id = @p1 AND user_id = @p2
		ELSE
			INSERT INTO node_progress (progress_id, user_id, node_id, status, updated_at, complete)
			VALUES (NEWID(), @p2, @p1, 'active', GETDATE(), 'true')
	`
	_, err := r.db.ExecContext(ctx, query, nodeID, userID)
	if err != nil {
		return fmt.Errorf("repo.UpdateNodeProgressCompletion exec failed: %w", err)
	}
	return nil
}

func (r *repositoryImpl) UpdateNodeProgress(ctx context.Context, nodeID string, userID string, status string) error {
	queryCheck := `SELECT COUNT(1) FROM node_progress WHERE node_id = @p1 AND user_id = @p2`
	var count int
	if err := r.db.QueryRowContext(ctx, queryCheck, nodeID, userID).Scan(&count); err != nil {
		return fmt.Errorf("failed to check existing node progress: %w", err)
	}

	if count > 0 {
		queryUpdate := `UPDATE node_progress SET status = @p1, updated_at = GETUTCDATE() WHERE node_id = @p2 AND user_id = @p3`
		_, err := r.db.ExecContext(ctx, queryUpdate, status, nodeID, userID)
		if err != nil {
			return fmt.Errorf("failed to update node progress: %w", err)
		}
	} else {
		queryInsert := `
			INSERT INTO node_progress (node_id, user_id, status, created_at, updated_at) 
			VALUES (@p1, @p2, @p3, GETUTCDATE(), GETUTCDATE())`
		_, err := r.db.ExecContext(ctx, queryInsert, nodeID, userID, status)
		if err != nil {
			return fmt.Errorf("failed to insert node progress: %w", err)
		}
	}

	return nil
}
