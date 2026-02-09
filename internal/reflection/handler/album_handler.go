package handler

import (
	"context"
	"time"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"

	"github.com/gofiber/fiber/v2"
)

// CreateAlbum handles the creation of a new album
func (h *Handler) CreateAlbum(c *fiber.Ctx) error {
	var req model.CreateAlbumRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	resp, err := h.reflectSvc.CreateAlbum(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "album created successfully", "album_id", resp)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "album created successfully",
		"data": fiber.Map{
			"album": resp,
		},
	})
}

// GetAlbumByID handles retrieving an album by its ID
func (h *Handler) GetAlbumByID(c *fiber.Ctx) error {
	albumID := c.Params("album_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	album, err := h.reflectSvc.GetAlbumByID(ctx, albumID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "album retrieved successfully", "album_id", albumID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "album retrieved successfully",
		"data": fiber.Map{
			"album": album,
		},
	})
}

// GetAlbumsByUserID handles retrieving all albums for a user
func (h *Handler) GetAlbumsByUserID(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	albums, err := h.reflectSvc.GetAlbumsByUserID(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "albums retrieved successfully", "user_id", userID, "count", len(albums))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "albums retrieved successfully",
		"data": fiber.Map{
			"albums": albums,
			"count":  len(albums),
		},
	})
}

// UpdateAlbum handles updating an existing album
func (h *Handler) UpdateAlbum(c *fiber.Ctx) error {
	albumID := c.Params("album_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.UpdateAlbumRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	err := h.reflectSvc.UpdateAlbum(ctx, albumID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "album updated successfully", "album_id", albumID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "album updated successfully",
		"data": fiber.Map{
			"album_id": albumID,
		},
	})
}

// DeleteAlbum handles deleting an album
func (h *Handler) DeleteAlbum(c *fiber.Ctx) error {
	albumID := c.Params("album_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	err := h.reflectSvc.DeleteAlbum(ctx, albumID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "album deleted successfully", "album_id", albumID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "album deleted successfully",
		"data": fiber.Map{
			"album_id": albumID,
		},
	})
}
