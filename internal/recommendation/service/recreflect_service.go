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

func (s *serviceImpl) RecommendPathsForUser(ctx context.Context, userID string, treeID string) (*recmodel.RecommendPathResponse, error) {
	if userID == "" {
		return nil, apperror.NewUnauthorized("Authentication session expired")
	}

	if treeID == "" {
		return nil, apperror.NewBadRequest("please select a specific tree to get recommendations")
	}

	s.logger.InfoContext(ctx, "generating recommendation from tree nodes", "user_id", userID, "tree_id", treeID)

	reflections, currentPathID, err := s.recRepo.GetUserReflectionsByTree(ctx, userID, treeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch tree reflections", "error", err.Error())
		return nil, apperror.NewInternal("could not analyze your reflection history")
	}

	if len(reflections) == 0 {
		return &recmodel.RecommendPathResponse{
			UserPersonaQuery: "No historical data found in this tree.",
			RecommendedPaths: []recmodel.RecommendedPath{},
		}, nil
	}

	var personaBuilder strings.Builder
	personaBuilder.WriteString("Find the best next learning path for a student who just finished a module with the following progress: ")

	for _, ref := range reflections {
		if ref.Summary != "" {
			personaBuilder.WriteString(" [Topic Summary: ")
			personaBuilder.WriteString(ref.Summary)
			personaBuilder.WriteString(". Emotion: ")
			personaBuilder.WriteString(ref.PrimaryEmotion)

			if ref.StrugglePoint != "" {
				personaBuilder.WriteString(". Struggled with: ")
				personaBuilder.WriteString(ref.StrugglePoint)
			}
			personaBuilder.WriteString("] ")
		}
	}

	userPersonaQuery := personaBuilder.String()

	aiReq := aiclient.SearchRequest{
		Query:        userPersonaQuery,
		TopK:         10,
		ResourceType: "learning_paths",
	}

	aiResp, err := s.aiClient.Search(ctx, aiReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "vector search failed", "error", err.Error())
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

		if resultPathID == currentPathID {
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

				if resultPathID == currentPathID {
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
					Reason:              "Recommended based on your reflections and current path.",
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
			}
		}
	}

	if len(finalRecommendations) == 0 {
		return s.getFallbackPopularPaths(ctx, "No relevant reflections found for your current path. Showing top popular paths.")
	}

	return &recmodel.RecommendPathResponse{
		RecommendedPaths: finalRecommendations,
		UserPersonaQuery: userPersonaQuery,
	}, nil
}
