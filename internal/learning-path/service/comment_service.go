package service

import (
	"passiontree/internal/learning-path/model"
)

func (s *service) AddComment(req model.CreateCommentRequest) (string, error) {
	return s.commentRepo.CreateComment(req)
}

func (s *service) GetNodeComments(nodeID string) ([]model.NodeComment, error) {
	return s.commentRepo.GetCommentsByNodeID(nodeID)
}

func (s *service) RemoveComment(commentID string) error {
	return s.commentRepo.DeleteComment(commentID)
}

func (s *service) AddReaction(req model.CreateReactionRequest) error {
	return s.commentRepo.CreateReaction(req)
}

func (s *service) AddMention(req model.CreateMentionRequest) (string, error) {
	return s.commentRepo.CreateMention(req)
}