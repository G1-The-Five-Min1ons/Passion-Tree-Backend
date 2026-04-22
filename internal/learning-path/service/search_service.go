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

	if s.aiClient == nil {
		s.logger.ErrorContext(ctx, "search aborted: AI client is nil")
		return nil, apperror.NewInternal("ai client is not initialized")
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

	s.logger.InfoContext(ctx, "search completed", "query", aiResp.Query, "total_found", len(results))

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

	if s.aiClient == nil {
		return nil, apperror.NewInternal("ai client is not initialized")
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
		"title":            path.Title,
		"description":      path.Description,
		"cover_img_url":    path.CoverImgURL,
		"objective":        path.Objective,
		"avg_rating":       path.Rating,
		"publish_status ":  path.Publish_status,
		"creator_id":       path.CreatorID,
		"creator_name":     path.CreatorName,
		"creator_username": path.CreatorUsername,
		"instructor":       path.Instructor,
		"modules":          path.Modules,
		"student":          path.Students,
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

// buildSyncRequest converts a LearningPath row into the payload expected by the AI sync endpoint.
func buildSyncRequest(path model.LearningPath) aiclient.SyncLearningPathRequest {
	metadata := map[string]interface{}{
		"title":            path.Title,
		"description":      path.Description,
		"cover_img_url":    path.CoverImgURL,
		"objective":        path.Objective,
		"avg_rating":       path.Rating,
		"publish_status":   path.Publish_status,
		"creator_id":       path.CreatorID,
		"creator_name":     path.CreatorName,
		"creator_username": path.CreatorUsername,
		"instructor":       path.Instructor,
		"modules":          path.Modules,
		"student":          path.Students,
	}
	if path.CreatedAt != nil {
		metadata["created_at"] = path.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if path.UpdatedAt != nil {
		metadata["updated_at"] = path.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return aiclient.SyncLearningPathRequest{
		PathID:         path.PathID,
		Title:          path.Title,
		Description:    path.Description,
		Metadata:       metadata,
		CollectionName: "learning_paths",
	}
}

// BulkSyncLearningPaths pushes every learning path in SQL to Qdrant in a single call.
func (s *serviceImpl) BulkSyncLearningPaths(ctx context.Context) (*aiclient.BulkSyncResponse, error) {
	if s.aiClient == nil {
		return nil, apperror.NewInternal("ai client is not initialized")
	}

	paths, err := s.pathRepo.GetAllLearningPath(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch paths for bulk sync", "error", err)
		return nil, apperror.NewInternal("failed to fetch learning paths: %w", err)
	}

	requests := make([]aiclient.SyncLearningPathRequest, 0, len(paths))
	for _, p := range paths {
		requests = append(requests, buildSyncRequest(p))
	}

	resp, err := s.aiClient.BulkSyncLearningPaths(ctx, aiclient.BulkSyncRequest{
		LearningPaths:  requests,
		CollectionName: "learning_paths",
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "bulk sync to AI failed", "error", err)
		return nil, apperror.NewInternal("failed to bulk sync learning paths: %w", err)
	}

	s.logger.InfoContext(ctx, "bulk sync completed", "total", resp.Total, "succeeded", resp.Succeeded, "failed", resp.Failed)
	return resp, nil
}

// ReconcileLearningPaths removes Qdrant points that no longer exist in SQL and upserts ones that are missing.
func (s *serviceImpl) ReconcileLearningPaths(ctx context.Context) (*model.ReconcileResponse, error) {
	if s.aiClient == nil {
		return nil, apperror.NewInternal("ai client is not initialized")
	}

	sqlPaths, err := s.pathRepo.GetAllLearningPath(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch paths for reconcile", "error", err)
		return nil, apperror.NewInternal("failed to fetch learning paths: %w", err)
	}

	sqlIDs := make(map[string]model.LearningPath, len(sqlPaths))
	for _, p := range sqlPaths {
		sqlIDs[p.PathID] = p
	}

	qdrantIDs, err := s.aiClient.ListCollectionIDs(ctx, "learning_paths")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list Qdrant ids", "error", err)
		return nil, apperror.NewInternal("failed to list Qdrant ids: %w", err)
	}

	qdrantSet := make(map[string]struct{}, len(qdrantIDs.IDs))
	for _, id := range qdrantIDs.IDs {
		qdrantSet[id] = struct{}{}
	}

	report := &model.ReconcileResponse{
		Success:       true,
		SQLTotal:      len(sqlPaths),
		QdrantTotal:   qdrantIDs.Total,
		StaleDeleted:  []string{},
		MissingSynced: []string{},
		Errors:        []string{},
	}

	// Stale: in Qdrant but not in SQL
	for id := range qdrantSet {
		if _, ok := sqlIDs[id]; ok {
			continue
		}
		if _, err := s.aiClient.SyncDeletePath(ctx, id, "learning_paths"); err != nil {
			report.Success = false
			report.Errors = append(report.Errors, fmt.Sprintf("delete %s: %v", id, err))
			continue
		}
		report.StaleDeleted = append(report.StaleDeleted, id)
	}

	// Missing: in SQL but not in Qdrant — bulk upsert
	missing := make([]aiclient.SyncLearningPathRequest, 0)
	for id, p := range sqlIDs {
		if _, ok := qdrantSet[id]; ok {
			continue
		}
		missing = append(missing, buildSyncRequest(p))
		report.MissingSynced = append(report.MissingSynced, id)
	}
	if len(missing) > 0 {
		resp, err := s.aiClient.BulkSyncLearningPaths(ctx, aiclient.BulkSyncRequest{
			LearningPaths:  missing,
			CollectionName: "learning_paths",
		})
		if err != nil {
			report.Success = false
			report.Errors = append(report.Errors, fmt.Sprintf("bulk upsert: %v", err))
		} else if resp.Failed > 0 {
			report.Success = false
			report.Errors = append(report.Errors, resp.Errors...)
		}
	}

	report.Message = fmt.Sprintf("reconcile finished: deleted=%d synced=%d errors=%d",
		len(report.StaleDeleted), len(report.MissingSynced), len(report.Errors))
	s.logger.InfoContext(ctx, "reconcile completed",
		"sql_total", report.SQLTotal,
		"qdrant_total", report.QdrantTotal,
		"stale_deleted", len(report.StaleDeleted),
		"missing_synced", len(report.MissingSynced),
		"errors", len(report.Errors),
	)
	return report, nil
}
