package handler

import (
	"context"
	"time"

	"passiontree/internal/pkg/apperror"
	"passiontree/internal/reflection/model"

	"github.com/gofiber/fiber/v2"
	"passiontree/internal/pkg/middleware"
)

// Create godoc
// @Summary      Create a reflection
// @Description  Creates a new reflection (a journal-style entry) for the authenticated user.
// @Tags         Reflections
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateReflectionRequest  true  "Reflection payload"
// @Success      201   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /reflections [post]
func (h *Handler) Create(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req model.CreateReflectionRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	res, err := h.reflectSvc.CreateReflection(ctx, req, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflection created successfully", "reflect_id", res)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "reflection created successfully",
		"data":    res,
	})
}

// Update godoc
// @Summary      Update a reflection
// @Tags         Reflections
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        reflect_id  path      string                         true  "Reflection ID"
// @Param        body        body      model.UpdateReflectionRequest  true  "Updated fields"
// @Success      200         {object}  apidoc.SuccessResponse
// @Failure      400         {object}  apidoc.ErrorResponse
// @Failure      401         {object}  apidoc.ErrorResponse
// @Failure      404         {object}  apidoc.ErrorResponse
// @Router       /reflections/{reflect_id} [put]
func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("reflect_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.UpdateReflectionRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.reflectSvc.UpdateReflection(ctx, id, req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflection updated successfully", "reflect_id", id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection updated successfully",
		"data": fiber.Map{
			"reflect_id": id,
		},
	})
}

// Delete godoc
// @Summary      Delete a reflection
// @Tags         Reflections
// @Produce      json
// @Security     BearerAuth
// @Param        reflect_id  path      string  true  "Reflection ID"
// @Success      200         {object}  apidoc.SuccessResponse
// @Failure      401         {object}  apidoc.ErrorResponse
// @Failure      404         {object}  apidoc.ErrorResponse
// @Router       /reflections/{reflect_id} [delete]
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("reflect_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.reflectSvc.DeleteReflection(ctx, id); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflection deleted successfully", "reflect_id", id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection deleted successfully",
		"data": fiber.Map{
			"reflect_id": id,
		},
	})
}

// GetByID godoc
// @Summary      Get a single reflection
// @Tags         Reflections
// @Produce      json
// @Security     BearerAuth
// @Param        reflect_id  path      string  true  "Reflection ID"
// @Success      200         {object}  apidoc.SuccessResponse
// @Failure      401         {object}  apidoc.ErrorResponse
// @Failure      404         {object}  apidoc.ErrorResponse
// @Router       /reflections/{reflect_id} [get]
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("reflect_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	res, err := h.reflectSvc.GetReflectionByID(ctx, id)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflection retrieved successfully", "reflect_id", id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflection retrieved successfully",
		"data":    res,
	})
}

// GetAll godoc
// @Summary      List reflections with keyset pagination
// @Description  Supports filtering by tree_node_id / tree_id / album_id / user_id and keyset pagination via before_created_at + before_reflect_id (must be passed together).
// @Tags         Reflections
// @Produce      json
// @Security     BearerAuth
// @Param        limit              query     int     false  "Page size (default 50)"
// @Param        before_created_at  query     string  false  "RFC3339Nano timestamp for keyset pagination"
// @Param        before_reflect_id  query     string  false  "Reflection ID for keyset pagination"
// @Param        tree_node_id       query     string  false  "Filter by tree node ID"
// @Param        tree_id            query     string  false  "Filter by tree ID"
// @Param        album_id           query     string  false  "Filter by album ID"
// @Param        user_id            query     string  false  "Filter by user ID"
// @Success      200                {object}  apidoc.SuccessResponse
// @Failure      400                {object}  apidoc.ErrorResponse
// @Failure      401                {object}  apidoc.ErrorResponse
// @Router       /reflections [get]
func (h *Handler) GetAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	// Parse query parameters for filtering.
	// Pagination uses keyset/seek method via before_created_at + before_reflect_id.
	filter := model.GetReflectionsFilter{
		Limit: c.QueryInt("limit", 50), // Default limit 50
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	beforeCreatedAtRaw := c.Query("before_created_at")
	beforeReflectID := c.Query("before_reflect_id")
	if (beforeCreatedAtRaw == "") != (beforeReflectID == "") {
		return h.handleError(c, apperror.NewBadRequest("before_created_at and before_reflect_id must be provided together"))
	}
	if beforeCreatedAtRaw != "" {
		beforeCreatedAt, err := time.Parse(time.RFC3339Nano, beforeCreatedAtRaw)
		if err != nil {
			return h.handleError(c, apperror.NewBadRequest("before_created_at must be RFC3339 format"))
		}
		filter.BeforeCreatedAt = &beforeCreatedAt
		filter.BeforeReflectID = &beforeReflectID
	}

	// Optional filters
	if treeNodeID := c.Query("tree_node_id"); treeNodeID != "" {
		filter.TreeNodeID = &treeNodeID
	}
	if treeID := c.Query("tree_id"); treeID != "" {
		filter.TreeID = &treeID
	}
	if albumID := c.Query("album_id"); albumID != "" {
		filter.AlbumID = &albumID
	}
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = &userID
	}

	res, err := h.reflectSvc.GetAllReflections(ctx, filter)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "reflections retrieved", "count", len(res), "filter", filter)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reflections retrieved successfully",
		"data":    res,
		"count":   len(res),
	})
}
