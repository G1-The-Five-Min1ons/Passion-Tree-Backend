package handler

import (
	"context"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"
	"passiontree/internal/reflection/model"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CreateTree godoc
// @Summary      Create a tree (within an album)
// @Tags         Trees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateTreeRequest  true  "Tree payload (album_id required)"
// @Success      201   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /trees [post]
func (h *Handler) CreateTree(c *fiber.Ctx) error {
	var req model.CreateTreeRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	resp, err := h.reflectSvc.CreateTree(ctx, req, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree created successfully", "tree_id", resp.TreeID, "album_id", req.AlbumID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "tree created successfully",
		"data":    resp,
	})
}

// GetTreeByID godoc
// @Summary      Get a tree by ID
// @Tags         Trees
// @Produce      json
// @Security     BearerAuth
// @Param        tree_id  path      string  true  "Tree ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /trees/{tree_id} [get]
func (h *Handler) GetTreeByID(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "fetching tree details", "tree_id", treeID)

	tree, err := h.reflectSvc.GetTreeByID(ctx, treeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree retrieved successfully",
		"data": fiber.Map{
			"tree": tree,
		},
	})
}

// GetTreesByAlbumID godoc
// @Summary      List trees in an album
// @Tags         Trees
// @Produce      json
// @Security     BearerAuth
// @Param        album_id       query     string  true   "Album ID"
// @Param        include_nodes  query     bool    false  "Include child tree nodes in the response"
// @Success      200            {object}  apidoc.SuccessResponse
// @Failure      400            {object}  apidoc.ErrorResponse
// @Failure      401            {object}  apidoc.ErrorResponse
// @Router       /trees [get]
func (h *Handler) GetTreesByAlbumID(c *fiber.Ctx) error {
	albumID := c.Query("album_id")
	includeNodes := c.QueryBool("include_nodes", false)
	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	if albumID == "" {
		return h.handleError(c, apperror.NewBadRequest("album_id is required as query parameter"))
	}

	// Get user_id from auth middleware
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	trees, err := h.reflectSvc.GetTreesByAlbumID(ctx, albumID, includeNodes, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	// Count handling based on type
	var count int
	switch v := trees.(type) {
	case []model.Tree:
		count = len(v)
	case []model.TreeResponse:
		count = len(v)
	}

	h.logger.InfoContext(ctx, "successfully retrieved trees for album", "album_id", albumID, "count", count, "include_nodes", includeNodes)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "trees retrieved successfully",
		"data": fiber.Map{
			"trees": trees,
			"count": count,
		},
	})
}

// UpdateTree godoc
// @Summary      Update a tree
// @Tags         Trees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        tree_id  path      string                   true  "Tree ID"
// @Param        body     body      model.UpdateTreeRequest  true  "Updated fields"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /trees/{tree_id} [put]
func (h *Handler) UpdateTree(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.UpdateTreeRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	err = h.reflectSvc.UpdateTree(ctx, treeID, req, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree updated successfully", "tree_id", treeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree updated successfully",
		"data": fiber.Map{
			"tree_id": treeID,
		},
	})
}

// RetrieveTree godoc
// @Summary      Retrieve (revive) a dead tree by spending hearts
// @Tags         Trees
// @Produce      json
// @Security     BearerAuth
// @Param        tree_id  path      string  true  "Tree ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse  "Not enough hearts or tree not in dead state"
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /trees/{tree_id}/retrieve [patch]
func (h *Handler) RetrieveTree(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	resp, err := h.reflectSvc.RetrieveTree(ctx, treeID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree retrieved successfully", "tree_id", treeID, "user_id", userID, "remaining_hearts", resp.HeartCount)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree retrieved successfully",
		"data":    resp,
	})
}

// EndReflecting godoc
// @Summary      End the reflection cycle for a tree
// @Description  Freezes a tree's reflection status — locks the tree from further reflections and marks it complete.
// @Tags         Trees
// @Produce      json
// @Security     BearerAuth
// @Param        tree_id  path      string  true  "Tree ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /trees/{tree_id}/end-reflecting [patch]
func (h *Handler) EndReflecting(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	resp, err := h.reflectSvc.EndReflecting(ctx, treeID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree reflection ended successfully", "tree_id", treeID, "user_id", userID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree reflection ended successfully",
		"data":    resp,
	})
}

// DeleteTree godoc
// @Summary      Delete a tree
// @Tags         Trees
// @Produce      json
// @Security     BearerAuth
// @Param        tree_id  path      string  true  "Tree ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /trees/{tree_id} [delete]
func (h *Handler) DeleteTree(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	err = h.reflectSvc.DeleteTree(ctx, treeID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree deleted successfully", "tree_id", treeID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree deleted successfully",
		"data": fiber.Map{
			"tree_id": treeID,
		},
	})
}

// PauseTree godoc
// @Summary      Toggle pause/resume state of a tree
// @Tags         Trees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        tree_id  path      string                  true   "Tree ID"
// @Param        body     body      model.PauseTreeRequest  false  "Optional pause options"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /trees/{tree_id}/pause [patch]
func (h *Handler) PauseTree(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	var req model.PauseTreeRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return h.handleError(c, apperror.NewBadRequest("invalid request body"))
		}
	}

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	resp, err := h.reflectSvc.PauseTree(ctx, treeID, userID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	action := "paused"
	if !resp.IsPause {
		action = "resumed"
	}

	h.logger.InfoContext(ctx, "tree state updated successfully", "action", action, "tree_id", treeID, "user_id", userID, "paused_at", resp.PausedAt, "remaining_hearts", resp.HeartCount)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree " + action + " successfully",
		"data":    resp,
	})
}

// CalculateTreeScore godoc
// @Summary      Calculate and persist a tree's score
// @Description  Computes the weighted average score (0–10 scale) for the tree's reflections and stores it.
// @Tags         Trees
// @Produce      json
// @Security     BearerAuth
// @Param        tree_id  path      string  true  "Tree ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /trees/{tree_id}/score [patch]
func (h *Handler) CalculateTreeScore(c *fiber.Ctx) error {
	treeID := c.Params("tree_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	score, err := h.reflectSvc.CalculateAndUpdateTreeScore(ctx, treeID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "tree score calculated", "tree_id", treeID, "score", score)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "tree score calculated successfully",
		"data": fiber.Map{
			"tree_id":    treeID,
			"tree_score": score,
		},
	})
}
