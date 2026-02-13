package service

import (
	"context"
	"database/sql"
	"strconv"
	"regexp"
	"strings"

	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
)

const (
	ContainerLearningPath = "learning-path"
)

func (s *serviceImpl) GetPaths(ctx context.Context) ([]model.LearningPaths, error) {

	paths, err := s.pathRepo.GetAllLearningPath(ctx)
	if err != nil {
		return nil, apperror.NewInternal("failed to get all paths: %w", err)
	}

	s.logger.InfoContext(ctx, "successfully retrieved paths", "count", len(paths))
	return paths, nil
}

func (s *serviceImpl) GetPathDetails(ctx context.Context, path_id string) (*model.LearningPath, error) {
	if path_id == "" {
		return nil, apperror.NewBadRequest("path_id is required")
	}

	path, err := s.pathRepo.GetLearningPathByID(ctx, path_id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "learning path not found", "path_id", path_id)
			return nil, apperror.NewNotFound("learning path with id '%s' not found", path_id)
		}

		s.logger.ErrorContext(ctx, "database error while fetching path", "error", err, "path_id", path_id)
		return nil, apperror.NewInternal("failed to fetch path details: %w", err)
	}

	s.logger.InfoContext(ctx, "learning path details retrieved successfully", "path_id", path_id)
	return path, nil
}

func (s *serviceImpl) CreatePath(ctx context.Context, req model.CreatePathRequest) (string, error) {
	if req.Title == "" {
		return "", apperror.NewBadRequest("title cannot be empty")
	}

	if req.CoverImgURL == "" {
		return "", apperror.NewBadRequest("cover_image_url is required")
	}

	// Validate Image Source & Existence
	// if !strings.Contains(req.CoverImgURL, ContainerLearningPath) {
	// 	return "", apperror.NewBadRequest("invalid image URL source")
	// }
	// if err := s.storage.ValidateUploadedFile(ctx, req.CoverImgURL, ContainerLearningPath); err != nil {
	// 	return "", apperror.NewBadRequest("image validation failed: %v", err)
	// }

	id, err := s.pathRepo.CreateLearningPath(ctx, req)
	if err != nil {
		if apperror.IsDuplicateKeyError(err) {
			s.logger.WarnContext(ctx, "conflict: path title already exists", "title", req.Title)
			return "", apperror.NewConflict("learning path with this title or ID already exists")
		}
		if apperror.IsForeignKeyError(err) {
			s.logger.WarnContext(ctx, "invalid creator: user not found", "creator_id", req.CreatorID)
			return "", apperror.NewBadRequest("invalid creator_id: user does not exist")
		}

		s.logger.ErrorContext(ctx, "database error during path creation", "error", err, "title", req.Title)
		return "", apperror.NewInternal("database error during path creation: %w", err)
	}

	s.logger.InfoContext(ctx, "learning path created successfully", "path_id", id)
	return id, nil
}

func (s *serviceImpl) UpdatePath(ctx context.Context, path_id string, req model.UpdatePathRequest) error {
	if path_id == "" {
		return apperror.NewBadRequest("path_id is required")
	}

	if req.Title == "" &&
		req.Objective == "" &&
		req.Description == "" &&
		req.CoverImgURL == "" &&
		req.Publish_status == "" {
		return apperror.NewBadRequest("request body cannot be empty")
	}

	if _, err := s.pathRepo.GetLearningPathByID(ctx, path_id); err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "update failed: path not found", "path_id", path_id)
			return apperror.NewNotFound("cannot update: path_id '%s' not found", path_id)
		}
		s.logger.ErrorContext(ctx, "database error during path verification", "error", err, "title", req.Title)
		return apperror.NewInternal("failed to verify learning path: %w", err)
	}

	if err := s.pathRepo.UpdateLearningPath(ctx, path_id, req); err != nil {
		if apperror.IsDuplicateKeyError(err) {
			return apperror.NewConflict("learning path with this title already exists")
		}

		if apperror.IsForeignKeyError(err) {
			return apperror.NewBadRequest("invalid creator_id: user does not exist")
		}

		s.logger.ErrorContext(ctx, "failed to update learning path", "error", err, "path_id", path_id)
		return apperror.NewInternal("failed to update learning path: %w", err)
	}

	s.logger.InfoContext(ctx, "learning path updated successfully", "path_id", path_id)
	return nil
}

func (s *serviceImpl) DeletePath(ctx context.Context, path_id string) error {
	if path_id == "" {
		return apperror.NewBadRequest("path_id is required")
	}

	s.logger.InfoContext(ctx, "requesting path deletion", "path_id", path_id)

	if err := s.pathRepo.DeleteLearningPath(ctx, path_id); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("learning path not found")
		}
		if apperror.IsForeignKeyError(err) {
			s.logger.WarnContext(ctx, "deletion blocked by dependencies", "path_id", path_id)
			return apperror.NewConflict("cannot delete path: there are existing enrollments or nodes associated with this path")
		}

		s.logger.ErrorContext(ctx, "database error during path deletion", "error", err, "path_id", path_id)
		return apperror.NewInternal("database error during path deletion: %w", err)
	}

	s.logger.InfoContext(ctx, "learning path deleted successfully", "path_id", path_id)
	return nil
}

