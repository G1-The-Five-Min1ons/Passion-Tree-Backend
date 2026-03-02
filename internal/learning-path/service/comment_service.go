package service

import (
	"context"
	"fmt"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/learning-path/repository"
)

func (s *commentServiceImpl) AddComment(ctx context.Context, req model.CreateCommentRequest) (string, error) {
	// user_id and node_id are set from JWT token / URL param by the handler

	// If this is a reply, validate the parent exists BEFORE inserting
	// to avoid FK constraint errors from the DB
	var parentOwnerID string
	if req.ParentID != nil && *req.ParentID != "" {
		ownerID, err := s.repo.GetCommentOwner(ctx, *req.ParentID)
		if err != nil {
			return "", fmt.Errorf("parent comment not found: %w", err)
		}
		parentOwnerID = ownerID
	}

	commentID, err := s.repo.CreateComment(ctx, req)
	if err != nil {
		return "", err
	}

	// Auto-mention the parent comment's author when this is a reply to someone else
	if parentOwnerID != "" && parentOwnerID != req.UserID {
		mentionReq := model.CreateMentionRequest{
			MentionerUserID: req.UserID,
			MentionedUserID: parentOwnerID,
			CommentID:       commentID,
		}
		// Best-effort: don't fail the reply if mention insert fails
		_, _ = s.repo.CreateMention(ctx, mentionReq)
	}

	return commentID, nil
}

func (s *commentServiceImpl) GetNodeComments(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
	return s.repo.GetCommentsByNodeID(ctx, nodeID)
}

func (s *commentServiceImpl) RemoveComment(ctx context.Context, userID, commentID string) error {
	return s.repo.DeleteComment(ctx, commentID, userID)
}

func (s *commentServiceImpl) UpdateComment(ctx context.Context, userID, commentID, message string) (bool, error) {
	// user_id is sourced from JWT token by the handler
	return s.repo.UpdateComment(ctx, userID, commentID, message)
}

func (s *commentServiceImpl) AddReaction(ctx context.Context, req model.CreateReactionRequest) error {
	return s.repo.CreateReaction(ctx, req)
}

func (s *commentServiceImpl) AddMention(ctx context.Context, req model.CreateMentionRequest) (string, error) {
	return s.repo.CreateMention(ctx, req)
}

type commentServiceImpl struct {
	repo repository.RepositoryComment
}

func NewCommentService(repo repository.RepositoryComment) ServiceComment {
	return &commentServiceImpl{repo: repo}
}
