package service

import (
	"passiontree/internal/learning-path/model"
)

func (s *serviceImpl) AddComment(req model.CreateCommentRequest) (string, error) {
	return s.commentRepo.CreateComment(req)
}

func (s *serviceImpl) GetNodeComments(nodeID string) ([]model.NodeComment, error) {
	return s.commentRepo.GetCommentsByNodeID(nodeID)
}

func (s *serviceImpl) RemoveComment(commentID string) error {
	return s.commentRepo.DeleteComment(commentID)
}

func (s *serviceImpl) AddReaction(req model.CreateReactionRequest) error {
	return s.commentRepo.CreateReaction(req)
}

func (s *serviceImpl) AddMention(req model.CreateMentionRequest) (string, error) {
	return s.commentRepo.CreateMention(req)
}