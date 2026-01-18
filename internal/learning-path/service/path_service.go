package service

import (
	"context"
	"database/sql"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
	"regexp"
	"strings"
	"fmt"
)

func (s *serviceImpl) GetPaths(ctx context.Context) ([]model.LearningPath, error) {
	paths, err := s.pathRepo.GetAllLearnningPath(ctx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return paths, nil
}

func (s *serviceImpl) GetPathDetails(ctx context.Context, id string) (*model.LearningPath, error) {
	if id == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}
	path, err := s.pathRepo.GetLearnningPathByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("learning path with id '%s' not found", id)
		}
		return nil, apperror.NewInternal(err)
	}
	return path, nil
}

func (s *serviceImpl) CreatePath(ctx context.Context, req model.CreatePathRequest) (string, error) {
	if req.Title == "" {
		return "", apperror.NewBadRequest("title cannot be empty")
	}
	id, err := s.pathRepo.CreateLearnningPath(ctx, req)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return "", apperror.NewConflict("learning path with this title or ID already exists")
		}
		if apperror.IsForeignKeyError(err) {
			return "", apperror.NewBadRequest("invalid creator_id: user does not exist")
		}
		return "", apperror.NewInternal(err)
	}
	return id, nil
}

func (s *serviceImpl) UpdatePath(ctx context.Context, id string, req model.UpdatePathRequest) error {
	if id == "" {
		return apperror.NewBadRequest("user_id is required")
	}
	if req.Title == "" &&
		req.Objective == "" &&
		req.Description == "" &&
		req.CoverImgURL == "" &&
		req.Status == "" {
		return apperror.NewBadRequest("request body cannot be empty")
	}
	if _, err := s.pathRepo.GetLearnningPathByID(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("cannot update: path id '%s' not found", id)
		}
		return apperror.NewInternal(err)
	}

	if err := s.pathRepo.UpdateLearnningPath(ctx, id, req); err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return apperror.NewConflict("learning path with this title already exists")
		}
		if apperror.IsForeignKeyError(err) {
			return apperror.NewBadRequest("invalid creator_id: user does not exist")
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) DeletePath(ctx context.Context, id string) error {
	if id == "" {
		return apperror.NewBadRequest("path_id is required")
	}
	if err := s.pathRepo.DeleteLearnningPath(ctx, id); err != nil {
		if apperror.IsForeignKeyError(err) {
			return apperror.NewConflict("cannot delete path: there are existing enrollments or nodes associated with this path")
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) StartPath(ctx context.Context, pathID string, userID string) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}
	if pathID == "" {
		return apperror.NewBadRequest("path_ID is required")
	}
	if err := s.pathRepo.EnrollLearnningPathUser(ctx, pathID, userID); err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return apperror.NewConflict("user is already enrolled in this learning path")
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *serviceImpl) GetEnrollmentStatus(ctx context.Context, pathID string, userID string) (*model.PathEnroll, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}
	if pathID == "" {
		return nil, apperror.NewBadRequest("path_id is required")
	}
	enroll, err := s.pathRepo.GetLearnningPathEnrollmentStatus(ctx, pathID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("enrollment not found for user '%s'", userID)
		}
		return nil, apperror.NewInternal(err)
	}
	return enroll, nil
}

func (s *serviceImpl) GetPathProgress(ctx context.Context, pathID string, userID string) (*model.PathProgressResponse, error) {
	if pathID == "" || userID == "" {
		return nil, apperror.NewBadRequest("path_id and user_id are required")
	}

	progress, err := s.pathRepo.GetUserPathProgress(ctx, pathID, userID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	return progress, nil
}

func (s *serviceImpl) GeneratePathWithAI(ctx context.Context, topic string) (*model.GeneratedPathResponse, error) {
	if topic == "" {
		return nil, apperror.NewBadRequest("topic is required")
	}

	rawResponse, err := s.aiClient.GenerateLearningPath(ctx, topic)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	nodes := parseAINodes(rawResponse.Result)

	return &model.GeneratedPathResponse{
		Topic: rawResponse.Topic,
		Nodes: nodes,
	}, nil
}

func parseAINodes(rawResult string) []model.GeneratedNode {
	var nodes []model.GeneratedNode
	
	segments := strings.Split(rawResult, ",")
	
	re := regexp.MustCompile(`Node\s+(\d+):\s+(.+)`)

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		matches := re.FindStringSubmatch(seg)
		if len(matches) == 3 {
			seq := 0
			fmt.Sscanf(matches[1], "%d", &seq)
			nodes = append(nodes, model.GeneratedNode{
				Sequence: seq,
				Title:    matches[2],
			})
		}
	}
	return nodes
}