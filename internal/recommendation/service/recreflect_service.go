package service

import (
	"context"
	"strconv"
	"strings"

	"passiontree/internal/pkg/apperror"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/model"
)

func (s *serviceImpl) RecommendPathsForUser(ctx context.Context, userID string, treeID string) (*model.RecommendPathResponse, error) {
	if userID == "" || treeID == "" {
		return nil, apperror.NewBadRequest("user_id and tree_id are required")
	}

	s.logger.InfoContext(ctx, "generating recommendation from tree nodes", "user_id", userID, "tree_id", treeID)

	reflections, currentPathID, err := s.recreflectRepo.GetUserReflectionsByTree(ctx, userID, treeID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch tree reflections", "error", err.Error())
		return nil, apperror.NewInternal("failed to fetch user reflections")
	}

	if len(reflections) == 0 {
		return &model.RecommendPathResponse{
			UserPersonaQuery: "No historical data found in this tree.",
			RecommendedPaths: []model.RecommendedPath{},
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

		if resultPathID == currentPathID {
			continue
		}

		recPath := model.RecommendedPath{
			PathID:              resultPathID,
			RecommendationScore: aiResult.Score,
			Reason:              "Based on your recent summaries and struggle points in the previous tree.",
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

	return &model.RecommendPathResponse{
		RecommendedPaths: finalRecommendations,
		UserPersonaQuery: userPersonaQuery,
	}, nil
}
