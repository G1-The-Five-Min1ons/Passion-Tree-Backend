package repository

import (
	"context"
	"fmt"
	"strings"

	"passiontree/internal/onboarding/model"
)

func (r *repositoryImpl) UpsertOnboarding(ctx context.Context, userID string, req model.SaveOnboardingRequest) error {
	subjects := strings.Join(req.Subjects, ",")
	learningStyles := strings.Join(req.LearningStyles, ",")

	query := `
		MERGE INTO onboarding_answers WITH (HOLDLOCK) AS target
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
