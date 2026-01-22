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
			if Publish_status, ok := aiResult.Payload["publish_status"].(string); ok {
				result.Publish_status = Publish_status
			}
			if creator, ok := aiResult.Payload["creator_id"].(string); ok {
				result.CreatorID = creator
			}
		}

		// If critical fields are missing from payload, query database
		if result.Title == "" {
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
			result.Publish_status = path.Publish_status
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
