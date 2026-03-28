package repository

import (
	"context"
	"fmt"
)

// XP reward constants
const (
	XPPerNodeComplete = 50
	XPPerPathComplete = 200
)

// AddXPAndRecalcLevel atomically adds XP to a user's profile and recalculates their level and rank.
func (r *repositoryImpl) AddXPAndRecalcLevel(ctx context.Context, userID string, xpAmount int) error {
	query := `
		UPDATE Profile 
		SET 
			xp = ISNULL(xp, 0) + @p2,
			level = (ISNULL(xp, 0) + @p2) / 1000 + 1,
			rank_name = CASE
				WHEN (ISNULL(xp, 0) + @p2) / 1000 + 1 >= 20 THEN 'Legend'
				WHEN (ISNULL(xp, 0) + @p2) / 1000 + 1 >= 15 THEN 'Master'
				WHEN (ISNULL(xp, 0) + @p2) / 1000 + 1 >= 10 THEN 'Scholar'
				WHEN (ISNULL(xp, 0) + @p2) / 1000 + 1 >= 5  THEN 'Explorer'
				ELSE 'Beginner'
			END,
			updated_at = GETDATE()
		WHERE user_id = @p1
	`

	result, err := r.db.ExecContext(ctx, query, userID, xpAmount)
	if err != nil {
		return fmt.Errorf("repo.AddXPAndRecalcLevel exec failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repo.AddXPAndRecalcLevel rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no profile found for user_id: %s", userID)
	}

	return nil
}
