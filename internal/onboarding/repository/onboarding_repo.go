package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"passiontree/internal/onboarding/model"
)

func (r *repositoryImpl) UpsertOnboarding(ctx context.Context, userID string, req model.SaveOnboardingRequest) error {
	subjects := strings.Join(req.Subjects, ",")
	learningStyles := strings.Join(req.LearningStyles, ",")

	query := `
		MERGE INTO onboarding_answers AS target
		USING (SELECT @p1 AS user_id) AS source ON target.user_id = source.user_id
		WHEN MATCHED THEN
			UPDATE SET
				subjects         = @p2,
				knowledge_level  = @p3,
				motivation       = @p4,
				daily_goal       = @p5,
				learning_styles  = @p6,
				reflection_habit = @p7,
				updated_at       = GETUTCDATE()
		WHEN NOT MATCHED THEN
			INSERT (id, user_id, subjects, knowledge_level, motivation, daily_goal, learning_styles, reflection_habit, created_at, updated_at)
			VALUES (NEWID(), @p1, @p2, @p3, @p4, @p5, @p6, @p7, GETUTCDATE(), GETUTCDATE());
	`

	_, err := r.db.ExecContext(ctx, query,
		userID,
		subjects,
		req.KnowledgeLevel,
		req.Motivation,
		req.DailyGoal,
		learningStyles,
		req.ReflectionHabit,
	)
	if err != nil {
		return fmt.Errorf("upsert onboarding failed (user: %s): %w", userID, err)
	}
	return nil
}

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
