package service

import (
	"context"
	"strconv"
	"strings"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/platform/aiclient"
	recmodel "passiontree/internal/recommendation/model"
)

func (s *serviceImpl) RecommendHomePathsForUser(ctx context.Context, userID string) (*recmodel.RecommendPathResponse, error) {
	if userID == "" {
		return nil, apperror.NewUnauthorized("Authentication session expired")
	}

	s.logger.InfoContext(ctx, "generating home recommendations", "user_id", userID)

	enrolledPaths, err := s.recRepo.GetUserEnrolledPathsForRec(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch user enrolled paths", "error", err.Error())
		return nil, apperror.NewInternal("could not retrieve your learning historys")
	}

	enrolledMap := make(map[string]bool)
	for _, p := range enrolledPaths {
		enrolledMap[p.PathID] = true
	}

	if len(enrolledPaths) == 0 {
		return s.getFallbackPopularPaths(ctx, "No enrollment history. Showing top popular paths.")
	}

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
		return nil, apperror.NewInternal("Recommendation engine is temporarily unavailable")
	}

	var pathIDsToFetch []string
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
		pathIDsToFetch = append(pathIDsToFetch, resultPathID)
	}

	var finalRecommendations []recmodel.RecommendedPath

	if len(pathIDsToFetch) > 0 {
		paths, err := s.pathRepo.GetLearningPathsByIDs(ctx, pathIDsToFetch)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to batch fetch paths from DB", "error", err.Error())
		} else {
			pathMap := make(map[string]model.LearningPath)
			for _, p := range paths {
				pathMap[p.PathID] = p
			}

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

				fullPath, exists := pathMap[resultPathID]
				if !exists {
					s.logger.WarnContext(ctx, "path from AI not found in DB", "path_id", resultPathID)
					continue
				}

				recPath := recmodel.RecommendedPath{
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
		}
	}

	if len(finalRecommendations) == 0 {
		return s.getFallbackPopularPaths(ctx, "AI yielded no new results. Showing top popular paths.")
	}

	return &recmodel.RecommendPathResponse{
		RecommendedPaths: finalRecommendations,
		UserPersonaQuery: userPersonaQuery,
	}, nil
}

// หาที่อยู่ให้ด้วยก็ดี
func (s *serviceImpl) getFallbackPopularPaths(ctx context.Context, message string) (*recmodel.RecommendPathResponse, error) {
	topPaths, err := s.recRepo.GetTopPopularPaths(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch fallback popular paths", "error", err.Error())
		return nil, apperror.NewInternal("failed to fetch popular paths")
	}

	return &recmodel.RecommendPathResponse{
		UserPersonaQuery: message,
		RecommendedPaths: topPaths,
	}, nil
}
