package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"passiontree/internal/learning-path/model"

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
	query := `
		SELECT 
			CONVERT(VARCHAR(36), c.user_id), 
			CONVERT(VARCHAR(36), c.comment_id), 
			c.message, 
			c.create_at, 
			c.edit_at, 
			CONVERT(VARCHAR(36), c.node_id), 
			CONVERT(VARCHAR(36), c.parent_id),
			COALESCE(u.first_name + ' ' + u.last_name, 'Unknown User')
		FROM Node_Comment c
		LEFT JOIN users u ON c.user_id = u.user_id
		WHERE c.node_id = @p1
		ORDER BY c.create_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.NodeComment
	var commentIDs []string
	for rows.Next() {
		var c model.NodeComment
		if err := rows.Scan(&c.UserID, &c.CommentID, &c.Message, &c.CreatedAt, &c.EditAt, &c.NodeID, &c.ParentID, &c.UserName); err != nil {
			return nil, fmt.Errorf("repo.GetCommentsByNodeID scan failed: %w", err)
		}
		comments = append(comments, c)
		commentIDs = append(commentIDs, c.CommentID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.GetCommentsByNodeID row iteration failed: %w", err)
	}

	if len(commentIDs) == 0 {
		return comments, nil
	}

	// Batch fetch all reactions in a single query — avoids N+1
	reactionsMap, err := r.batchGetReactionsByCommentIDs(ctx, commentIDs)
	if err != nil {
		return nil, err
	}
	for i := range comments {
		if reacts, ok := reactionsMap[comments[i].CommentID]; ok {
			comments[i].Reactions = reacts
		}
	}

	return comments, nil
}

// batchGetReactionsByCommentIDs fetches reactions for multiple comment IDs in one query.
func (r *repositoryImpl) batchGetReactionsByCommentIDs(ctx context.Context, commentIDs []string) (map[string][]model.CommentReaction, error) {
	// Build parameterized IN clause: @p1, @p2, ...
	placeholders := make([]string, len(commentIDs))
	args := make([]interface{}, len(commentIDs))
	for i, id := range commentIDs {
		placeholders[i] = fmt.Sprintf("@p%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT CONVERT(VARCHAR(36), reaction_id), reaction_type, CONVERT(VARCHAR(36), comment_id), CONVERT(VARCHAR(36), user_id) FROM comment_reaction WHERE comment_id IN (%s)`,
		strings.Join(placeholders, ","),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repo.batchGetReactionsByCommentIDs: query failed: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]model.CommentReaction)
	for rows.Next() {
		var rc model.CommentReaction
		var nullableUserID sql.NullString
		if err := rows.Scan(&rc.ReactionID, &rc.ReactionType, &rc.CommentID, &nullableUserID); err != nil {
			return nil, fmt.Errorf("repo.batchGetReactionsByCommentIDs scan failed: %w", err)
		}
		if nullableUserID.Valid {
			rc.UserID = nullableUserID.String
		}
		result[rc.CommentID] = append(result[rc.CommentID], rc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo.batchGetReactionsByCommentIDs row iteration failed: %w", err)
	}

	return result, nil
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

func (r *repositoryImpl) ToggleReaction(ctx context.Context, req model.CreateReactionRequest) (bool, error) {
	// First check if user already reacted
	var existingReactionID string
	err := r.db.QueryRowContext(ctx, `SELECT CONVERT(VARCHAR(36), reaction_id) FROM comment_reaction WHERE comment_id = @p1 AND user_id = @p2`, req.CommentID, req.UserID).Scan(&existingReactionID)

	if err == sql.ErrNoRows {
		// Doesn't exist, insert
		id := uuid.New().String()
		query := `INSERT INTO comment_reaction (reaction_id, reaction_type, comment_id, user_id) VALUES (@p1, @p2, @p3, @p4)`
		_, err := r.db.ExecContext(ctx, query, id, req.ReactionType, req.CommentID, req.UserID)
		return true, err // true means added
	} else if err != nil {
		return false, err
	}

	// Exists, so delete it
	query := `DELETE FROM comment_reaction WHERE reaction_id = @p1`
	_, err = r.db.ExecContext(ctx, query, existingReactionID)
	return false, err // false means removed
}

func (r *repositoryImpl) GetReactionsByCommentID(ctx context.Context, commentID string) ([]model.CommentReaction, error) {
	query := `SELECT CONVERT(VARCHAR(36), reaction_id), reaction_type, CONVERT(VARCHAR(36), comment_id), CONVERT(VARCHAR(36), user_id) FROM comment_reaction WHERE comment_id = @p1`
	rows, err := r.db.QueryContext(ctx, query, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []model.CommentReaction
	for rows.Next() {
		var rc model.CommentReaction
		var nullableUserID sql.NullString
		if err := rows.Scan(&rc.ReactionID, &rc.ReactionType, &rc.CommentID, &nullableUserID); err != nil {
			return nil, fmt.Errorf("repo.GetReactionsByCommentID scan failed: %w", err)
		}
		if nullableUserID.Valid {
			rc.UserID = nullableUserID.String
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
