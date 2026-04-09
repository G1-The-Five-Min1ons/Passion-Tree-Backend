package repository

import (
	"context"
	"database/sql"
	"fmt"

	"passiontree/internal/onboarding/model"
)

func (r *repositoryImpl) GetOnboardingByUserID(ctx context.Context, userID string) (*model.OnboardingData, error) {
	query := `
		SELECT
			CONVERT(VARCHAR(36), id)      AS id,
			CONVERT(VARCHAR(36), user_id) AS user_id,
			subjects,
			knowledge_level,
			motivation,
			daily_goal,
			learning_styles,
			reflection_habit,
			created_at,
			updated_at
		FROM onboarding_answers
		WHERE user_id = @p1
	`

	row := r.db.QueryRowContext(ctx, query, userID)

	var d model.OnboardingData
	err := row.Scan(
		&d.ID,
		&d.UserID,
		&d.Subjects,
		&d.KnowledgeLevel,
		&d.Motivation,
		&d.DailyGoal,
		&d.LearningStyles,
		&d.ReflectionHabit,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get onboarding failed (user: %s): %w", userID, err)
	}
	return &d, nil
}
