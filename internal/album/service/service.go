package service

import (
	"context"
	"passiontree/internal/album/model"
	"passiontree/internal/album/repository"
)

type AlbumService interface {
	CreateAlbum(ctx context.Context, req model.CreateAlbumRequest) (string, error)
	GetAlbumByID(ctx context.Context, albumID string) (*model.TreeAlbum, error)
	GetAlbumsByUserID(ctx context.Context, userID string) ([]model.TreeAlbum, error)
	UpdateAlbum(ctx context.Context, albumID string, req model.UpdateAlbumRequest) error
	DeleteAlbum(ctx context.Context, albumID string) error
	
	GetAlbumWithTrees(ctx context.Context, albumID string) (*model.AlbumWithTreesResponse, error)
	AddTreeToAlbum(ctx context.Context, albumID string, req model.AddTreeToAlbumRequest) error
	RemoveTreeFromAlbum(ctx context.Context, treeID string) error
}

type serviceImpl struct {
	repo repository.AlbumRepository
}

func NewService(repo repository.AlbumRepository) AlbumService {
	return &serviceImpl{
		repo: repo,
	}
}
