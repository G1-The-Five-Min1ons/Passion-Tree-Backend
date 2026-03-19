package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"passiontree/internal/reflection/model"
	"strings"
	"time"

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
	normalizedDifficulties := strings.ToLower(strings.TrimSpace(req.Difficulties))

	query := `
		INSERT INTO tree (tree_id, title, difficulties, path_id, status, is_pause, node_count, create_at, last_update, album_id)
		VALUES (@p1, @p2, @p3, @p4, 'growing', 0, 0, GETDATE(), GETDATE(), @p5)
	`

	_, err = tx.ExecContext(ctx, query, treeID, req.Title, normalizedDifficulties, req.PathID, req.AlbumID)
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
		       create_at, last_update, CONVERT(VARCHAR(36), album_id) as album_id,
		       last_reflect_at, paused_at, tree_score
		FROM tree
		WHERE tree_id = @p1
	`

	var tree model.Tree
	var lastReflectAt, pausedAt sql.NullTime
	var treeScore sql.NullFloat64
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
		&lastReflectAt,
		&pausedAt,
		&treeScore,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetTreeByID scan failed: %w", err)
	}
	if lastReflectAt.Valid {
		tree.LastReflectAt = &lastReflectAt.Time
	}
	if pausedAt.Valid {
		tree.PausedAt = &pausedAt.Time
	}
	if treeScore.Valid {
		tree.TreeScore = &treeScore.Float64
	}

	return &tree, nil
}

// GetTreesByAlbumID retrieves all trees for a specific album
func (r *repositoryImpl) GetTreesByAlbumID(ctx context.Context, albumID string) ([]model.Tree, error) {
	query := `
		SELECT CONVERT(VARCHAR(36), tree_id) as tree_id, title, difficulties,
		       CONVERT(VARCHAR(36), path_id) as path_id, status, is_pause,
		       ISNULL(node_count, 0) as node_count,
		       create_at, last_update, CONVERT(VARCHAR(36), album_id) as album_id,
		       last_reflect_at, paused_at, tree_score
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
		var lastReflectAt, pausedAt sql.NullTime
		var treeScore sql.NullFloat64
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
			&lastReflectAt,
			&pausedAt,
			&treeScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tree: %w", err)
		}
		if lastReflectAt.Valid {
			tree.LastReflectAt = &lastReflectAt.Time
		}
		if pausedAt.Valid {
			tree.PausedAt = &pausedAt.Time
		}
		if treeScore.Valid {
			tree.TreeScore = &treeScore.Float64
		}
		trees = append(trees, tree)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetTreesByAlbumID row iteration failed: %w", err)
	}

	return trees, nil
}

// UpdateTreeStatus persists the computed status without changing last_update,
// so read-driven status refreshes do not reorder trees unexpectedly.
func (r *repositoryImpl) UpdateTreeStatus(ctx context.Context, treeID string, status string) error {
	query := `
		UPDATE tree
		SET status = @p1
		WHERE tree_id = @p2
	`

	if _, err := r.db.ExecContext(ctx, query, status, treeID); err != nil {
		return fmt.Errorf("failed to update tree status: %w", err)
	}

	return nil
}

