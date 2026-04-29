package handler

import (
	"context"
	"time"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"
	"passiontree/internal/reflection/model"

	"github.com/gofiber/fiber/v2"
)

// CreateAlbum godoc
// @Summary      Create an album
// @Tags         Albums
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateAlbumRequest  true  "Album payload"
// @Success      201   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /albums [post]
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

// GetAlbumByID godoc
// @Summary      Get an album by ID
// @Tags         Albums
// @Produce      json
// @Security     BearerAuth
// @Param        album_id  path      string  true  "Album ID"
// @Success      200       {object}  apidoc.SuccessResponse
// @Failure      401       {object}  apidoc.ErrorResponse
// @Failure      404       {object}  apidoc.ErrorResponse
// @Router       /albums/{album_id} [get]
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

// GetAlbumsByUserID godoc
// @Summary      List albums for the authenticated user
// @Tags         Albums
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Router       /albums [get]
func (h *Handler) GetAlbumsByUserID(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

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

// UpdateAlbum godoc
// @Summary      Update an album
// @Tags         Albums
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        album_id  path      string                    true  "Album ID"
// @Param        body      body      model.UpdateAlbumRequest  true  "Updated fields"
// @Success      200       {object}  apidoc.SuccessResponse
// @Failure      400       {object}  apidoc.ErrorResponse
// @Failure      401       {object}  apidoc.ErrorResponse
// @Failure      404       {object}  apidoc.ErrorResponse
// @Router       /albums/{album_id} [put]
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

// DeleteAlbum godoc
// @Summary      Delete an album
// @Tags         Albums
// @Produce      json
// @Security     BearerAuth
// @Param        album_id  path      string  true  "Album ID"
// @Success      200       {object}  apidoc.SuccessResponse
// @Failure      401       {object}  apidoc.ErrorResponse
// @Failure      404       {object}  apidoc.ErrorResponse
// @Router       /albums/{album_id} [delete]
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
