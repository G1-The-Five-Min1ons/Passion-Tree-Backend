package repository

import (
	"context"
	"fmt"
)

func (r *repositoryImpl) UpdateStreak(ctx context.Context, userID string) error {
	query := `
		UPDATE Profile
		SET
			Learning_streak = CASE
				WHEN Last_Active_Date = CAST(GETUTCDATE() AS DATE) THEN Learning_streak
				WHEN Last_Active_Date = CAST(DATEADD(day, -1, GETUTCDATE()) AS DATE) THEN Learning_streak + 1
				ELSE 1
			END,
			Last_Active_Date = CAST(GETUTCDATE() AS DATE)
		WHERE user_id = @p1
	`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("repo.UpdateStreak exec failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repo.UpdateStreak: could not get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("repo.UpdateStreak: no profile found for user %s", userID)
	}

	return nil
}