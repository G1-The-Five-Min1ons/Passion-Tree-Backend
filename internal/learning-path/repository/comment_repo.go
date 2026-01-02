package repository

import (
	"time"
    "github.com/google/uuid"
    "passiontree/internal/learning-path/model"
)

func (r *repository) CreateComment(req model.CreateCommentRequest) (string, error) {
	id := uuid.New().String()
	now := time.Now()
	query := `INSERT INTO node_comment (comment_id, parent_id, content, create_at, node_id) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, id, req.ParentID, req.Content, now, req.NodeID)
	return id, err
}

func (r *repository) GetCommentsByNodeID(nodeID string) ([]model.NodeComment, error) {
	query := `SELECT comment_id, parent_id, content, create_at, edit_at FROM node_comment WHERE node_id = ?`
	rows, err := r.db.Query(query, nodeID)
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
		reactions, _ := r.GetReactionsByCommentID(c.CommentID)
		c.Reactions = reactions

		comments = append(comments, c)
	}
	return comments, nil
}

func (r *repository) DeleteComment(commentID string) error {
	_, err := r.db.Exec(`DELETE FROM node_comment WHERE comment_id = ?`, commentID)
	return err
}

func (r *repository) CreateReaction(req model.CreateReactionRequest) error {
	id := uuid.New().String()
	query := `INSERT INTO comment_reaction (reaction_id, reaction_type, comment_id) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, id, req.ReactionType, req.CommentID)
	return err
}

func (r *repository) GetReactionsByCommentID(commentID string) ([]model.CommentReaction, error) {
	query := `SELECT reaction_id, reaction_type, comment_id FROM comment_reaction WHERE comment_id = ?`
	rows, err := r.db.Query(query, commentID)
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
	return reactions, nil
}

func (r *repository) CreateMention(req model.CreateMentionRequest) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO comment_mention (reaction_id, create_at, comment_id) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, id, time.Now(), req.CommentID)
	return id, err
}