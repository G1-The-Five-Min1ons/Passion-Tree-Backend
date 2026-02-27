package repository

import (
	"context"
	"fmt"
	"passiontree/internal/learning-path/model"
)

func (r *repositoryImpl) GetUserPathProgress(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error) {
    query := `
        SELECT 
            COUNT(n.node_id) as total_nodes,
            COUNT(CASE WHEN np.complete = 'true' THEN 1 END) as completed_nodes
        FROM node n
        LEFT JOIN node_progress np ON n.node_id = np.node_id AND np.user_id = @p1
        WHERE n.path_id = @p2`

    var totalNodes, completedNodes int
    err := r.db.QueryRowContext(ctx, query, userID, pathID).Scan(&totalNodes, &completedNodes)
    if err != nil {
        return nil, fmt.Errorf("repo.GetUserPathProgress count failed: %w", err)
    }

    var percentage float64
    var status string

    if totalNodes > 0 {
        percentage = (float64(completedNodes) / float64(totalNodes)) * 100
        
        if completedNodes == totalNodes {
            status = "Completed"
        } else {
            status = "In progress"
        }
    } else {
        percentage = 0
        status = "In progress"
    }

    return &model.PathProgressResponse{
        PathID:         pathID,
        UserID:         userID,
        TotalNodes:     totalNodes,
        CompletedNodes: completedNodes,
        Progress:       percentage,
        Status:         status,
    }, nil
}
