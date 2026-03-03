package service

import (
	"context"
	"strings"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

func (s *serviceImpl) AddComment(ctx context.Context, req model.CreateCommentRequest) (string, error) {
	// user_id and node_id are set from JWT token / URL param by the handler

	// If this is a reply, validate the parent exists BEFORE inserting
	// to avoid FK constraint errors from the DB
	var parentOwnerID string
	if req.ParentID != nil && *req.ParentID != "" {
		ownerID, err := s.commentRepo.GetCommentOwner(ctx, *req.ParentID)
		if err != nil {
			return "", apperror.NewNotFound("parent comment not found")
		}
		parentOwnerID = ownerID
	}

	commentID, err := s.commentRepo.CreateComment(ctx, req)
	if err != nil {
		if apperror.IsForeignKeyError(err) {
			return "", apperror.NewBadRequest("invalid node_id: node does not exist")
		}
		return "", apperror.NewInternal("failed to create comment: %v", err)
	}

	// Auto-mention the parent comment's author when this is a reply to someone else
	if parentOwnerID != "" && parentOwnerID != req.UserID {
		mentionReq := model.CreateMentionRequest{
			MentionerUserID: req.UserID,
			MentionedUserID: parentOwnerID,
			CommentID:       commentID,
		}
		// Best-effort: don't fail the reply if mention insert fails
		_, _ = s.commentRepo.CreateMention(ctx, mentionReq)
	}

	return commentID, nil
}

func (s *serviceImpl) GetNodeComments(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
	comments, err := s.commentRepo.GetCommentsByNodeID(ctx, nodeID)
	if err != nil {
		return nil, apperror.NewInternal("failed to retrieve comments: %v", err)
	}
	return comments, nil
}

func (s *serviceImpl) RemoveComment(ctx context.Context, userID, commentID string) error {
	err := s.commentRepo.DeleteComment(ctx, commentID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found or not owned by user") {
			return apperror.NewForbidden("comment not found or not owned by you")
		}
		return apperror.NewInternal("failed to delete comment: %v", err)
	}
	return nil
}

func (s *serviceImpl) UpdateComment(ctx context.Context, userID, commentID, message string) error {
	updated, err := s.commentRepo.UpdateComment(ctx, userID, commentID, message)
	if err != nil {
		return apperror.NewInternal("failed to update comment: %v", err)
	}
	if !updated {
		return apperror.NewForbidden("comment not found or not owned by you")
	}
	return nil
}

func (s *serviceImpl) ToggleReaction(ctx context.Context, req model.CreateReactionRequest) (bool, error) {
	if !model.IsValidReactionType(req.ReactionType) {
		return false, apperror.NewBadRequest("invalid reaction_type: must be one of like, love, haha, wow, sad, angry")
	}

	added, err := s.commentRepo.ToggleReaction(ctx, req)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return false, apperror.NewConflict("reaction already exists")
		}
		return false, apperror.NewInternal("failed to toggle reaction: %v", err)
	}
	return added, nil
}

func (s *serviceImpl) AddMention(ctx context.Context, req model.CreateMentionRequest) (string, error) {
	id, err := s.commentRepo.CreateMention(ctx, req)
	if err != nil {
		return "", apperror.NewInternal("failed to create mention: %v", err)
	}
	return id, nil
}
