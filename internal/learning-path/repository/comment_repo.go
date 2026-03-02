package repository

import (
	"context"
	"fmt"
	"passiontree/internal/learning-path/model"
	"time"

	"github.com/google/uuid"
)

func (r *repositoryImpl) CreateComment(ctx context.Context, req model.CreateCommentRequest) (string, error) {
	id := uuid.New().String()
	now := time.Now()

	// Treat empty string parent_id as NULL to avoid UNIQUEIDENTIFIER conversion failure
	var parentID interface{}
	if req.ParentID != nil && *req.ParentID != "" {
		parentID = *req.ParentID
	}

	query := `INSERT INTO Node_Comment (comment_id, user_id, node_id, message, parent_id, create_at) VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`
	_, err := r.db.ExecContext(ctx, query, id, req.UserID, req.NodeID, req.Message, parentID, now)
	return id, err
}

func (r *repositoryImpl) GetCommentsByNodeID(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
	query := `SELECT user_id, comment_id, message, create_at, edit_at, node_id, parent_id FROM Node_Comment WHERE node_id = @p1`
	rows, err := r.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.NodeComment
	for rows.Next() {
		var c model.NodeComment
		if err := rows.Scan(&c.UserID, &c.CommentID, &c.Message, &c.CreatedAt, &c.EditAt, &c.NodeID, &c.ParentID); err != nil {
			continue
		}
		reactions, _ := r.GetReactionsByCommentID(ctx, c.CommentID)
		c.Reactions = reactions
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetCommentsByNodeID row iteration failed: %w", err)
	}

	return comments, nil
}

func (r *repositoryImpl) DeleteComment(ctx context.Context, commentID, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo.DeleteComment: begin tx failed: %w", err)
	}
	defer tx.Rollback()

	// 1. Delete mentions on all replies
	_, err = tx.ExecContext(ctx, `
		DELETE FROM Comment_Mention 
		WHERE comment_id IN (
			SELECT comment_id FROM Node_Comment WHERE parent_id = @p1
		)`, commentID)
	if err != nil {
		return fmt.Errorf("repo.DeleteComment: delete reply mentions failed: %w", err)
	}

	// 2. Delete reactions on all replies
	_, err = tx.ExecContext(ctx, `
		DELETE FROM comment_reaction 
		WHERE comment_id IN (
			SELECT comment_id FROM Node_Comment WHERE parent_id = @p1
		)`, commentID)
	if err != nil {
		return fmt.Errorf("repo.DeleteComment: delete reply reactions failed: %w", err)
	}

	// 3. Delete all replies
	_, err = tx.ExecContext(ctx, `DELETE FROM Node_Comment WHERE parent_id = @p1`, commentID)
	if err != nil {
		return fmt.Errorf("repo.DeleteComment: delete replies failed: %w", err)
	}

	// 4. Delete mentions on the parent comment
	_, err = tx.ExecContext(ctx, `DELETE FROM Comment_Mention WHERE comment_id = @p1`, commentID)
	if err != nil {
		return fmt.Errorf("repo.DeleteComment: delete parent mentions failed: %w", err)
	}

	// 5. Delete reactions on the parent comment
	_, err = tx.ExecContext(ctx, `DELETE FROM comment_reaction WHERE comment_id = @p1`, commentID)
	if err != nil {
		return fmt.Errorf("repo.DeleteComment: delete parent reactions failed: %w", err)
	}

	// 6. Delete the parent comment (ownership enforced here)
	res, err := tx.ExecContext(ctx, `DELETE FROM Node_Comment WHERE comment_id = @p1 AND user_id = @p2`, commentID, userID)
	if err != nil {
		return fmt.Errorf("repo.DeleteComment: delete parent failed: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("repo.DeleteComment: comment not found or not owned by user")
	}

	return tx.Commit()
}

func (r *repositoryImpl) CreateReaction(ctx context.Context, req model.CreateReactionRequest) error {
	id := uuid.New().String()
	query := `INSERT INTO comment_reaction (reaction_id, reaction_type, comment_id) VALUES (@p1, @p2, @p3)`
	_, err := r.db.ExecContext(ctx, query, id, req.ReactionType, req.CommentID)
	return err
}

func (r *repositoryImpl) GetReactionsByCommentID(ctx context.Context, commentID string) ([]model.CommentReaction, error) {
	query := `SELECT reaction_id, reaction_type, comment_id FROM comment_reaction WHERE comment_id = @p1`
	rows, err := r.db.QueryContext(ctx, query, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []model.CommentReaction
	for rows.Next() {
		var rc model.CommentReaction
		if err := rows.Scan(&rc.ReactionID, &rc.ReactionType, &rc.CommentID); err != nil {
			continue
		}
		reactions = append(reactions, rc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetReactionsByCommentID row iteration failed: %w", err)
	}

	return reactions, nil
}

func (r *repositoryImpl) CreateMention(ctx context.Context, req model.CreateMentionRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO Comment_Mention (reaction_id, create_at, comment_id, mentioned_user_id) VALUES (@p1, @p2, @p3, @p4)`
	_, err := r.db.ExecContext(ctx, query, id, time.Now(), req.CommentID, req.MentionedUserID)
	return id, err
}

func (r *repositoryImpl) UpdateComment(ctx context.Context, userID, messageID, message string) (bool, error) {
	query := `UPDATE Node_Comment SET message = @p1, edit_at = @p2 WHERE comment_id = @p3 AND user_id = @p4`
	res, err := r.db.ExecContext(ctx, query, message, time.Now(), messageID, userID)
	if err != nil {
		return false, err
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

// GetCommentOwner returns the user_id of the original comment author.
// Used to auto-create a mention when someone replies to a comment.
func (r *repositoryImpl) GetCommentOwner(ctx context.Context, commentID string) (string, error) {
	var userID string
	query := `SELECT CONVERT(VARCHAR(36), user_id) FROM Node_Comment WHERE comment_id = @p1`
	err := r.db.QueryRowContext(ctx, query, commentID).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("repo.GetCommentOwner: comment not found or query failed: %w", err)
	}
	return userID, nil
}
