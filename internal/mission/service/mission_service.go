package service

import (
	"context"
	"time"

	"passiontree/internal/mission/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) CreateTemplate(ctx context.Context, req model.CreateTemplateRequest, userID string) (string, error) {
	validConditions := map[string]bool{
		model.ConditionCompleteNode: true,
		model.ConditionWriteReflect: true,
		model.ConditionEnrollPath:   true,
		model.ConditionCompletePath: true,
		model.ConditionCommon:       true,
	}
	if req.Title == "" || req.RewardXP <= 0 || req.TargetValue <= 0 || !validConditions[req.ConditionType] {
		return "", apperror.NewBadRequest("invalid template parameters")
	}

	user, _, err := s.repoUser.GetUserByID(ctx, userID)
	if err != nil {
		return "", apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return "", apperror.NewNotFound("user with id '%s' not found", userID)
	}
	if user.Role != "admin" {
		return "", apperror.NewUnauthorized("User not admin")
	}

	id, err := s.repo.CreateTemplate(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create mission template", "error", err)
		return "", apperror.NewInternal("database error during template creation")
	}

	s.logger.InfoContext(ctx, "mission template created", "mission_id", id)
	return id, nil
}

func (s *serviceImpl) GetMyMissions(ctx context.Context, userID string) ([]model.UserMission, error) {
	startedAt := time.Now()
	s.logger.InfoContext(ctx, "service.GetMyMissions request",
		"user_id", userID,
	)

	missions, err := s.repo.GetUserActiveMissions(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch user missions", "error", err, "user_id", userID)
		return nil, apperror.NewInternal("failed to fetch user missions")
	}
	if missions == nil {
		missions = []model.UserMission{}
	}

	s.logger.InfoContext(ctx, "service.GetMyMissions response",
		"user_id", userID,
		"missions_count", len(missions),
		"elapsed_ms", time.Since(startedAt).Milliseconds(),
	)

	return missions, nil
}

func (s *serviceImpl) AutoAssignWeeklyMissions(ctx context.Context) error {
	s.logger.InfoContext(ctx, "starting intelligent weekly mission auto-assignment")

	templates, err := s.repo.GetActiveTemplates(ctx)
	if err != nil || len(templates) == 0 {
		return apperror.NewInternal("no active templates found or db error")
	}

	templateMap := make(map[string]model.MissionTemplate)
	for _, t := range templates {
		if _, exists := templateMap[t.ConditionType]; !exists {
			templateMap[t.ConditionType] = t
		}
	}

	userStats, err := s.repo.GetAllUserBehaviorStats(ctx)
	if err != nil {
		return apperror.NewInternal("failed to fetch user behavior stats")
	}

	expireDate := time.Now().AddDate(0, 0, 7)
	usersProcessed := 0
	missionsAssigned := 0

	for _, stat := range userStats {
		activeMissions, err := s.repo.GetUserActiveMissions(ctx, stat.UserID)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to get active missions", "user_id", stat.UserID, "error", err)
			continue
		}
		var missionsToAssign []string

		if commonTemplate, ok := templateMap[model.ConditionCommon]; ok {
			missionsToAssign = append(missionsToAssign, commonTemplate.MissionID)
		}

		var personalizedTemplate model.MissionTemplate
		templateFound := false

		if len(activeMissions) == 0 {
			if stat.NodesDoneLast7Days >= 2 && stat.ReflectsDoneLast7Days == 0 {
				if t, ok := templateMap[model.ConditionWriteReflect]; ok {
					personalizedTemplate = t
					templateFound = true
				}
			}

			if !templateFound && stat.PathsDoneLast7Days > 0 {
				if t, ok := templateMap[model.ConditionEnrollPath]; ok {
					personalizedTemplate = t
					templateFound = true
				}
			}

			if !templateFound {
				if t, ok := templateMap[model.ConditionCompleteNode]; ok {
					personalizedTemplate = t
					templateFound = true
				}
			}
		}

		if templateFound {
			missionsToAssign = append(missionsToAssign, personalizedTemplate.MissionID)
		}

		if len(missionsToAssign) > 0 {
			err := s.repo.AssignMissionsToUser(ctx, stat.UserID, missionsToAssign, expireDate)
			if err != nil {
				s.logger.WarnContext(ctx, "failed to assign missions to user", "user_id", stat.UserID, "error", err)
			} else {
				missionsAssigned += len(missionsToAssign)
			}
		}

		usersProcessed++
	}

	s.logger.InfoContext(ctx, "intelligent weekly mission auto-assignment completed",
		"users_processed", usersProcessed,
		"missions_assigned", missionsAssigned,
	)
	return nil
}

func (s *serviceImpl) ProcessMissionEvent(ctx context.Context, userID string, conditionType string) error {
	missions, err := s.repo.GetActiveMissionsByCondition(ctx, userID, conditionType)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get active missions for event", "error", err, "user_id", userID)
		return apperror.NewInternal("database error during mission event processing")
	}

	if len(missions) == 0 {
		return nil
	}

	var missionsToUpdate []model.UserMission
	var totalRewardXP int64 = 0
	var totalRewardHeart int64 = 0

	for _, m := range missions {
		m.CurrentValue += 1
		isCompleted := m.CurrentValue >= m.TargetValue

		if isCompleted {
			m.Status = "completed"
			totalRewardXP += m.RewardXP
			totalRewardHeart += m.RewardHeart
			s.logger.InfoContext(ctx, "user completed a mission!", "user_id", userID, "mission_title", m.Title)
		}

		missionsToUpdate = append(missionsToUpdate, m)
	}

	err = s.repo.BatchUpdateMissionProgressAndReward(ctx, userID, missionsToUpdate, totalRewardXP, totalRewardHeart)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to batch update mission progress", "error", err, "user_id", userID)
		return apperror.NewInternal("failed to update mission progress")
	}

	return nil
}

func (s *serviceImpl) CleanupExpiredMissions(ctx context.Context) error {
	s.logger.InfoContext(ctx, "starting cleanup of expired and unfinished missions")

	err := s.repo.DeleteExpiredUnfinishedMissions(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to cleanup expired missions", "error", err)
		return apperror.NewInternal("database error during mission cleanup")
	}

	s.logger.InfoContext(ctx, "successfully cleaned up expired missions")
	return nil
}
