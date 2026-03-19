package repository

import (
	"context"
	"database/sql"
	"fmt"
	"passiontree/internal/reflection/model"
	"strings"

	"github.com/google/uuid"
)

func (r *repositoryImpl) CreateStandaloneNode(ctx context.Context, title string) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	var nextSequence int
	sequenceQuery := `
		SELECT ISNULL(MAX(sequence), 0) + 1
		FROM Node WITH (UPDLOCK, HOLDLOCK)
		WHERE path_id IS NULL
	`

	if err := tx.QueryRowContext(ctx, sequenceQuery).Scan(&nextSequence); err != nil {
		return "", fmt.Errorf("failed to allocate next standalone sequence: %w", err)
	}

	nodeID := uuid.New().String()
	insertQuery := `
		INSERT INTO Node (node_id, title, description, path_id, sequence, link_vdo)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6)
	`

	_, err = tx.ExecContext(ctx, insertQuery, nodeID, title, "Created from reflection tree", sql.NullString{}, nextSequence, "")
	if err != nil {
		return "", fmt.Errorf("failed to create standalone node: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit transaction failed: %w", err)
	}

	return nodeID, nil
}

// AddSingleTreeNode creates a single tree node record (use when adding one custom node)
func (r *repositoryImpl) AddSingleTreeNode(ctx context.Context, req model.CreateTreeNodeRequest) (string, error) {
	treeNodeID := uuid.New().String()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO Tree_Node (tree_node_id, node_title, node_id, tree_id, child_node, create_at)
		VALUES (@p1, @p2, @p3, @p4, @p5, GETDATE())
	`

	_, err = tx.ExecContext(ctx, insertQuery, treeNodeID, req.NodeTitle, req.NodeID, req.TreeID, req.ChildNode)
	if err != nil {
		return "", fmt.Errorf("failed to create tree_node: %w", err)
	}

	updateCountQuery := `
		UPDATE tree
		SET node_count = (
			SELECT COUNT(*)
			FROM Tree_Node
			WHERE tree_id = @p1
		),
		last_update = GETDATE()
		WHERE tree_id = @p1
	`

	if _, err = tx.ExecContext(ctx, updateCountQuery, req.TreeID); err != nil {
		return "", fmt.Errorf("failed to update tree node_count: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("commit transaction failed: %w", err)
	}

	return treeNodeID, nil
}

// GetTreeNodesByTreeID retrieves all tree nodes for a specific tree
func (r *repositoryImpl) GetTreeNodesByTreeID(ctx context.Context, treeID string) ([]model.TreeNode, error) {
	query := `
		SELECT CONVERT(VARCHAR(36), tn.tree_node_id) as tree_node_id,
		       tn.node_title,
		       ISNULL(CONVERT(VARCHAR(36), tn.node_id), '') as node_id,
		       tn.node_score,
		       tn.create_at,
		       CONVERT(VARCHAR(36), tn.tree_id) as tree_id,
		       CASE WHEN tn.child_node IS NOT NULL THEN CONVERT(VARCHAR(36), tn.child_node) ELSE NULL END as child_node,
		       ISNULL(n.sequence, 0) as sequence,
		       CASE WHEN n.path_id IS NULL THEN 1 ELSE 0 END as is_standalone
		FROM Tree_Node tn
		LEFT JOIN node n ON tn.node_id = n.node_id
		WHERE tn.tree_id = @p1
		ORDER BY n.sequence ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, treeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree nodes: %w", err)
	}
	defer rows.Close()
	
	var nodes []model.TreeNode
	for rows.Next() {
		var node model.TreeNode
		var isStandalone int
		err := rows.Scan(
			&node.TreeNodeID,
			&node.NodeTitle,
			&node.NodeID,
			&node.NodeScore,
			&node.CreatedAt,
			&node.TreeID,
			&node.ChildNode,
			&node.Sequence,
			&isStandalone,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tree node: %w", err)
		}
		node.IsStandalone = isStandalone == 1
		nodes = append(nodes, node)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetTreeNodesByTreeID row iteration failed: %w", err)
	}

	return nodes, nil
}

