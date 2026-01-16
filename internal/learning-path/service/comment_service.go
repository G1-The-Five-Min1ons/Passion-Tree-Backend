package service

import (
	"context"
	"passiontree/internal/learning-path/model"
)

func (s *serviceImpl) AddComment(ctx context.Context, req model.CreateCommentRequest) (string, error) {
	return s.commentRepo.CreateComment(ctx, req)
}

func (s *serviceImpl) GetNodeComments(ctx context.Context, nodeID string) ([]model.NodeComment, error) {
	return s.commentRepo.GetCommentsByNodeID(ctx, nodeID)
}

func (s *serviceImpl) RemoveComment(ctx context.Context, commentID string) error {
	return s.commentRepo.DeleteComment(ctx, commentID)
}

func (s *serviceImpl) AddReaction(ctx context.Context, req model.CreateReactionRequest) error {
	return s.commentRepo.CreateReaction(ctx, req)
}

func (s *serviceImpl) AddMention(ctx context.Context, req model.CreateMentionRequest) (string, error) {
	return s.commentRepo.CreateMention(ctx, req)
}
