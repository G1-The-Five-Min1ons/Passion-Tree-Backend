package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"passiontree/internal/mission/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) CreateTemplate(ctx context.Context, req model.CreateTemplateRequest, userID string) (string, error) {
	if req.Title == "" || req.RewardXP <= 0 || req.TargetValue <= 0 || req.RewardHeart < 0 {
		return "", apperror.NewBadRequest("invalid template parameters")
	}

	if req.ConditionType != model.ConditionCompleteNode &&
		req.ConditionType != model.ConditionWriteReflect &&
		req.ConditionType != model.ConditionEnrollPath &&
		req.ConditionType != model.ConditionCompletePath &&
		req.ConditionType != model.ConditionCommon {
		return "", apperror.NewBadRequest("invalid condition_type")
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

func (s *serviceImpl) DeleteTemplate(ctx context.Context, missionID string, userID string) error {
	user, _, err := s.repoUser.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user with id '%s' not found", userID)
	}
	if user.Role != "admin" {
		return apperror.NewUnauthorized("User not admin")
	}

	if err := s.repo.DeleteTemplate(ctx, missionID); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete mission template", "mission_id", missionID, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFound("mission template not found or already inactive")
		}
		return apperror.NewInternal("failed to delete mission template: %w", err)
	}

	s.logger.InfoContext(ctx, "mission template deleted", "mission_id", missionID)
	return nil
}

func (s *serviceImpl) GetAllTemplates(ctx context.Context, userID string) ([]model.MissionTemplate, error) {
	user, _, err := s.repoUser.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return nil, apperror.NewNotFound("user with id '%s' not found", userID)
	}
	if user.Role != "admin" {
		return nil, apperror.NewUnauthorized("User not admin")
	}

	templates, err := s.repo.GetAllTemplates(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch all mission templates", "error", err)
		return nil, apperror.NewInternal("failed to fetch mission templates")
	}
	return templates, nil
}

func (s *serviceImpl) GetMyMissions(ctx context.Context, userID string) ([]model.UserMission, error) {
	missions, err := s.repo.GetUserActiveMissions(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch user missions", "error", err, "user_id", userID)
		return nil, apperror.NewInternal("failed to fetch user missions")
	}
	if missions == nil {
		missions = []model.UserMission{}
	}
	return missions, nil
}

func (s *serviceImpl) AutoAssignWeeklyMissionsByUser(ctx context.Context, userID string) error {
	user, _, err := s.repoUser.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.NewInternal("failed to get user by ID: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user with id '%s' not found", userID)
	}
	if user.Role != "admin" {
		return apperror.NewUnauthorized("User not admin")
	}

	return s.AutoAssignWeeklyMissions(ctx)
}

func (s *serviceImpl) AutoAssignWeeklyMissions(ctx context.Context) error {
	s.logger.InfoContext(ctx, "starting intelligent weekly mission auto-assignment")

	templates, err := s.repo.GetActiveTemplates(ctx)
	if err != nil || len(templates) == 0 {
		return apperror.NewInternal("no active templates found or db error")
	}

	var commonTemplates []model.MissionTemplate
	var personalizedTemplates []model.MissionTemplate
	for _, t := range templates {
		if t.ConditionType == model.ConditionCommon {
			commonTemplates = append(commonTemplates, t)
		} else {
			personalizedTemplates = append(personalizedTemplates, t)
		}
	}

	userStats, err := s.repo.GetAllUserBehaviorStats(ctx)
	if err != nil {
		return apperror.NewInternal("failed to fetch user behavior stats")
	}

	allSeenIDs, err := s.repo.GetAllUserSeenMissionIDs(ctx)
	if err != nil {
		return apperror.NewInternal("failed to fetch user seen mission ids")
	}

	expireDate := time.Now().AddDate(0, 0, 7)
	usersProcessed := 0
	missionsAssigned := 0

	for _, stat := range userStats {
		seenIDs := allSeenIDs[stat.UserID]
		if seenIDs == nil {
			seenIDs = make(map[string]bool)
		}

		missionsToAssign := pickWithRotation(commonTemplates, seenIDs, 3)
		missionsToAssign = append(missionsToAssign, pickPersonalized(personalizedTemplates, seenIDs, stat)...)

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

// pickWithRotation เลือก templates โดย unseen ก่อน แล้วค่อย seen จนครบ limit
func pickWithRotation(templates []model.MissionTemplate, seenIDs map[string]bool, limit int) []string {
	var unseen, seen []model.MissionTemplate
	for _, t := range templates {
		if seenIDs[t.MissionID] {
			seen = append(seen, t)
		} else {
			unseen = append(unseen, t)
		}
	}

	var result []string
	for _, t := range unseen {
		if len(result) >= limit {
			break
		}
		result = append(result, t.MissionID)
	}
	for _, t := range seen {
		if len(result) >= limit {
			break
		}
		result = append(result, t.MissionID)
	}
	return result
}

// pickPersonalized เลือก 2 personalized โดย behavior-priority + rotation
func pickPersonalized(templates []model.MissionTemplate, seenIDs map[string]bool, stat model.UserBehaviorStat) []string {
	// จัด behavior-preferred types ก่อน แล้วตามด้วย fallback
	preferredTypes := []string{}
	if stat.NodesDoneLast7Days >= 2 && stat.ReflectsDoneLast7Days == 0 {
		preferredTypes = append(preferredTypes, model.ConditionWriteReflect)
	}
	if stat.PathsDoneLast7Days > 0 {
		preferredTypes = append(preferredTypes, model.ConditionEnrollPath)
	}
	fallbackTypes := []string{model.ConditionCompleteNode, model.ConditionCompletePath, model.ConditionEnrollPath, model.ConditionWriteReflect}

	ordered := orderTemplates(templates, preferredTypes, fallbackTypes)
	return pickWithRotation(ordered, seenIDs, 2)
}

// orderTemplates จัด templates ตาม preferredTypes ก่อน แล้วตาม fallbackTypes (ใช้ทุก template ในแต่ละ type)
func orderTemplates(templates []model.MissionTemplate, preferredTypes, fallbackTypes []string) []model.MissionTemplate {
	byType := make(map[string][]model.MissionTemplate)
	for _, t := range templates {
		byType[t.ConditionType] = append(byType[t.ConditionType], t)
	}

	seenType := make(map[string]bool)
	var ordered []model.MissionTemplate
	for _, ct := range append(preferredTypes, fallbackTypes...) {
		if seenType[ct] {
			continue
		}
		seenType[ct] = true
		ordered = append(ordered, byType[ct]...)
	}
	return ordered
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
