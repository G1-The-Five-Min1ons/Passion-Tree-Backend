package repository

import (
	"context"
	"database/sql"
	"time"

	"passiontree/internal/connection"
	"passiontree/internal/mission/model"
)

type RepositoryMission interface {
	// Admin (Template)
	CreateTemplate(ctx context.Context, req model.CreateTemplateRequest) (string, error)
	GetActiveTemplates(ctx context.Context) ([]model.MissionTemplate, error)

	// User (Transaction)
	GetUserActiveMissions(ctx context.Context, userID string) ([]model.UserMission, error)
	AssignMissionsToUser(ctx context.Context, userID string, missionIDs []string, expireAt time.Time) error
	GetActiveMissionsByCondition(ctx context.Context, userID string, conditionType string) ([]model.UserMission, error)

	// สำหรับ Background Job
	GetAllActiveUsers(ctx context.Context) ([]string, error)
	GetAllUserBehaviorStats(ctx context.Context) ([]model.UserBehaviorStat, error)
	DeleteExpiredUnfinishedMissions(ctx context.Context) error
	BatchUpdateMissionProgressAndReward(ctx context.Context, userID string, missions []model.UserMission, totalXP int64, totalHeart int64) error
}

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Database interface {
	DBTX
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
}

type Repository interface {
	RepositoryMission
}

type repositoryImpl struct {
	db Database
}

func NewRepository(ds connection.Database) Repository {
	return &repositoryImpl{
		db: ds.GetDB(),
	}
}
