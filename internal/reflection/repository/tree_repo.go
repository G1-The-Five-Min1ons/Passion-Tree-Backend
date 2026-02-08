package repository

import (
	"context"
	"database/sql"
	"fmt"
	"passiontree/internal/reflection/model"
	"strings"

	"github.com/google/uuid"
)

// CreateTree creates a new tree in the database
func (r *repositoryImpl) CreateTree(ctx context.Context, req model.CreateTreeRequest) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	treeID := uuid.New().String()
	
	query := `
		INSERT INTO tree (tree_id, title, difficulties, path_id, status, is_pause, node_count, create_at, last_update, album_id)
		VALUES (@p1, @p2, @p3, @p4, 'active', 0, 0, GETDATE(), GETDATE(), @p5)
	`
	
	_, err = tx.ExecContext(ctx, query, treeID, req.Title, req.Difficulties, req.PathID, req.AlbumID)
	if err != nil {
		return "", fmt.Errorf("insert tree failed: %w", err)
	}
	
	// Create tree nodes from learning path nodes
	nodesQuery := `
		SELECT CONVERT(VARCHAR(36), node_id) as node_id, title, sequence
		FROM node
		WHERE path_id = @p1
		ORDER BY sequence ASC
	`
	
	rows, err := tx.QueryContext(ctx, nodesQuery, req.PathID)
	if err != nil {
		return "", fmt.Errorf("failed to get nodes for tree: %w", err)
	}
	
	// Collect all nodes for bulk insert
	type nodeData struct {
		treeNodeID string
		title      string
		nodeID     string
	}
	var nodesToInsert []nodeData
	
	for rows.Next() {
		var nodeID, title string
		var sequence int
		
		if err := rows.Scan(&nodeID, &title, &sequence); err != nil {
			rows.Close()
			return "", fmt.Errorf("failed to scan node: %w", err)
		}
		
		nodesToInsert = append(nodesToInsert, nodeData{
			treeNodeID: uuid.New().String(),
			title:      title,
			nodeID:     nodeID,
		})
	}
	rows.Close()
	
	if len(nodesToInsert) == 0 {
		return "", fmt.Errorf("no nodes found for path_id: %s", req.PathID)
	}
	
	// True bulk insert - single query with multiple VALUES
	var valueStrings []string
	var valueArgs []interface{}
	paramCount := 0
	
	for _, node := range nodesToInsert {
		valueStrings = append(valueStrings, fmt.Sprintf("(@p%d, @p%d, @p%d, @p%d, GETDATE())",
			paramCount+1, paramCount+2, paramCount+3, paramCount+4))
		valueArgs = append(valueArgs, node.treeNodeID, node.title, node.nodeID, treeID)
		paramCount += 4
	}
	
	bulkInsertQuery := fmt.Sprintf(`
		INSERT INTO Tree_Node (tree_node_id, node_title, node_id, tree_id, create_at)
		VALUES %s
	`, strings.Join(valueStrings, ","))
	
	_, err = tx.ExecContext(ctx, bulkInsertQuery, valueArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to bulk insert tree_nodes: %w", err)
	}
	
	nodeCount := len(nodesToInsert)
	
	// Update node_count in tree
	updateTreeQuery := `
		UPDATE tree
		SET node_count = @p1
		WHERE tree_id = @p2
	`
	_, err = tx.ExecContext(ctx, updateTreeQuery, nodeCount, treeID)
	if err != nil {
		return "", fmt.Errorf("failed to update tree node_count: %w", err)
	}
	
	// Increment tree_count in the album
	updateAlbumQuery := `
		UPDATE tree_album
		SET tree_count = tree_count + 1, last_edit = GETDATE()
		WHERE album_id = @p1
	`
	_, err = tx.ExecContext(ctx, updateAlbumQuery, req.AlbumID)
	if err != nil {
		return "", fmt.Errorf("update album tree count failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit transaction failed: %w", err)
	}
	
	return treeID, nil
}

