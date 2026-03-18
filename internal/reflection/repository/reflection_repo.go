package repository

import (
	"context"
	"database/sql"
	"fmt"
	"passiontree/internal/reflection/model"

	"github.com/google/uuid"
)

func (r *repositoryImpl) CreateReflection(ctx context.Context, req model.CreateReflectionRequest, summary, sentimentAnalysis string, primaryEmotion *string, strugglePoint string, aiConfidentScore, reflectionScore, weightedReflectionScore float64) (string, error) {
	id := uuid.New().String()

	// Convert primary emotion to string for database
	var primaryEmotionStr sql.NullString
	if primaryEmotion != nil {
		primaryEmotionStr = sql.NullString{String: *primaryEmotion, Valid: true}
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `INSERT INTO Reflect
		(reflect_id, reflect_score, reflect_description, reflect, progress_score, challenge_score, 
		summary, sentiment_analysis, primary_emotion, struggle_point, ai_confident_score, reflection_score, weighted_reflection_score,
		create_at, tree_node_id) 
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11, @p12, @p13, GETDATE(), @p14)`

	_, err = tx.ExecContext(ctx, insertQuery,
		id,
		req.FeelScore,
		req.LearningReflect,
		req.MoodReflect,
		req.ProgressScore,
		req.ChallengeScore,
		summary,
		sentimentAnalysis,
		primaryEmotionStr,
		strugglePoint,
		aiConfidentScore,
		reflectionScore,
		weightedReflectionScore,
		req.TreeNodeID,
	)

	if err != nil {
		return "", fmt.Errorf("repo.CreateReflection exec failed: %w", err)
	}

	// Sync last_reflect_at on the parent tree using step-based recovery:
	//   died => no change (stays dead)
	//   dying => step back to just inside fading zone
	//   fading / growing => reset to NOW (becomes growing)
	//
	// Thresholds in seconds (must match service/tree_status.go):
	//   easy:   fading=2592000(30d) dying=5184000(60d) died=7776000(90d)
	//   medium: fading=604800(7d)   dying=1209600(14d) died=1814400(21d)
	//   hard:   fading=86400(1d)    dying=172800(2d)   died=259200(3d)
	//
	// Only update when the tree is NOT paused — paused trees freeze their timer.
	syncQuery := `
		UPDATE t
		SET t.last_reflect_at = CASE
		    -- DIED: no recovery
		    WHEN calc.elapsed_sec IS NOT NULL AND (
		             (t.difficulties = 'easy'   AND calc.elapsed_sec >= 7776000)
		          OR (t.difficulties = 'medium' AND calc.elapsed_sec >= 1814400)
		          OR (t.difficulties = 'hard'   AND calc.elapsed_sec >= 259200)
		         ) THEN t.last_reflect_at

		    -- DYING → FADING: set last_reflect_at to fading_threshold + 1 sec ago
		    WHEN calc.elapsed_sec IS NOT NULL AND t.difficulties = 'easy'
		         AND calc.elapsed_sec >= 5184000
		         THEN DATEADD(SECOND, -2592001, dt.now_ts)   -- 30d+1s ago

		    WHEN calc.elapsed_sec IS NOT NULL AND t.difficulties = 'medium'
		         AND calc.elapsed_sec >= 1209600
		         THEN DATEADD(SECOND, -604801, dt.now_ts)    -- 7d+1s ago

		    WHEN calc.elapsed_sec IS NOT NULL AND t.difficulties = 'hard'
		         AND calc.elapsed_sec >= 172800
		         THEN DATEADD(SECOND, -86401, dt.now_ts)     -- 1d+1s ago

		    -- FADING or GROWING: reset to NOW → becomes growing
		    ELSE dt.now_ts
		END,
		t.status = CASE
		    WHEN calc.elapsed_sec IS NOT NULL AND (
		             (t.difficulties = 'easy'   AND calc.elapsed_sec >= 7776000)
		          OR (t.difficulties = 'medium' AND calc.elapsed_sec >= 1814400)
		          OR (t.difficulties = 'hard'   AND calc.elapsed_sec >= 259200)
		         ) THEN 'died'
		    WHEN calc.elapsed_sec IS NOT NULL AND t.difficulties = 'easy'
		         AND calc.elapsed_sec >= 5184000
		         THEN 'fading'
		    WHEN calc.elapsed_sec IS NOT NULL AND t.difficulties = 'medium'
		         AND calc.elapsed_sec >= 1209600
		         THEN 'fading'
		    WHEN calc.elapsed_sec IS NOT NULL AND t.difficulties = 'hard'
		         AND calc.elapsed_sec >= 172800
		         THEN 'fading'
		    ELSE 'growing'
		END,
		t.last_update = dt.now_ts
		FROM tree t
		INNER JOIN tree_node tn ON tn.tree_id = t.tree_id
		CROSS APPLY (SELECT GETDATE() AS now_ts) dt
		CROSS APPLY (
			SELECT CASE
				WHEN t.last_reflect_at IS NULL THEN NULL
				ELSE DATEDIFF(SECOND, t.last_reflect_at, dt.now_ts)
			END AS elapsed_sec
		) calc
		WHERE tn.tree_node_id = @p1
		  AND t.is_pause = 0
	`
	if _, err = tx.ExecContext(ctx, syncQuery, req.TreeNodeID); err != nil {
		return "", fmt.Errorf("repo.CreateReflection sync last_reflect_at failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("repo.CreateReflection commit failed: %w", err)
	}

	return id, nil
}

func (r *repositoryImpl) GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error) {
	query := `SELECT
		CONVERT(VARCHAR(36), reflect_id) as reflect_id,
		reflect_score, 
		reflect_description, 
		reflect, 
		progress_score, 
		challenge_score, 
		summary,
		sentiment_analysis,
		primary_emotion,
		struggle_point,
		ai_confident_score,
		reflection_score,
		weighted_reflection_score,
		create_at, 
		CONVERT(VARCHAR(36), tree_node_id) as tree_node_id 
		FROM Reflect
		WHERE reflect_id = @p1`

	var ref model.Reflection
	var primaryEmotionStr sql.NullString
	err := r.db.QueryRowContext(ctx, query, reflectID).Scan(
		&ref.ReflectID,
		&ref.ReflectScore,
		&ref.ReflectDescription,
		&ref.Reflect,
		&ref.ProgressScore,
		&ref.ChallengeScore,
		&ref.Summary,
		&ref.SentimentAnalysis,
		&primaryEmotionStr,
		&ref.StrugglePoint,
		&ref.AIConfidentScore,
		&ref.ReflectionScore,
		&ref.WeightedReflectionScore,
		&ref.CreatedAt,
		&ref.TreeNodeID,
	)

	if primaryEmotionStr.Valid {
		ref.PrimaryEmotion = &primaryEmotionStr.String
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("repo.GetReflectionByID scan failed: %w", err)
	}

	return &ref, nil
}

func (r *repositoryImpl) GetAllReflections(ctx context.Context, filter model.GetReflectionsFilter) ([]model.Reflection, error) {
	query := `SELECT 
		CONVERT(VARCHAR(36), r.reflect_id) as reflect_id, 
		r.reflect_score, 
		r.reflect_description, 
		r.reflect, 
		r.progress_score, 
		r.challenge_score, 
		r.summary,
		r.sentiment_analysis,
		r.primary_emotion,
		r.struggle_point,
		r.ai_confident_score,
		r.reflection_score,
		r.weighted_reflection_score,
		r.create_at, 
		CONVERT(VARCHAR(36), r.tree_node_id) as tree_node_id 
		FROM Reflect r`

	// Build WHERE clause dynamically
	whereClauses := []string{}
	args := []interface{}{}
	paramCount := 1

	// Join tables only when necessary based on filters
	needTreeNode := filter.TreeID != nil || filter.AlbumID != nil || filter.UserID != nil
	needTree := filter.AlbumID != nil || filter.UserID != nil
	needAlbum := filter.UserID != nil

	if needTreeNode {
		query += ` 
		INNER JOIN tree_node tn ON r.tree_node_id = tn.tree_node_id`
	}

	if needTree {
		query += `
		INNER JOIN tree t ON tn.tree_id = t.tree_id`
	}

	if needAlbum {
		query += `
		INNER JOIN Album a ON t.album_id = a.album_id`
	}

	// Add filter conditions
	if filter.TreeNodeID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("r.tree_node_id = @p%d", paramCount))
		args = append(args, *filter.TreeNodeID)
		paramCount++
	}

	if filter.TreeID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("tn.tree_id = @p%d", paramCount))
		args = append(args, *filter.TreeID)
		paramCount++
	}

	if filter.AlbumID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.album_id = @p%d", paramCount))
		args = append(args, *filter.AlbumID)
		paramCount++
	}

	if filter.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("a.user_id = @p%d", paramCount))
		args = append(args, *filter.UserID)
		paramCount++
	}

	// Keyset (seek) pagination cursor on (create_at, reflect_id).
	if (filter.BeforeCreatedAt == nil) != (filter.BeforeReflectID == nil) {
		return nil, fmt.Errorf("both before_created_at and before_reflect_id are required for keyset pagination")
	}
	if filter.BeforeCreatedAt != nil && filter.BeforeReflectID != nil {
		whereClauses = append(whereClauses,
			fmt.Sprintf("(r.create_at < @p%d OR (r.create_at = @p%d AND r.reflect_id < CAST(@p%d AS UNIQUEIDENTIFIER)))", paramCount, paramCount+1, paramCount+2),
		)
		args = append(args, *filter.BeforeCreatedAt, *filter.BeforeCreatedAt, *filter.BeforeReflectID)
		paramCount += 3
	}

	// Add WHERE clause if any filters exist
	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}

	// Add ORDER BY with tie-breaker for stable keyset pagination.
	query += ` ORDER BY r.create_at DESC, r.reflect_id DESC`

	// Add page size; seek cursor is handled by WHERE clause above.
	if filter.Limit > 0 {
		query += fmt.Sprintf(" OFFSET 0 ROWS FETCH NEXT @p%d ROWS ONLY", paramCount)
		args = append(args, filter.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repo.GetAllReflections query failed: %w", err)
	}
	defer rows.Close()

	var reflections []model.Reflection
	for rows.Next() {
		var ref model.Reflection
		var primaryEmotionStr sql.NullString
		if err := rows.Scan(
			&ref.ReflectID,
			&ref.ReflectScore,
			&ref.ReflectDescription,
			&ref.Reflect,
			&ref.ProgressScore,
			&ref.ChallengeScore,
			&ref.Summary,
			&ref.SentimentAnalysis,
			&primaryEmotionStr,
			&ref.StrugglePoint,
			&ref.AIConfidentScore,
			&ref.ReflectionScore,
			&ref.WeightedReflectionScore,
			&ref.CreatedAt,
			&ref.TreeNodeID,
		); err != nil {
			return nil, fmt.Errorf("repo.GetAllReflections scan failed: %w", err)
		}
		if primaryEmotionStr.Valid {
			ref.PrimaryEmotion = &primaryEmotionStr.String
		}
		reflections = append(reflections, ref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetAllReflections row iteration failed: %w", err)
	}

	return reflections, nil
}

func (r *repositoryImpl) UpdateReflection(ctx context.Context, reflectID string, req model.UpdateReflectionRequest) error {
	query := `UPDATE Reflect
		SET
			reflect_score = @p1,
			reflect_description = @p2,
			reflect = @p3,
			progress_score = @p4,
			challenge_score = @p5
		WHERE reflect_id = @p6`

	res, err := r.db.ExecContext(ctx, query,
		req.FeelScore,
		req.LearningReflect,
		req.MoodReflect,
		req.ProgressScore,
		req.ChallengeScore,
		reflectID,
	)

	if err != nil {
		return fmt.Errorf("repo.UpdateReflection exec failed: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *repositoryImpl) DeleteReflection(ctx context.Context, reflectID string) error {
	query := `DELETE FROM Reflect WHERE reflect_id = @p1`

	res, err := r.db.ExecContext(ctx, query, reflectID)
	if err != nil {
		return fmt.Errorf("repo.DeleteReflection exec failed: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
