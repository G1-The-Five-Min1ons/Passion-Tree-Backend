package model

import "time"

const (
	ConditionCompleteNode = "COMPLETE_NODE"
	ConditionWriteReflect = "WRITE_REFLECT"
	ConditionEnrollPath   = "ENROLL_PATH"
	ConditionCompletePath = "COMPLETE_PATH"
	ConditionCommon       = "COMMON_ROUTINE"
)

type MissionTemplate struct {
	MissionID     string     `json:"mission_id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	RewardXP      int64      `json:"reward_xp"`
	RewardHeart   int64      `json:"reward_heart"`
	ConditionType string     `json:"condition_type"`
	TargetValue   int        `json:"target_value"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

type CreateTemplateRequest struct {
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description"`
	RewardXP      int64  `json:"reward_xp" binding:"required"`
	RewardHeart   int64  `json:"reward_heart"`
	ConditionType string `json:"condition_type" binding:"required"`
	TargetValue   int    `json:"target_value" binding:"required"`
}

type UserMission struct {
	UserMissionID string     `json:"user_mission_id"`
	UserID        string     `json:"user_id"`
	MissionID     string     `json:"mission_id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	RewardXP      int64      `json:"reward_xp"`
	RewardHeart   int64      `json:"reward_heart"`
	CurrentValue  int        `json:"current_value"`
	TargetValue   int        `json:"target_value"`
	Status        string     `json:"status"`
	ExpireAt      *time.Time `json:"expire_at"`
	CompleteAt    *time.Time `json:"complete_at"`
}

type UserBehaviorStat struct {
	UserID                string
	NodesDoneLast7Days    int
	ReflectsDoneLast7Days int
	PathsDoneLast7Days    int
}
