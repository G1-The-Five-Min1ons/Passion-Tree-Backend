package service

import (
	"context"
	"log/slog"
	"passiontree/internal/pkg/storage"
	"passiontree/internal/upload/model"
)

type Service interface {
	GeneratePresignedURL(ctx context.Context, req model.UploadImageRequest) (*model.UploadImageResponse, error)
}

type serviceImpl struct {
	logger  *slog.Logger
	storage *storage.BlobService
}

func NewService(logger *slog.Logger, storage *storage.BlobService) Service {
	return &serviceImpl{
		logger:  logger,
		storage: storage,
	}
}