// GetTreeByID retrieves a tree by its ID
func (r *repositoryImpl) GetTreeByID(ctx context.Context, treeID string) (*model.Tree, error) {
	query := `
		SELECT CONVERT(VARCHAR(36), tree_id) as tree_id, title, difficulties, 
		       CONVERT(VARCHAR(36), path_id) as path_id, status, is_pause, 
		       ISNULL(node_count, 0) as node_count,
		       create_at, last_update, CONVERT(VARCHAR(36), album_id) as album_id
		FROM tree
		WHERE tree_id = @p1
	`
	
	var tree model.Tree
	err := r.db.QueryRowContext(ctx, query, treeID).Scan(
		&tree.TreeID,
		&tree.Title,
		&tree.Difficulties,
		&tree.PathID,
		&tree.Status,
		&tree.IsPause,
		&tree.NodeCount,
		&tree.CreatedAt,
		&tree.LastUpdate,
		&tree.AlbumID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetTreeByID scan failed: %w", err)
	}

	return &tree, nil
}

// GetTreesByAlbumID retrieves all trees for a specific album
func (r *repositoryImpl) GetTreesByAlbumID(ctx context.Context, albumID string) ([]model.Tree, error) {
	query := `
		SELECT CONVERT(VARCHAR(36), tree_id) as tree_id, title, difficulties, 
		       CONVERT(VARCHAR(36), path_id) as path_id, status, is_pause, 
		       ISNULL(node_count, 0) as node_count,
		       create_at, last_update, CONVERT(VARCHAR(36), album_id) as album_id
		FROM tree
		WHERE album_id = @p1
		ORDER BY last_update DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trees: %w", err)
	}
	defer rows.Close()
	
	var trees []model.Tree
	for rows.Next() {
		var tree model.Tree
		err := rows.Scan(
			&tree.TreeID,
			&tree.Title,
			&tree.Difficulties,
			&tree.PathID,
			&tree.Status,
			&tree.IsPause,
			&tree.NodeCount,
			&tree.CreatedAt,
			&tree.LastUpdate,
			&tree.AlbumID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tree: %w", err)
		}
		trees = append(trees, tree)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetTreesByAlbumID row iteration failed: %w", err)
	}

	return trees, nil
}

// UpdateTree updates an existing tree
func (r *repositoryImpl) UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
	query := `
		UPDATE tree
		SET title = @p1, status = @p2, is_pause = @p3, last_update = GETDATE()
		WHERE tree_id = @p4
	`
	
	result, err := r.db.ExecContext(ctx, query, req.Title, req.Status, req.IsPause, treeID)
	if err != nil {
		return fmt.Errorf("failed to update tree: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

// DeleteTree deletes a tree by its ID
func (r *repositoryImpl) DeleteTree(ctx context.Context, treeID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// First, get the album_id to decrement tree_count
	var albumID string
	getQuery := `SELECT album_id FROM tree WHERE tree_id = @p1`
	err = tx.QueryRowContext(ctx, getQuery, treeID).Scan(&albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return fmt.Errorf("repo.DeleteTree scan failed: %w", err)
	}
	
	// Delete the tree
	deleteQuery := `DELETE FROM tree WHERE tree_id = @p1`
	result, err := tx.ExecContext(ctx, deleteQuery, treeID)
	if err != nil {
		return fmt.Errorf("failed to delete tree: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	// Decrement tree_count in the album
	updateQuery := `
		UPDATE tree_album
		SET tree_count = tree_count - 1, last_edit = GETDATE()
		WHERE album_id = @p1
	`
	_, err = tx.ExecContext(ctx, updateQuery, albumID)
	if err != nil {
		return fmt.Errorf("failed to update album tree count: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}
	
	return nil
}

// PauseTree toggles the pause state of a tree
func (r *repositoryImpl) PauseTree(ctx context.Context, treeID string, isPause bool) error {
	query := `
		UPDATE tree
		SET is_pause = @p1, last_update = GETDATE()
		WHERE tree_id = @p2
	`
	
	result, err := r.db.ExecContext(ctx, query, isPause, treeID)
	if err != nil {
		return fmt.Errorf("failed to pause tree: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}