// GetTreeNodeByID retrieves a specific tree node by its ID
func (r *repositoryImpl) GetTreeNodeByID(ctx context.Context, treeNodeID string) (*model.TreeNode, error) {
	query := `
		SELECT CONVERT(VARCHAR(36), tn.tree_node_id) as tree_node_id,
		       tn.node_title,
		       ISNULL(CONVERT(VARCHAR(36), tn.node_id), '') as node_id,
		       tn.node_score,
		       tn.create_at,
		       CONVERT(VARCHAR(36), tn.tree_id) as tree_id,
		       CASE WHEN tn.child_node IS NOT NULL THEN CONVERT(VARCHAR(36), tn.child_node) ELSE NULL END as child_node,
		       ISNULL(n.sequence, 0) as sequence,
		       CASE WHEN n.path_id IS NULL THEN 1 ELSE 0 END as is_standalone
		FROM Tree_Node tn
		LEFT JOIN node n ON tn.node_id = n.node_id
		WHERE tn.tree_node_id = @p1
	`
	
	var node model.TreeNode
	var isStandalone int
	err := r.db.QueryRowContext(ctx, query, treeNodeID).Scan(
		&node.TreeNodeID,
		&node.NodeTitle,
		&node.NodeID,
		&node.NodeScore,
		&node.CreatedAt,
		&node.TreeID,
		&node.ChildNode,
		&node.Sequence,
		&isStandalone,
	)
	node.IsStandalone = isStandalone == 1
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetTreeNodeByID scan failed: %w", err)
	}
	
	return &node, nil
}

// UpdateTreeNode updates a tree node
func (r *repositoryImpl) UpdateTreeNode(ctx context.Context, treeNodeID string, req model.UpdateTreeNodeRequest) error {
	query := `
		UPDATE Tree_Node
		SET node_title = @p1, node_score = @p2, child_node = @p3
		WHERE tree_node_id = @p4
	`
	
	result, err := r.db.ExecContext(ctx, query, req.NodeTitle, req.NodeScore, req.ChildNode, treeNodeID)
	if err != nil {
		return fmt.Errorf("failed to update tree node: %w", err)
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

// DeleteTreeNode deletes a tree node
func (r *repositoryImpl) DeleteTreeNode(ctx context.Context, treeNodeID string) error {
	query := `DELETE FROM Tree_Node WHERE tree_node_id = @p1`
	
	result, err := r.db.ExecContext(ctx, query, treeNodeID)
	if err != nil {
		return fmt.Errorf("failed to delete tree node: %w", err)
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

// CreateTreeNodes creates tree_node records for all nodes in a learning path (bulk operation with transaction)
func (r *repositoryImpl) CreateTreeNodes(ctx context.Context, treeID string, pathID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// First, get all nodes from the learning path
	nodesQuery := `
		SELECT CONVERT(VARCHAR(36), node_id) as node_id, title, sequence
		FROM node
		WHERE path_id = @p1
		ORDER BY sequence ASC
	`
	
	rows, err := tx.QueryContext(ctx, nodesQuery, pathID)
	if err != nil {
		return fmt.Errorf("failed to get nodes for tree: %w", err)
	}
	
	//bulk insert
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
			return fmt.Errorf("failed to scan node: %w", err)
		}
		
		nodesToInsert = append(nodesToInsert, nodeData{
			treeNodeID: uuid.New().String(),
			title:      title,
			nodeID:     nodeID,
		})
	}
	rows.Close()
	
	if len(nodesToInsert) == 0 {
		return fmt.Errorf("no nodes found for path_id: %s", pathID)
	}
	
	// True bulk insert - single query with multiple VALUES
	// Build dynamic query: INSERT INTO ... VALUES (...), (...), (...)
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
		return fmt.Errorf("failed to bulk insert tree_nodes: %w", err)
	}
	
	// Update the node_count in the tree table
	updateQuery := `
		UPDATE tree
		SET node_count = @p1, last_update = GETDATE()
		WHERE tree_id = @p2
	`
	
	_, err = tx.ExecContext(ctx, updateQuery, len(nodesToInsert), treeID)
	if err != nil {
		return fmt.Errorf("failed to update tree node_count: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}
	
	return nil
}

// GetNodesByPathID retrieves all nodes for a specific learning path, ordered by sequence
func (r *repositoryImpl) GetNodesByPathID(ctx context.Context, pathID string) ([]model.TreeNode, error) {
	query := `
		SELECT CONVERT(VARCHAR(36), node_id) as node_id, title, 
		       ISNULL(description, '') as description, 
		       sequence, CONVERT(VARCHAR(36), path_id) as path_id
		FROM node
		WHERE path_id = @p1
		ORDER BY sequence ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, pathID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	defer rows.Close()
	
	var nodes []model.TreeNode
	for rows.Next() {
		var node model.TreeNode
		var title, description, pathID string
		var sequence int
		
		err := rows.Scan(
			&node.NodeID,
			&title,
			&description,
			&sequence,
			&pathID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		
		node.NodeTitle = title
		node.Sequence = sequence
		nodes = append(nodes, node)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetNodesByPathID row iteration failed: %w", err)
	}

	return nodes, nil
}
