package service

import (
	"context"
	"database/sql"
	"regexp"
	"strconv"
	"strings"
	"time"

	"passiontree/internal/config"
	"passiontree/internal/learning-path/model"
	missionModel "passiontree/internal/mission/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/utils"
)

func (s *serviceImpl) GetPaths(ctx context.Context) ([]model.LearningPath, error) {

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
	if req.CreatorID == "" {
		return "", apperror.NewBadRequest("creator_id is required")
	}

	creatorRole, isTeacherVerified, err := s.pathRepo.GetPathCreatorVerification(ctx, req.CreatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", apperror.NewBadRequest("invalid creator_id: user does not exist")
		}
		return "", apperror.NewInternal("failed to validate creator verification status: %w", err)
	}

	if creatorRole == "teacher" && !isTeacherVerified {
		return "", apperror.NewForbidden("teacher account must be approved before creating learning paths (bind phone number, submit application, and wait for admin approval)")
	}

	if req.Title == "" {
		return "", apperror.NewBadRequest("title cannot be empty")
	}

	if req.CoverImgURL == "" {
		return "", apperror.NewBadRequest("cover_image_url is required")
	}

	cfg, err := config.LoadDBConfig()
	// Validate Image Source & Existence
	if cfg != nil && s.storage != nil {
		if !strings.Contains(req.CoverImgURL, cfg.ContainerLearningPath) {
			return "", apperror.NewBadRequest("invalid image URL source")
		}
		if err := s.storage.ValidateUploadedFile(ctx, req.CoverImgURL, cfg.ContainerLearningPath); err != nil {
			return "", apperror.NewBadRequest("image validation failed: %v", err)
		}
	}

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

	existingPath, err := s.pathRepo.GetLearningPathByID(ctx, path_id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "update failed: path not found", "path_id", path_id)
			return apperror.NewNotFound("cannot update: path_id '%s' not found", path_id)
		}
		s.logger.ErrorContext(ctx, "database error during path verification", "error", err, "title", req.Title)
		return apperror.NewInternal("failed to verify learning path: %w", err)
	}

	if req.Title == "" {
		req.Title = existingPath.Title
	}
	if req.Objective == "" {
		req.Objective = existingPath.Objective
	}
	if req.Description == "" {
		req.Description = existingPath.Description
	}
	if req.CoverImgURL == "" {
		req.CoverImgURL = existingPath.CoverImgURL
	}
	if req.Publish_status == "" {
		req.Publish_status = existingPath.Publish_status
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
	if req.Publish_status == "published" {
		_, err = s.SyncLearningPath(ctx, path_id)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to sync learning path after creation", "error", err, "path_id", path_id)
			return apperror.NewInternal("failed to sync learning path: %w", err)
		}
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

	if s.aiClient != nil {
		if _, err := s.aiClient.SyncDeletePath(ctx, path_id, "learning_paths"); err != nil {
			s.logger.WarnContext(ctx, "learning path removed from SQL but Qdrant delete failed", "path_id", path_id, "error", err)
		}
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

	// Check if user is already enrolled to prevent duplicate rows in path_enroll
	existing, err := s.pathRepo.GetLearningPathEnrollmentStatus(ctx, path_id, user_id)
	if err != nil && err != sql.ErrNoRows {
		return apperror.NewInternal("failed to check enrollment status: %w", err)
	}
	if existing != nil {
		s.logger.WarnContext(ctx, "user already enrolled, skipping duplicate enrollment", "user_id", user_id, "path_id", path_id)
		return apperror.NewConflict("user is already enrolled in this learning path")
	}

	s.logger.InfoContext(ctx, "enrolling user in path", "user_id", user_id, "path_id", path_id)

	if err := s.pathRepo.EnrollLearningPathUser(ctx, path_id, user_id); err != nil {
		if apperror.IsDuplicateKeyError(err) {
			s.logger.WarnContext(ctx, "user already enrolled (DB constraint)", "user_id", user_id, "path_id", path_id)
			return apperror.NewConflict("user is already enrolled in this learning path")
		}
		return apperror.NewInternal("failed to enroll user '%s' in path '%s': %w", user_id, path_id, err)
	}

	utils.SafeGo(ctx, s.logger, "Mission_EnrollPath", 10*time.Second, func(bgCtx context.Context) error {
		return s.missionSvc.ProcessMissionEvent(bgCtx, user_id, missionModel.ConditionEnrollPath)
	})

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

	nodes := ParseAINodes(rawResponse.Result)

	s.logger.InfoContext(ctx, "AI path generation successful", "topic", topic, "nodes_generated", len(nodes))

	return &model.GeneratedPathResponse{
		Topic: rawResponse.Topic,
		Nodes: nodes,
	}, nil
}

func ParseAINodes(rawResult string) []model.GeneratedNode {
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

func (s *serviceImpl) GetUserEnrolledPaths(ctx context.Context, userID string) ([]model.EnrolledPathResponse, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	paths, err := s.pathRepo.GetUserEnrolledPaths(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get enrolled paths", "error", err)
		return nil, apperror.NewInternal("failed to retrieve enrolled paths")
	}

	return paths, nil
}

func (s *serviceImpl) UpsertRating(ctx context.Context, pathID string, userID string, req model.RatingRequest) error {
	if pathID == "" {
		return apperror.NewBadRequest("path_id is required")
	}

	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	if req.RatingContent < 1 || req.RatingContent > 5 || req.RatingInstruct < 1 || req.RatingInstruct > 5 {
		return apperror.NewBadRequest("rating scores must be between 1 and 5")
	}

	overall := float64(req.RatingContent+req.RatingInstruct) / 2.0

	rating := &model.LearningPathRating{
		RatingContent:  req.RatingContent,
		RatingInstruct: req.RatingInstruct,
		RatingOverall:  overall,
		UserID:         userID,
		PathID:         pathID,
	}

	if err := s.ratingRepo.UpsertLearningPathRating(ctx, rating); err != nil {
		s.logger.ErrorContext(ctx, "failed to upsert rating", "error", err, "path_id", pathID)
		return apperror.NewInternal("failed to submit rating: %w", err)
	}

	return nil
}

func (s *serviceImpl) GetMyRating(ctx context.Context, pathID string, userID string) (*model.LearningPathRating, error) {
	if pathID == "" {
		return nil, apperror.NewBadRequest("path_id is required")
	}

	if userID == "" {
		return nil, apperror.NewBadRequest("user_id are required")
	}

	rating, err := s.ratingRepo.GetRatingByUser(ctx, pathID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NewNotFound("rating not found for this path")
		}
		return nil, apperror.NewInternal("failed to get rating: %w", err)
	}

	return rating, nil
}

func (s *serviceImpl) DeleteRating(ctx context.Context, pathID string, userID string) error {
	if pathID == "" {
		return apperror.NewBadRequest("path_id is required")
	}

	if userID == "" {
		return apperror.NewBadRequest("user_id are required")
	}

	if err := s.ratingRepo.DeleteLearningPathRating(ctx, pathID, userID); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("rating not found")
		}
		s.logger.ErrorContext(ctx, "failed to delete rating", "error", err, "path_id", pathID)
		return apperror.NewInternal("failed to delete rating: %w", err)
	}

	return nil
}
