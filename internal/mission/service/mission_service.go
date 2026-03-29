package service

import (
	"context"
	"time"

	"passiontree/internal/mission/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) CreateTemplate(ctx context.Context, req model.CreateTemplateRequest, userID string) (string, error) {
	if req.Title == "" || req.RewardXP <= 0 {
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
	missions, err := s.repo.GetUserActiveMissions(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch user missions", "error", err, "user_id", userID)
		return nil, apperror.NewInternal("failed to fetch user missions")
	}
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
		
		// สเตปที่ 1: แจกภารกิจ Common ให้ทุกคนเสมอ
		if commonTemplate, ok := templateMap[model.ConditionCommon]; ok {
			err := s.repo.AssignMissionToUser(ctx, stat.UserID, commonTemplate.MissionID, expireDate)
			if err != nil {
				s.logger.WarnContext(ctx, "failed to assign common mission", "user_id", stat.UserID, "error", err)
			} else {
				missionsAssigned++
			}
		}

		// สเตปที่ 2: แจกภารกิจ Personalized ตามพฤติกรรม
		var personalizedTemplate model.MissionTemplate
		templateFound := false

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

		if templateFound {
			err := s.repo.AssignMissionToUser(ctx, stat.UserID, personalizedTemplate.MissionID, expireDate)
			if err != nil {
				s.logger.WarnContext(ctx, "failed to assign personalized mission", "user_id", stat.UserID, "error", err)
			} else {
				missionsAssigned++
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

	for _, m := range missions {
		newValue := m.CurrentValue + 1
		isCompleted := newValue >= m.TargetValue

		err := s.repo.UpdateMissionProgressAndReward(ctx, m.UserMissionID, userID, newValue, isCompleted, m.RewardXP, m.RewardHeart)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to update mission progress", "error", err, "user_mission_id", m.UserMissionID)
			continue
		}

		if isCompleted {
			s.logger.InfoContext(ctx, "user completed a mission!", "user_id", userID, "mission_title", m.Title, "xp_reward", m.RewardXP)
		}
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
