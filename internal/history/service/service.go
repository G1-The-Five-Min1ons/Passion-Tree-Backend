package service

import (
	"context"
	"passiontree/internal/history/model"
	"passiontree/internal/history/repository"
	"passiontree/internal/pkg/apperror"
)

type ServiceHistory interface {
	GetUserHistory(ctx context.Context, userID string) ([]model.HistoryResponse, error)
}

type serviceImpl struct {
	repo repository.RepositoryHistory
}

func NewService(repo repository.RepositoryHistory) ServiceHistory {
	return &serviceImpl{
		repo: repo,
	}
}

func (s *serviceImpl) GetUserHistory(ctx context.Context, userID string) ([]model.HistoryResponse, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	historyList, err := s.repo.GetHistoryByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("failed to retrieve history for user %s: %w", userID, err)
	}

	if historyList == nil {
		historyList = []model.HistoryResponse{}
	}

	return historyList, nil
}