func (s *serviceImpl) StartPath(ctx context.Context, path_id string, user_id string) error {
	if user_id == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	if path_id == "" {
		return apperror.NewBadRequest("path_id is required")
	}

	s.logger.InfoContext(ctx, "enrolling user in path", "user_id", user_id, "path_id", path_id)

	if err := s.pathRepo.EnrollLearningPathUser(ctx, path_id, user_id); err != nil {
		if apperror.IsDuplicateKeyError(err) {
			s.logger.WarnContext(ctx, "user already enrolled", "user_id", user_id, "path_id", path_id)
			return apperror.NewConflict("user is already enrolled in this learning path")
		}
		return apperror.NewInternal("failed to enroll user '%s' in path '%s': %w", user_id, path_id, err)
	}

	s.logger.InfoContext(ctx, "user enrollment successful", "user_id", user_id, "path_id", path_id)
	return nil
}

func (s *serviceImpl) GetEnrollmentStatus(ctx context.Context, path_id string, user_id string) (*model.PathEnroll, error) {
	if user_id == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	if path_id == "" {
		return nil, apperror.NewBadRequest("path_id is required")
	}

	s.logger.InfoContext(ctx, "checking enrollment status", "user_id", user_id, "path_id", path_id)

	enroll, err := s.pathRepo.GetLearningPathEnrollmentStatus(ctx, path_id, user_id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.InfoContext(ctx, "no enrollment found", "user_id", user_id, "path_id", path_id)
			return nil, apperror.NewNotFound("enrollment not found for user '%s'", user_id)
		}

		s.logger.ErrorContext(ctx, "failed to fetch enrollment status", "error", err, "user_id", user_id, "path_id", path_id)
		return nil, apperror.NewInternal("failed to get enrollment status for user '%s' in path '%s': %w", user_id, path_id, err)
	}
	return enroll, nil
}

func (s *serviceImpl) GetPathProgress(ctx context.Context, path_id string, user_id string) (*model.PathProgressResponse, error) {
	if path_id == "" || user_id == "" {
		return nil, apperror.NewBadRequest("path_id and user_id are required")
	}

	s.logger.InfoContext(ctx, "calculating path progress", "user_id", user_id, "path_id", path_id)

	progress, err := s.pathRepo.GetUserPathProgress(ctx, path_id, user_id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to calculate progress", "error", err, "user_id", user_id, "path_id", path_id)
		return nil, apperror.NewInternal("failed to calculate progress for user '%s' in path '%s': %w", user_id, path_id, err)
	}

	return progress, nil
}

func (s *serviceImpl) GeneratePathWithAI(ctx context.Context, topic string) (*model.GeneratedPathResponse, error) {
	if topic == "" {
		return nil, apperror.NewBadRequest("topic is required")
	}

	s.logger.InfoContext(ctx, "generating learning path with AI", "topic", topic)

	rawResponse, err := s.aiClient.GenerateLearningPath(ctx, topic)
	if err != nil {
		s.logger.ErrorContext(ctx, "AI generation failed", "error", err, "topic", topic)
		return nil, apperror.NewInternal("AI learning path generation failed for topic '%s': %w", topic, err)
	}

	nodes := parseAINodes(rawResponse.Result)

	s.logger.InfoContext(ctx, "AI path generation successful", "topic", topic, "nodes_generated", len(nodes))

	return &model.GeneratedPathResponse{
		Topic: rawResponse.Topic,
		Nodes: nodes,
	}, nil
}

func parseAINodes(rawResult string) []model.GeneratedNode {
    var nodes []model.GeneratedNode
    segments := strings.Split(rawResult, ",")
    re := regexp.MustCompile(`Node\s+(\d+):\s+(.+)`)

    // ลำดับสำรอง (Fallback) เริ่มที่ 1
    fallbackSeq := 1

    for _, seg := range segments {
        seg = strings.TrimSpace(seg)
        matches := re.FindStringSubmatch(seg)
        
        if len(matches) == 3 {
            aiSeq, err := strconv.Atoi(matches[1])
            
            finalSeq := 0
            if err == nil {
                finalSeq = aiSeq
                fallbackSeq = aiSeq + 1
            } else {
                finalSeq = fallbackSeq
                fallbackSeq++
            }

            nodes = append(nodes, model.GeneratedNode{
                Sequence: finalSeq,
                Title:    matches[2],
            })
        }
    }
    return nodes
}

func (s *serviceImpl) UpdatePathCoverImage(ctx context.Context, pathID string, coverImgURL string) error {
    if pathID == "" {
        return apperror.NewBadRequest("path_id is required")
    }
    if coverImgURL == "" {
        return apperror.NewBadRequest("cover_image_url is required")
    }

    s.logger.InfoContext(ctx, "updating learning path cover image", "path_id", pathID)

    if !strings.Contains(coverImgURL, "learning-path") {
        return apperror.NewBadRequest("Invalid image URL source")
    }

    err := s.storage.ValidateUploadedFile(ctx, coverImgURL, "learning-path")
    if err != nil {
        s.logger.WarnContext(ctx, "image validation failed", "error", err, "url", coverImgURL)
        return apperror.NewBadRequest("Image validation failed: %v", err)
    }

    if _, err := s.pathRepo.GetLearningPathByID(ctx, pathID); err != nil {
        if err == sql.ErrNoRows {
            return apperror.NewNotFound("learning path not found")
        }
        return apperror.NewInternal("failed to verify learning path: %w", err)
    }

    if err := s.pathRepo.UpdateLearningPathImage(ctx, pathID, coverImgURL); err != nil {
        if err == sql.ErrNoRows { 
            return apperror.NewNotFound("learning path not found")
        }
        return apperror.NewInternal("failed to update cover image for path %s: %w", pathID, err)
    }

    s.logger.InfoContext(ctx, "learning path cover image updated successfully", "path_id", pathID)
    return nil
}
