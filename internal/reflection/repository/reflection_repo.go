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

	query := `INSERT INTO Reflect
		(reflect_id, reflect_score, reflect_description, reflect, progress_score, challenge_score, 
		summary, sentiment_analysis, primary_emotion, struggle_point, ai_confident_score, reflection_score, weighted_reflection_score,
		create_at, tree_node_id) 
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11, @p12, @p13, GETDATE(), @p14)`

	_, err := r.db.ExecContext(ctx, query,
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

	// Add WHERE clause if any filters exist
	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}

	// Add ORDER BY
	query += ` ORDER BY r.create_at DESC`

	// Add pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" OFFSET @p%d ROWS FETCH NEXT @p%d ROWS ONLY", paramCount, paramCount+1)
		args = append(args, filter.Offset, filter.Limit)
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