// UpdateTree updates an existing tree
func (r *repositoryImpl) UpdateTree(ctx context.Context, treeID string, req model.UpdateTreeRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Get current album_id from DB first to verify actual change
	var currentAlbumID string
	getQuery := `SELECT CAST(album_id AS VARCHAR(36)) FROM tree WHERE tree_id = @p1`
	err = tx.QueryRowContext(ctx, getQuery, treeID).Scan(&currentAlbumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return fmt.Errorf("failed to get current album_id: %w", err)
	}

	// Determine if album is actually being moved based on DB data
	isAlbumChanged := req.AlbumID != "" && currentAlbumID != req.AlbumID

	// Update tree - include album_id only if it's actually changing
	var updateQuery string
	var result sql.Result

	if isAlbumChanged {
		updateQuery = `
			UPDATE tree
			SET title = @p1, album_id = @p2, last_update = GETDATE()
			WHERE tree_id = @p3
		`
		result, err = tx.ExecContext(ctx, updateQuery, req.Title, req.AlbumID, treeID)
	} else {
		updateQuery = `
			UPDATE tree
			SET title = @p1, last_update = GETDATE()
			WHERE tree_id = @p2
		`
		result, err = tx.ExecContext(ctx, updateQuery, req.Title, treeID)
	}

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

	// Update tree_count ONLY if album actually changed (verified from DB)
	if isAlbumChanged {
		// Decrement tree_count in old album
		decrementQuery := `
			UPDATE tree_album
			SET tree_count = tree_count - 1, last_edit = GETDATE()
			WHERE album_id = @p1
		`
		_, err = tx.ExecContext(ctx, decrementQuery, currentAlbumID)
		if err != nil {
			return fmt.Errorf("failed to decrement old album tree_count: %w", err)
		}

		// Increment tree_count in new album
		incrementQuery := `
			UPDATE tree_album
			SET tree_count = tree_count + 1, last_edit = GETDATE()
			WHERE album_id = @p1
		`
		_, err = tx.ExecContext(ctx, incrementQuery, req.AlbumID)
		if err != nil {
			return fmt.Errorf("failed to increment new album tree_count: %w", err)
		}
	}

	return tx.Commit()
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
	getQuery := `SELECT CAST(album_id AS VARCHAR(36)) FROM tree WHERE tree_id = @p1`
	err = tx.QueryRowContext(ctx, getQuery, treeID).Scan(&albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return fmt.Errorf("repo.DeleteTree scan failed: %w", err)
	}

	// Delete reflections linked to this tree's nodes first to satisfy FK_reflect_tree_node
	deleteReflectionsQuery := `
		DELETE r
		FROM Reflect r
		INNER JOIN tree_node tn ON r.tree_node_id = tn.tree_node_id
		WHERE tn.tree_id = @p1
	`
	_, err = tx.ExecContext(ctx, deleteReflectionsQuery, treeID)
	if err != nil {
		return fmt.Errorf("failed to delete reflections for tree nodes: %w", err)
	}

	// Delete all tree nodes first (cascade delete)
	deleteNodesQuery := `DELETE FROM tree_node WHERE tree_id = @p1`
	_, err = tx.ExecContext(ctx, deleteNodesQuery, treeID)
	if err != nil {
		return fmt.Errorf("failed to delete tree nodes: %w", err)
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
		WHERE album_id = CAST(@p1 AS UNIQUEIDENTIFIER)
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

// PauseTree toggles the pause state of a tree.
// Pausing records paused_at = NOW so we know when the pause started.
// Unpausing shifts last_reflect_at forward by the pause duration so the
// progress timer resumes exactly where it left off, then clears paused_at.
func (r *repositoryImpl) PauseTree(ctx context.Context, treeID string, isPause bool) error {
	var query string
	if isPause {
		// Pausing: record the moment we paused.
		query = `
			UPDATE tree
			SET is_pause   = 1,
			    paused_at  = GETDATE(),
			    last_update = GETDATE()
			WHERE tree_id = @p1
		`
	} else {
		// Unpausing: advance last_reflect_at by the pause duration so the
		// effective elapsed time stays unchanged, then clear paused_at.
		query = `
			UPDATE tree
			SET is_pause       = 0,
			    last_reflect_at = CASE
			                          WHEN last_reflect_at IS NOT NULL AND paused_at IS NOT NULL
			                          THEN DATEADD(SECOND, DATEDIFF(SECOND, paused_at, GETDATE()), last_reflect_at)
			                          ELSE last_reflect_at
			                      END,
			    paused_at       = NULL,
			    last_update     = GETDATE()
			WHERE tree_id = @p1
		`
	}

	result, err := r.db.ExecContext(ctx, query, treeID)
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

// GetTreesWithNodesByAlbumID retrieves all trees with their nodes for a specific album ใช้ join มา optimize
func (r *repositoryImpl) GetTreesWithNodesByAlbumID(ctx context.Context, albumID string, userID string) ([]model.TreeResponse, error) {
	// Determine userID for query - if empty, pass NULL to SQL
	var dbUserID interface{}
	if userID != "" {
		dbUserID = userID
	} else {
		dbUserID = nil
	}

	query := `
		SELECT
			CONVERT(VARCHAR(36), t.tree_id) as tree_id,
			t.title as tree_title,
			t.difficulties,
			CONVERT(VARCHAR(36), t.path_id) as path_id,
			t.status,
			t.is_pause,
			ISNULL(t.node_count, 0) as node_count,
			t.create_at,
			t.last_update,
			CONVERT(VARCHAR(36), t.album_id) as album_id,
			t.last_reflect_at,
			t.paused_at,
			CONVERT(VARCHAR(36), tn.tree_node_id) as tree_node_id,
			tn.node_title,
			CONVERT(VARCHAR(36), tn.node_id) as node_id,
			tn.create_at as node_create_at,
			CASE WHEN @p2 IS NULL THEN NULL ELSE np.status END as node_status,
			CASE WHEN @p2 IS NULL THEN NULL ELSE np.complete END as node_complete,
			ISNULL(n.sequence, 0) as sequence,
			CONVERT(VARCHAR(36), r.reflect_id) as reflection_id,
			CASE WHEN n.path_id IS NULL THEN 1 ELSE 0 END as is_standalone,
			t.tree_score
		FROM tree t
		LEFT JOIN Tree_Node tn ON t.tree_id = tn.tree_id
		LEFT JOIN node n ON tn.node_id = n.node_id
		LEFT JOIN node_progress np ON tn.node_id = np.node_id AND np.user_id = @p2
		LEFT JOIN Reflect r ON tn.tree_node_id = r.tree_node_id
		WHERE t.album_id = @p1
		ORDER BY t.last_update DESC, n.sequence ASC, tn.create_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, albumID, dbUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trees with nodes: %w", err)
	}
	defer rows.Close()

	// Map to aggregate nodes per tree
	treeMap := make(map[string]*model.TreeResponse)
	var treeOrder []string

	for rows.Next() {
		var treeID, title, difficulties, pathID, status, treeAlbumID string
		var isPause bool
		var nodeCount int
		var createdAt, lastUpdate time.Time
		var lastReflectAt, pausedAt sql.NullTime

		// Nullable node fields
		var treeNodeID, nodeTitle, nodeID sql.NullString
		var nodeCreatedAt sql.NullTime
		var nodeStatus, nodeComplete sql.NullString
		var reflectionID sql.NullString
		var sequence int
		var isStandalone sql.NullInt64
		var treeScore sql.NullFloat64

		err := rows.Scan(
			&treeID, &title, &difficulties, &pathID, &status, &isPause, &nodeCount,
			&createdAt, &lastUpdate, &treeAlbumID,
			&lastReflectAt,
			&pausedAt,
			&treeNodeID, &nodeTitle, &nodeID, &nodeCreatedAt,
			&nodeStatus, &nodeComplete,
			&sequence,
			&reflectionID,
			&isStandalone,
			&treeScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tree with nodes: %w", err)
		}

		// If tree not yet in map, add it
		if _, exists := treeMap[treeID]; !exists {
			tr := &model.TreeResponse{
				TreeID:       treeID,
				Title:        title,
				Difficulties: difficulties,
				PathID:       pathID,
				Status:       status,
				IsPause:      isPause,
				NodeCount:    nodeCount,
				AlbumID:      treeAlbumID,
				CreatedAt:    createdAt,
				LastUpdate:   lastUpdate,
				Nodes:        []model.TreeNode{},
			}
			if lastReflectAt.Valid {
				tr.LastReflectAt = &lastReflectAt.Time
			}
			if pausedAt.Valid {
				tr.PausedAt = &pausedAt.Time
			}
			if treeScore.Valid {
				tr.TreeScore = &treeScore.Float64
			}
			treeMap[treeID] = tr
			treeOrder = append(treeOrder, treeID)
		}

		// Add node if it exists (LEFT JOIN might have NULL nodes)
		if treeNodeID.Valid {
			node := model.TreeNode{
				TreeNodeID:   treeNodeID.String,
				NodeTitle:    nodeTitle.String,
				NodeID:       nodeID.String,
				TreeID:       treeID,
				Sequence:     sequence,
				IsStandalone: isStandalone.Valid && isStandalone.Int64 == 1,
			}

			if nodeCreatedAt.Valid {
				node.CreatedAt = nodeCreatedAt.Time
			}

			if nodeStatus.Valid {
				status := nodeStatus.String
				node.Status = &status
			}

			if nodeComplete.Valid {
				complete := nodeComplete.String
				node.Complete = &complete
			}

			if reflectionID.Valid {
				reflectID := reflectionID.String
				node.ReflectionID = &reflectID
			}

			treeMap[treeID].Nodes = append(treeMap[treeID].Nodes, node)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetTreesWithNodesByAlbumID row iteration failed: %w", err)
	}

	// Convert map to ordered slice
	var trees []model.TreeResponse
	for _, treeID := range treeOrder {
		trees = append(trees, *treeMap[treeID])
	}

	return trees, nil
}

func (r *repositoryImpl) CalculateAndUpdateTreeScore(ctx context.Context, treeID string) (*float64, error) {

	query := `
		SELECT r.weighted_reflection_score
		FROM tree_node tn
		INNER JOIN (
			SELECT tree_node_id, MAX(create_at) AS latest_at
			FROM Reflect
			GROUP BY tree_node_id
		) latest ON latest.tree_node_id = tn.tree_node_id
		INNER JOIN Reflect r
			ON  r.tree_node_id = latest.tree_node_id
			AND r.create_at    = latest.latest_at
		WHERE tn.tree_id = @p1
	`

	rows, err := r.db.QueryContext(ctx, query, treeID)
	if err != nil {
		return nil, fmt.Errorf("repo.CalculateAndUpdateTreeScore query failed: %w", err)
	}
	defer rows.Close()

	var scores []float64
	for rows.Next() {
		var score float64
		if err := rows.Scan(&score); err != nil {
			return nil, fmt.Errorf("repo.CalculateAndUpdateTreeScore scan failed: %w", err)
		}
		scores = append(scores, score)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.CalculateAndUpdateTreeScore row iteration failed: %w", err)
	}

	if len(scores) == 0 {
		// No reflected nodes yet — clear tree_score
		_, err = r.db.ExecContext(ctx,
			`UPDATE tree SET tree_score = NULL, last_update = GETDATE() WHERE tree_id = @p1`,
			treeID,
		)
		if err != nil {
			return nil, fmt.Errorf("repo.CalculateAndUpdateTreeScore clear score failed: %w", err)
		}
		return nil, nil
	}

	treeScore := computeTreeScore(scores)
	_, err = r.db.ExecContext(ctx,
		`UPDATE tree SET tree_score = @p1, last_update = GETDATE() WHERE tree_id = @p2`,
		treeScore, treeID,
	)
	if err != nil {
		return nil, fmt.Errorf("repo.CalculateAndUpdateTreeScore update failed: %w", err)
	}

	return &treeScore, nil
}

func computeTreeScore(scores []float64) float64 {
	var total float64
	for _, score := range scores {
		total += score
	}
	average := total / float64(len(scores))
	if average < 0 {
		average = 0
	}
	if average > 10 {
		average = 10
	}
	return math.Round(average*100) / 100
}
