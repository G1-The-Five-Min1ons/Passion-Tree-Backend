package service

import (
	"context"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) GetUserHistory(ctx context.Context, userID string) ([]model.HistoryResponse, error) {
    if userID == "" {
        return nil, apperror.NewBadRequest("user_id is required")
    }

    historyList, err := s.historyRepo.GetHistoryByUserID(ctx, userID)
    if err != nil {
        return nil, apperror.NewInternal("history fetch failed: %w", err)
    }

    if historyList == nil {
        return []model.HistoryResponse{}, nil
    }

	s.logger.InfoContext(ctx, "user history retrieved successfully", "user_id", userID, "count", len(historyList))
    return historyList, nil
}