package service

import (
	"context"
	"log/slog"

	repoUser "passiontree/internal/auth/repository"
	"passiontree/internal/mission/model"
	"passiontree/internal/mission/repository"
)

type ServiceMission interface {
	CreateTemplate(ctx context.Context, req model.CreateTemplateRequest, userID string) (string, error)
	DeleteTemplate(ctx context.Context, missionID string, userID string) error
	GetMyMissions(ctx context.Context, userID string) ([]model.UserMission, error)
	ProcessMissionEvent(ctx context.Context, userID string, conditionType string) error
	AutoAssignWeeklyMissions(ctx context.Context) error
	AutoAssignWeeklyMissionsByUser(ctx context.Context, userID string) error
	CleanupExpiredMissions(ctx context.Context) error
}

type Service interface {
	ServiceMission
}

type serviceImpl struct {
	repo     repository.RepositoryMission
	repoUser repoUser.RepositoryUser
	logger   *slog.Logger
}

func NewService(repo repository.Repository, repoUser repoUser.RepositoryUser, logger *slog.Logger) Service {
	return &serviceImpl{
		repo:     repo,
		repoUser: repoUser,
		logger:   logger,
	}
}
