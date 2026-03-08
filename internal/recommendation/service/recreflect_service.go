package service

import (
	"context"
	"strconv"
	"strings"

	pathmodel "passiontree/internal/learning-path/model" // ตรวจสอบ import นี้ให้ตรงโปรเจกต์คุณด้วยนะครับ
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/platform/aiclient"
	"passiontree/internal/recommendation/model"
)

func (s *serviceImpl) RecommendPathsForUser(ctx context.Context, userID string, treeID string) (*model.RecommendPathResponse, error) {
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
		return nil, apperror.NewInternal("Recommendation engine is temporarily unavailable")
	}

	var pathIDsToFetch []string
	for _, aiResult := range aiResp.Results {
		resultPathID, ok := s.extractPathID(aiResult.ID)
		if !ok || resultPathID == currentPathID {
			continue
		}
		pathIDsToFetch = append(pathIDsToFetch, resultPathID)
	}

	if len(pathIDsToFetch) == 0 {
		return s.getFallbackPopularPaths(ctx, "No relevant reflections found for your current tree. Showing top popular paths.")
	}

	paths, err := s.pathRepo.GetLearningPathsByIDs(ctx, pathIDsToFetch)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to batch fetch paths from DB", "error", err.Error())
		return s.getFallbackPopularPaths(ctx, "Could not fetch specific paths. Showing top popular paths.")
	}

	pathMap := make(map[string]pathmodel.LearningPath)
	for _, p := range paths {
		pathMap[p.PathID] = p
	}

	var finalRecommendations []model.RecommendedPath
	for _, aiResult := range aiResp.Results {
		resultPathID, ok := s.extractPathID(aiResult.ID)
		if !ok || resultPathID == currentPathID {
			continue
		}

		fullPath, exists := pathMap[resultPathID]
		if !exists {
			s.logger.WarnContext(ctx, "path from AI not found in DB", "path_id", resultPathID)
			continue
		}

		recPath := model.RecommendedPath{
			LearningPath:        fullPath,
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

	if len(finalRecommendations) == 0 {
		return s.getFallbackPopularPaths(ctx, "No relevant reflections found for your current path. Showing top popular paths.")
	}

	return &model.RecommendPathResponse{
		RecommendedPaths: finalRecommendations,
		UserPersonaQuery: userPersonaQuery,
	}, nil
}

func (s *serviceImpl) extractPathID(id any) (string, bool) {
	switch v := id.(type) {
	case string:
		return v, true
	case float64:
		return strconv.Itoa(int(v)), true
	default:
		return "", false
	}
}
