package service

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
)

var phonePattern = regexp.MustCompile(`^[0-9+]{9,15}$`)

func (s *userServiceImpl) GetTeacherVerificationStatus(ctx context.Context, userID string) (*model.TeacherVerificationStatus, error) {
	if userID == "" {
		return nil, apperror.NewBadRequest("user_id is required")
	}

	status, err := s.repo.GetTeacherVerificationStatus(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get teacher verification status", "error", err, "user_id", userID)
		return nil, apperror.NewInternal("failed to get teacher verification status: %w", err)
	}
	if status == nil {
		return nil, apperror.NewNotFound("user with id '%s' not found", userID)
	}

	return status, nil
}

func (s *userServiceImpl) ApplyForTeacher(ctx context.Context, userID string, req model.ApplyTeacherRequest) error {
	if userID == "" {
		return apperror.NewBadRequest("user_id is required")
	}

	phoneNumber := strings.TrimSpace(req.PhoneNumber)
	reason := strings.TrimSpace(req.Reason)
	teachingHistory := strings.TrimSpace(req.TeachingHistory)

	if phoneNumber == "" {
		return apperror.NewBadRequest("phone_number is required")
	}
	if !phonePattern.MatchString(phoneNumber) {
		return apperror.NewBadRequest("phone_number format is invalid")
	}
	if reason == "" {
		return apperror.NewBadRequest("reason is required")
	}
	if teachingHistory == "" {
		return apperror.NewBadRequest("teaching_history is required")
	}

	user, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check user for teacher application", "error", err, "user_id", userID)
		return apperror.NewInternal("failed to verify user before applying: %w", err)
	}
	if user == nil {
		return apperror.NewNotFound("user with id '%s' not found", userID)
	}

	if err := s.repo.UpsertTeacherApplication(ctx, userID, phoneNumber, reason, teachingHistory); err != nil {
		s.logger.ErrorContext(ctx, "failed to apply for teacher", "error", err, "user_id", userID)
		return apperror.NewInternal("failed to submit teacher application: %w", err)
	}

	s.logger.InfoContext(ctx, "teacher verification application submitted",
		"user_id", userID,
		"phone_number", phoneNumber,
	)

	return nil
}

func (s *userServiceImpl) GetTeacherApplications(ctx context.Context, status string) ([]model.TeacherVerificationRequest, error) {
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		return nil, apperror.NewBadRequest("status must be one of: pending, approved, rejected")
	}

	applications, err := s.repo.ListTeacherApplications(ctx, status)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list teacher applications", "error", err)
		return nil, apperror.NewInternal("failed to list teacher applications: %w", err)
	}

	return applications, nil
}

func (s *userServiceImpl) ReviewTeacherApplication(ctx context.Context, requestID, reviewedBy string, req model.ReviewTeacherApplicationRequest) error {
	if requestID == "" {
		return apperror.NewBadRequest("request_id is required")
	}
	if reviewedBy == "" {
		return apperror.NewBadRequest("reviewed_by is required")
	}
	if req.Status != "approved" && req.Status != "rejected" {
		return apperror.NewBadRequest("status must be either 'approved' or 'rejected'")
	}

	if err := s.repo.ReviewTeacherApplication(ctx, requestID, req.Status, reviewedBy); err != nil {
		if err == sql.ErrNoRows {
			return apperror.NewNotFound("teacher application with id '%s' not found", requestID)
		}
		s.logger.ErrorContext(ctx, "failed to review teacher application", "error", err, "request_id", requestID)
		return apperror.NewInternal("failed to review teacher application: %w", err)
	}

	return nil
}
