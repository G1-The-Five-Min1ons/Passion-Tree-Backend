package repository_test

import (
	"context"
	"passiontree/internal/recommendation/model"
)

type Repopository struct {
	GetUserReflectionsByTreeFunc func(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error)
}

func (m *Repopository) GetUserReflectionsByTree(ctx context.Context, userID string, treeID string) ([]model.UserReflection, string, error) {
	if m.GetUserReflectionsByTreeFunc != nil {
		return m.GetUserReflectionsByTreeFunc(ctx, userID, treeID)
	}
	return nil, "", nil
}
