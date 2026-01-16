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
	query := `INSERT INTO node_comment (comment_id, parent_id, content, create_at, node_id) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, id, req.ParentID, req.Content, now, req.NodeID)
	return id, err
}

func (r *repositoryImpl) GetCommentsByNodeID(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
	query := `SELECT comment_id, parent_id, content, create_at, edit_at FROM node_comment WHERE node_id = ?`
	rows, err := r.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.NodeComment
	for rows.Next() {
		var c model.NodeComment
		if err := rows.Scan(&c.CommentID, &c.ParentID, &c.Content, &c.CreatedAt, &c.EditAt); err != nil {
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

func (r *repositoryImpl) DeleteComment(ctx context.Context, commentID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM node_comment WHERE comment_id = ?`, commentID)
	return err
}

func (r *repositoryImpl) CreateReaction(ctx context.Context, req model.CreateReactionRequest) error {
	id := uuid.New().String()
	query := `INSERT INTO comment_reaction (reaction_id, reaction_type, comment_id) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, id, req.ReactionType, req.CommentID)
	return err
}

func (r *repositoryImpl) GetReactionsByCommentID(ctx context.Context, commentID string) ([]model.CommentReaction, error) {
	query := `SELECT reaction_id, reaction_type, comment_id FROM comment_reaction WHERE comment_id = ?`
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
	query := `INSERT INTO comment_mention (reaction_id, create_at, comment_id) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, id, time.Now(), req.CommentID)
	return id, err
}
