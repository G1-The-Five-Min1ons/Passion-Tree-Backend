package reflection

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, r Reflection) error
	Update(ctx context.Context, r Reflection) error
	Delete(ctx context.Context, r Reflection) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, ref Reflection) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO Reflect
		(reflect_id, reflect_score, reflect_description, reflect, mood, tag, create_at, tree_node_id, progress_score, challenge_score)
		VALUES(@p1,@p2,@p3,@p4,@p5,@p6,@p7,@p8,@p9,@p10)
	`,
		ref.ReflectID,
		ref.ReflectScore,
		ref.ReflectDescription,
		ref.Reflect,
		ref.Mood,
		ref.Tag,
		ref.CreatedAt,
		ref.TreeNodeID,
		ref.ProgressScore,
		ref.ChallengeScore,
	)

	return err
}

func (r *repository) Update(ctx context.Context, ref Reflection) error {

	_, err := r.db.ExecContext(ctx, `
		UPDATE Reflect
		SET
			reflect_score = @p1,
			reflect_description = @p2,
			reflect = @p3,
			mood = @p4,
			tag = @p5,
			progress_score = @p6,
			challenge_score = @p7
		WHERE reflect_id = @p8
	`,
		ref.ReflectScore,
		ref.ReflectDescription,
		ref.Reflect,
		ref.Mood,
		ref.Tag,
		ref.ProgressScore,
		ref.ChallengeScore,
		ref.ReflectID,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, ref Reflection) error {

	_, err := r.db.ExecContext(ctx, `
		DELETE FROM Reflect
		WHERE reflect_id = @p1
	`,
		ref.ReflectID,
	)

	return err
}
