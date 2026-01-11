package service

import (
	"passiontree/internal/history/model"
	"passiontree/internal/history/repository"
	"passiontree/internal/pkg/apperror"
)

type ServiceHistory interface {
	GetUserHistory(userID string) ([]model.HistoryResponse, error)
}

type serviceImpl struct {
	repo repository.RepositoryHistory
}

func NewService(repo repository.RepositoryHistory) ServiceHistory {
	return &serviceImpl{
		repo: repo,
	}
}

func (s *serviceImpl) GetUserHistory(userID string) ([]model.HistoryResponse, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	historyList, err := s.repo.GetHistoryByUserID(userID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	if historyList == nil {
		historyList = []model.HistoryResponse{}
	}

	return historyList, nil
}