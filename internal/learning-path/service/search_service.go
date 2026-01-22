package service

import (
	"context"
	"database/sql"
	"fmt"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/platform/aiclient"
	"strconv"
)

func (s *serviceImpl) SearchLearningPaths(ctx context.Context, req model.SearchPathRequest) (*model.SearchPathResponse, error) {
	if req.Query == "" {
		return nil, apperror.NewBadRequest("search query cannot be empty")
	}

	// Set default TopK
	if req.TopK == 0 {
		req.TopK = 7
	}

	aiReq := aiclient.SearchRequest{
		Query:        req.Query,
		TopK:         req.TopK,
		Filters:      req.Filters,
		ResourceType: "learning_paths",
	}

	// Call AI service to get results with payload
	aiResp, err := s.aiClient.Search(ctx, aiReq)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Errorf("failed to search via AI service: %w", err))
	}

	// If no results, return empty response
	if len(aiResp.Results) == 0 {
		return &model.SearchPathResponse{
			Query:   aiResp.Query,
			Total:   0,
			Results: []model.SearchPathResult{},
		}, nil
	}

	// Process results from AI service
	results := make([]model.SearchPathResult, 0, len(aiResp.Results))

	for _, aiResult := range aiResp.Results {
		// Extract PathID from ID field
		var pathID string
		if id, ok := aiResult.ID.(float64); ok {
			pathID = strconv.Itoa(int(id))
		} else if id, ok := aiResult.ID.(string); ok {
			pathID = id
		} else if id, ok := aiResult.ID.(int); ok {
			pathID = strconv.Itoa(id)
		} else {
			continue
		}

		// Build result from payload
		result := model.SearchPathResult{
			PathID: pathID,
			Score:  aiResult.Score,
		}

		// Extract data from payload if available
		if aiResult.Payload != nil {
			if title, ok := aiResult.Payload["title"].(string); ok {
				result.Title = title
			}
			if desc, ok := aiResult.Payload["description"].(string); ok {
				result.Description = desc
			}
			if cover, ok := aiResult.Payload["cover_img_url"].(string); ok {
				result.CoverImgURL = cover
			}
			if obj, ok := aiResult.Payload["objective"].(string); ok {
				result.Objective = obj
			}
			if rating, ok := aiResult.Payload["avg_rating"].(float64); ok {
				result.AvgRating = rating
			}
			if status, ok := aiResult.Payload["status"].(string); ok {
				result.Status = status
			}
			if creator, ok := aiResult.Payload["creator_id"].(string); ok {
				result.CreatorID = creator
			}
		}

		// If critical fields are missing from payload, query database
		if result.Title == "" || result.Description == "" {
			path, err := s.pathRepo.GetLearnningPathByID(ctx, pathID)
			if err != nil {
				if err == sql.ErrNoRows {
					// Skip if path not found in database
					continue
				}
				return nil, apperror.NewInternal(fmt.Errorf("failed to fetch path details: %w", err))
			}

			// Fill in missing fields from database
			result.Title = path.Title
			result.Description = path.Description
			result.CoverImgURL = path.CoverImgURL
			result.Objective = path.Objective
			result.AvgRating = path.AvgRating
			result.Status = path.Status
			result.CreatorID = path.CreatorID
			result.CreatedAt = path.CreatedAt
			result.UpdatedAt = path.UpdatedAt
		}

		results = append(results, result)
	}

	return &model.SearchPathResponse{
		Query:   aiResp.Query,
		Total:   len(results),
		Results: results,
	}, nil
}

// GetCollectionInfo retrieves debug information about a collection from AI service
func (s *serviceImpl) GetCollectionInfo(collectionName string) (*aiclient.CollectionInfoResponse, error) {
	if collectionName == "" {
		return nil, apperror.NewBadRequest("collection name cannot be empty")
	}

	// Call AI service to get collection info
	info, err := s.aiClient.GetCollectionInfo(collectionName)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Errorf("failed to get collection info from AI service: %w", err))
	}

	return info, nil
}

// SyncLearningPath syncs a single learning path from Azure DB to Qdrant vector database
func (s *serviceImpl) SyncLearningPath(pathID string) (*model.SyncPathResponse, error) {
	// Validate pathID
	if pathID == "" {
		return nil, apperror.NewBadRequest("path ID cannot be empty")
	}

	// Get learning path from database
	path, err := s.pathRepo.GetLearnningPathByID(pathID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("learning path not found")
		}
		return nil, apperror.NewInternal(fmt.Errorf("failed to fetch learning path from database: %w", err))
	}

	// Convert pathID to int
	pathIDInt, err := strconv.Atoi(pathID)
	if err != nil {
		return nil, apperror.NewBadRequest("invalid path ID format")
	}

	// Prepare metadata for filtering
	metadata := map[string]interface{}{
		"title":         path.Title,
		"description":   path.Description,
		"cover_img_url": path.CoverImgURL,
		"objective":     path.Objective,
		"avg_rating":    path.AvgRating,
		"status":        path.Status,
		"creator_id":    path.CreatorID,
		"created_at":    path.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":    path.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Create sync request for AI service
	syncReq := aiclient.SyncLearningPathRequest{
		PathID:         pathIDInt,
		Title:          path.Title,
		Description:    path.Description,
		Metadata:       metadata,
		CollectionName: "learning_paths",
	}

	// Call AI service to sync to Qdrant
	syncResp, err := s.aiClient.SyncLearningPath(syncReq)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Errorf("failed to sync learning path to Qdrant: %w", err))
	}

	// Return response
	return &model.SyncPathResponse{
		Success: syncResp.Success,
		Message: syncResp.Message,
		PathID:  syncResp.PathID,
	}, nil
}
