package service

import (
	"passiontree/internal/learning-path/model"
)

func (s *service) AddComment(req model.CreateCommentRequest) (string, error) {
	return s.repo.CreateComment(req)
}

func (s *service) GetNodeComments(nodeID string) ([]model.NodeComment, error) {
	return s.repo.GetCommentsByNodeID(nodeID)
}

func (s *service) RemoveComment(commentID string) error {
	return s.repo.DeleteComment(commentID)
}

func (s *service) AddReaction(req model.CreateReactionRequest) error {
	return s.repo.CreateReaction(req)
}

func (s *service) AddMention(req model.CreateMentionRequest) (string, error) {
	return s.repo.CreateMention(req)
}