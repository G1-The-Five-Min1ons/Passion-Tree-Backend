package repository

import (
	"context"
	"database/sql"
	"fmt"
	"passiontree/internal/reflection/model"

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
		INSERT INTO tree (tree_id, title, status, is_pause, create_at, last_update, album_id)
		VALUES (@p1, @p2, 'active', 0, GETDATE(), GETDATE(), @p3)
	`
	
	_, err = tx.ExecContext(ctx, query, treeID, req.Title, req.AlbumID)
	if err != nil {
		return "", fmt.Errorf("insert tree failed: %w", err)
	}
	
	// Increment tree_count in the album
	updateQuery := `
		UPDATE tree_album
		SET tree_count = tree_count + 1, last_edit = GETDATE()
		WHERE album_id = @p1
	`
	_, err = tx.ExecContext(ctx, updateQuery, req.AlbumID)
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
		SELECT CONVERT(VARCHAR(36), tree_id) as tree_id, title, status, is_pause, create_at, last_update, CONVERT(VARCHAR(36), album_id) as album_id
		FROM tree
		WHERE tree_id = @p1
	`
	
	var tree model.Tree
	err := r.db.QueryRowContext(ctx, query, treeID).Scan(
		&tree.TreeID,
		&tree.Title,
		&tree.Status,
		&tree.IsPause,
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
		SELECT CONVERT(VARCHAR(36), tree_id) as tree_id, title, status, is_pause, create_at, last_update, CONVERT(VARCHAR(36), album_id) as album_id
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
			&tree.Status,
			&tree.IsPause,
			&tree.CreatedAt,
			&tree.LastUpdate,
			&tree.AlbumID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tree: %w", err)
		}
		trees = append(trees, tree)
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
	// First, get the album_id to decrement tree_count
	var albumID string
	getQuery := `SELECT album_id FROM tree WHERE tree_id = @p1`
	err := r.db.QueryRowContext(ctx, getQuery, treeID).Scan(&albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return fmt.Errorf("repo.DeleteTree scan failed: %w", err)
	}
	
	// Delete the tree
	deleteQuery := `DELETE FROM tree WHERE tree_id = @p1`
	result, err := r.db.ExecContext(ctx, deleteQuery, treeID)
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
	_, err = r.db.ExecContext(ctx, updateQuery, albumID)
	if err != nil {
		return fmt.Errorf("failed to update album tree count: %w", err)
	}
	
	return nil
}
