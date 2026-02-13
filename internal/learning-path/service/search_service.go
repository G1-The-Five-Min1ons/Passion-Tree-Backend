package service

import (
	"context"
	"database/sql"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/platform/aiclient"
	"strconv"
)

func (s *serviceImpl) SearchLearningPaths(ctx context.Context, req model.SearchPathRequest) (*model.SearchPathResponse, error) {
	s.logger.InfoContext(ctx, "searching learning paths", 
		"query", req.Query, 
		"top_k", req.TopK, 
		"filters", req.Filters,
	)

	if req.Query == "" {
		return nil, apperror.NewBadRequest("search query cannot be empty")
	}

	// Set default TopK
	if req.TopK == 0 {
		req.TopK = 10
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
		s.logger.ErrorContext(ctx, "failed to search via AI service", "error", err, "query", req.Query)
		return nil, apperror.NewInternal("failed to search via AI service: %w", err)
	}

	// If no results, return empty response
	if len(aiResp.Results) == 0 {
		s.logger.InfoContext(ctx, "no matching results from AI search", "query", req.Query)
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
			if Publish_status, ok := aiResult.Payload["publish_status"].(string); ok {
				result.Publish_status = Publish_status
			}
			if creator, ok := aiResult.Payload["creator_id"].(string); ok {
				result.CreatorID = creator
			}
		}

		// If critical fields are missing from payload, query database
		if result.Title == "" {
			s.logger.WarnContext(ctx, "incomplete payload in vector db, fetching from SQL", "path_id", pathID)
			path, err := s.pathRepo.GetLearningPathByID(ctx, pathID)
			if err != nil {
				if err == sql.ErrNoRows {
					s.logger.WarnContext(ctx, "path in vector db not found in SQL database", "path_id", pathID)
					continue
				}
				return nil, apperror.NewInternal("failed to fetch path details: %w", err)
			}

			// Fill in missing fields from database
			result.Title = path.Title
			result.Description = path.Description
			result.CoverImgURL = path.CoverImgURL
			result.Objective = path.Objective
			result.AvgRating = path.Rating
			result.Publish_status = path.Publish_status
			result.CreatorID = path.CreatorID
			result.CreatedAt = path.CreatedAt
			result.UpdatedAt = path.UpdatedAt
		}

		results = append(results, result)
	}

	s.logger.InfoContext(ctx, "search completed", "query", aiResp.Query,  "total_found", len(results))
	
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
		s.logger.ErrorContext(context.Background(), "failed to get collection info", "error", err, "collection", collectionName)
		return nil, apperror.NewInternal("failed to get collection info from AI service: %w", err)
	}

	s.logger.InfoContext(context.Background(), "retrieved collection info successfully", "collection", collectionName)
	return info, nil
}

// SyncLearningPath syncs a single learning path from Azure DB to Qdrant vector database
func (s *serviceImpl) SyncLearningPath(ctx context.Context, pathID string) (*model.SyncPathResponse, error) {
	s.logger.DebugContext(ctx, "SyncLearningPath called", 
		"path_id", pathID, 
		"ai_client_nil", s.aiClient == nil,
	)

	if s.aiClient != nil {
		s.aiClient.DebugClientPointer()
	}
	// Validate pathID
	if pathID == "" {
		return nil, apperror.NewBadRequest("path ID cannot be empty")
	}

	// Get learning path from database
	path, err := s.pathRepo.GetLearningPathByID(ctx, pathID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "sync failed: path not found", "path_id", pathID)
			return nil, apperror.NewNotFound("learning path not found")
		}
		return nil, apperror.NewInternal("failed to fetch learning path from database: %w", err)
	}

	// Prepare metadata for filtering
	metadata := map[string]interface{}{
		"title":           path.Title,
		"description":     path.Description,
		"cover_img_url":   path.CoverImgURL,
		"objective":       path.Objective,
		"avg_rating":      path.Rating,
		"Publish_status ": path.Publish_status,
		"creator_id":      path.CreatorID,
	}

	// Handle nullable time fields
	if path.CreatedAt != nil {
		metadata["created_at"] = path.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if path.UpdatedAt != nil {
		metadata["updated_at"] = path.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	// Create sync request for AI service
	syncReq := aiclient.SyncLearningPathRequest{
		PathID:         pathID,
		Title:          path.Title,
		Description:    path.Description,
		Metadata:       metadata,
		CollectionName: "learning_paths",
	}

	// Call AI service to sync to Qdrant
	if s.aiClient == nil {
		s.logger.ErrorContext(ctx, "sync aborted: AI client is nil")
		return nil, apperror.NewInternal("ai client is not initialized")
	}

	syncResp, err := s.aiClient.SyncLearningPath(ctx, syncReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "vector sync failed", "error", err, "path_id", pathID)
		return nil, apperror.NewInternal("failed to sync learning path to Qdrant: %w", err)
	}

	s.logger.InfoContext(ctx, "sync completed successfully", "path_id", pathID, "msg", syncResp.Message)
	return &model.SyncPathResponse{
		Success: syncResp.Success,
		Message: syncResp.Message,
		PathID:  syncResp.PathID,
	}, nil
}
