package service

import (
	"context"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/upload/model"
	"strings"
	"time"
)

func (s *serviceImpl) GeneratePresignedURL(ctx context.Context, req model.UploadImageRequest) (*model.UploadImageResponse, error) {
	if req.Filename == "" {
		return nil, apperror.NewBadRequest("filename is required")
	}

	if req.Folder == "" {
		req.Folder = "general"
	}

	if !strings.HasSuffix(strings.ToLower(req.Filename), ".jpg") &&
		!strings.HasSuffix(strings.ToLower(req.Filename), ".png") &&
		!strings.HasSuffix(strings.ToLower(req.Filename), ".jpeg") {
		return nil, apperror.NewBadRequest("Only JPEG and PNG images are allowed")
	}

	s.logger.InfoContext(ctx, "generating presigned URL", "filename", req.Filename, "folder", req.Folder)

	uploadURL, publicURL, err := s.storage.GeneratePresignedURL(req.Filename, req.Folder, 15*time.Minute)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate presigned URL", "error", err, "filename", req.Filename)
		return nil, apperror.NewInternal("failed to generate presigned URL for file %s: %w", req.Filename, err)
	}

	return &model.UploadImageResponse{
		UploadURL: uploadURL,
		PublicURL: publicURL,
		ExpiresIn: "15m",
		Folder:    req.Folder,
	}, nil
}
