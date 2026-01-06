package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"passiontree/internal/pkg/apperor"
	"passiontree/internal/reflection/model"
)

func (r *repositoryImpl) CreateReflection(ctx context.Context, req model.CreateReflectionRequest) (string, error) {
	id := uuid.New().String()

	query := `INSERT INTO Reflect
		(reflect_id, reflect_score, reflect_description, reflect, mood, tag, progress_score, challenge_score, create_at, tree_node_id) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?)`

	_, err := r.db.ExecContext(ctx, query,
		id,
		req.FeelScore,
		req.Learned,
		req.Reflect,
		"",
		"",
		req.ProgressScore,
		req.ChallengeScore,
		req.TreeNodeID,
	)

	if err != nil {
		return "", apperor.NewInternal(err)
	}

	return id, nil
}

func (r *repositoryImpl) GetReflectionByID(ctx context.Context, reflectID string) (*model.Reflection, error) {
	query := `SELECT reflect_id, reflect_score, reflect_description, reflect, mood, tag, progress_score, challenge_score, create_at, tree_node_id 
		FROM Reflect
		WHERE reflect_id = ?`

	var ref model.Reflection
	err := r.db.QueryRowContext(ctx, query, reflectID).Scan(
		&ref.ReflectID,
		&ref.ReflectScore,
		&ref.ReflectDescription,
		&ref.Reflect,
		&ref.Mood,
		&ref.Tag,
		&ref.ProgressScore,
		&ref.ChallengeScore,
		&ref.CreatedAt,
		&ref.TreeNodeID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperor.NewNotFound("reflection with id '%s' not found", reflectID)
		}
		return nil, apperor.NewInternal(err)
	}

	return &ref, nil
}

func (r *repositoryImpl) GetAllReflections(ctx context.Context) ([]model.Reflection, error) {
	query := `SELECT reflect_id, reflect_score, reflect_description, reflect, mood, tag, progress_score, challenge_score, create_at, tree_node_id 
		FROM Reflect`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperor.NewInternal(err)
	}
	defer rows.Close()

	var reflections []model.Reflection
	for rows.Next() {
		var ref model.Reflection
		if err := rows.Scan(
			&ref.ReflectID,
			&ref.ReflectScore,
			&ref.ReflectDescription,
			&ref.Reflect,
			&ref.Mood,
			&ref.Tag,
			&ref.ProgressScore,
			&ref.ChallengeScore,
			&ref.CreatedAt,
			&ref.TreeNodeID,
		); err != nil {
			return nil, fmt.Errorf("repo.GetAllReflections scan failed: %w", err)
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
			reflect_score = ?, 
			reflect_description = ?, 
			reflect = ?, 
			mood = ?, 
			tag = ?, 
			progress_score = ?, 
			challenge_score = ? 
		WHERE reflect_id = ?`

	res, err := r.db.ExecContext(ctx, query,
		req.FeelScore,
		req.Learned,
		req.Reflect,
		req.Mood,
		req.Tag,
		req.ProgressScore,
		req.ChallengeScore,
		reflectID,
	)

	if err != nil {
		return apperor.NewInternal(err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return apperor.NewNotFound("reflection with id '%s' not found", reflectID)
	}

	return nil
}

func (r *repositoryImpl) DeleteReflection(ctx context.Context, reflectID string) error {
	query := `DELETE FROM Reflect WHERE reflect_id = ?`

	res, err := r.db.ExecContext(ctx, query, reflectID)
	if err != nil {
		return apperor.NewInternal(err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return apperor.NewNotFound("reflection with id '%s' not found", reflectID)
	}

	return nil
}
