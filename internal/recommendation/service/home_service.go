package service

import (
	"context"
	"strconv"
	"strings"

	"passiontree/internal/pkg/apperror"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/model"
)

func (s *serviceImpl) RecommendHomePathsForUser(ctx context.Context, userID string) (*model.RecommendPathResponse, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	s.logger.InfoContext(ctx, "generating home recommendations", "user_id", userID)

	enrolledPaths, err := s.recRepo.GetUserEnrolledPathsForRec(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch user enrolled paths", "error", err.Error())
		return nil, apperror.NewInternal("failed to fetch enrolled paths")
	}

	enrolledMap := make(map[string]bool)
	for _, p := range enrolledPaths {
		enrolledMap[p.PathID] = true
	}

	if len(enrolledPaths) == 0 {
		topPaths, err := s.recRepo.GetTopPopularPaths(ctx)
		if err != nil {
			return nil, apperror.NewInternal("failed to fetch popular paths")
		}
		return &model.RecommendPathResponse{
			UserPersonaQuery: "No enrollment history. Showing top popular paths.",
			RecommendedPaths: topPaths,
		}, nil
	}

	// สร้าง Persona Query ส่งให้ AI Vector Search
	var personaBuilder strings.Builder
	personaBuilder.WriteString("Recommend similar learning paths for a student who is interested in and has studied the following topics: ")

	limit := len(enrolledPaths)
	personaBuilder.Grow(100 + (limit * 150))
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		personaBuilder.WriteString("[")
		personaBuilder.WriteString(enrolledPaths[i].Title)
		personaBuilder.WriteString(": ")
		personaBuilder.WriteString(enrolledPaths[i].Description)
		personaBuilder.WriteString("] ")
	}

	userPersonaQuery := personaBuilder.String()

	aiReq := aiclient.SearchRequest{
		Query:        userPersonaQuery,
		TopK:         15,
		ResourceType: "learning_paths",
	}

	aiResp, err := s.aiClient.Search(ctx, aiReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "vector search failed in home recommendation", "error", err.Error())
		return nil, apperror.NewInternal("vector search failed")
	}

	var finalRecommendations []model.RecommendedPath

	for _, aiResult := range aiResp.Results {
		var resultPathID string
		if id, ok := aiResult.ID.(float64); ok {
			resultPathID = strconv.Itoa(int(id))
		} else if id, ok := aiResult.ID.(string); ok {
			resultPathID = id
		} else {
			continue
		}

		if enrolledMap[resultPathID] {
			continue
		}

		fullPath, err := s.pathRepo.GetLearningPathByID(ctx, resultPathID)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to get path details from DB", "path_id", resultPathID, "error", err.Error())
			continue
		}

		recPath := model.RecommendedPath{
			PathID:              fullPath.PathID,
			Title:               fullPath.Title,
			Description:         fullPath.Description,
			CoverImgURL:         fullPath.CoverImgURL,
			Objective:           fullPath.Objective,
			Rating:              fullPath.Rating,
			Publish_status:      fullPath.Publish_status,
			CreatedAt:           fullPath.CreatedAt,
			UpdatedAt:           fullPath.UpdatedAt,
			CreatorID:           fullPath.CreatorID,
			Instructor:          fullPath.Instructor,
			Modules:             fullPath.Modules,
			Students:            fullPath.Students,
			RecommendationScore: aiResult.Score,
			Reason:              "Recommended based on your current enrolled learning paths.",
		}

		if aiResult.Payload != nil {
			if title, ok := aiResult.Payload["title"].(string); ok {
				recPath.Title = title
			}
			if cover, ok := aiResult.Payload["cover_img_url"].(string); ok {
				recPath.CoverImgURL = cover
			}
			if obj, ok := aiResult.Payload["objective"].(string); ok {
				recPath.Objective = obj
			}
		}

		finalRecommendations = append(finalRecommendations, recPath)

		if len(finalRecommendations) == 5 {
			break
		}
	}

	// ถ้า AI กรองไปกรองมาแล้วไม่เหลือผลลัพธ์เลย (เช่น Path ในระบบมีน้อยและเคยเรียนหมดแล้ว) ให้ Fallback
	if len(finalRecommendations) == 0 {
		topPaths, _ := s.recRepo.GetTopPopularPaths(ctx)
		return &model.RecommendPathResponse{
			UserPersonaQuery: "AI yielded no new results. Showing top popular paths.",
			RecommendedPaths: topPaths,
		}, nil
	}

	return &model.RecommendPathResponse{
		RecommendedPaths: finalRecommendations,
		UserPersonaQuery: userPersonaQuery,
	}, nil
}
