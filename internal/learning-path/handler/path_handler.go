package handler

import (
	"context"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetAll godoc
// @Summary      List all learning paths
// @Description  Returns every published learning path (public catalog).
// @Tags         Learning Paths
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      500  {object}  apidoc.ErrorResponse
// @Router       /learningpaths [get]
func (h *Handler) GetAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	paths, err := h.pathSvc.GetPaths(ctx)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "successfully retrieved all learning paths", "count", len(paths))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning paths retrieved all successfully",
		"data":    paths,
	})
}

// GetOne godoc
// @Summary      Get a learning path by ID
// @Description  Returns full details (nodes, materials, ratings) of a single learning path.
// @Tags         Learning Paths
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string  true  "Learning path ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/{path_id} [get]
func (h *Handler) GetOne(c *fiber.Ctx) error {
	id := c.Params("path_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	path, err := h.pathSvc.GetPathDetails(ctx, id)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "successfully retrieved learning path details", "path_id", id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path retrieved successfully",
		"data":    path,
	})
}

func (h *Handler) GetUploadURL(c *fiber.Ctx) error {
	filename := c.Query("filename")
	if filename == "" {
		return h.handleError(c, apperror.NewBadRequest("filename is required"))
	}

	if !strings.HasSuffix(strings.ToLower(filename), ".jpg") &&
		!strings.HasSuffix(strings.ToLower(filename), ".png") &&
		!strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
		return h.handleError(c, apperror.NewBadRequest("Only JPEG and PNG images are allowed"))
	}

	uploadURL, publicURL, err := h.storage.GeneratePresignedURL(filename, "learning-path", 15*time.Minute)
	if err != nil {
		return h.handleError(c, apperror.NewInternal("failed to generate presigned URL for file %s: %w", filename, err))
	}

	h.logger.InfoContext(c.UserContext(), "generated presigned URL for learning path cover image", "filename", filename)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Presigned URL generated successfully",
		"data": fiber.Map{
			"upload_url": uploadURL,
			"public_url": publicURL,
			"expires_in": "15m",
		},
	})
}

// Create godoc
// @Summary      Create a learning path
// @Description  Creates a new learning path owned by the authenticated user.
// @Tags         Learning Paths
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreatePathRequest  true  "Learning path payload"
// @Success      201   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /learningpaths [post]
func (h *Handler) Create(c *fiber.Ctx) error {
	var req model.CreatePathRequest
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	req.CreatorID = userID

	id, err := h.pathSvc.CreatePath(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "learning path created successfully", "path_id", id)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Learning path created successfully",
		"data": fiber.Map{
			"path_id": id,
		},
	})
}

// Update godoc
// @Summary      Update a learning path
// @Tags         Learning Paths
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string                   true  "Learning path ID"
// @Param        body     body      model.UpdatePathRequest  true  "Updated fields"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/{path_id} [put]
func (h *Handler) Update(c *fiber.Ctx) error {
	path_id := c.Params("path_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	var req model.UpdatePathRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	h.logger.InfoContext(ctx, "updating learning path", "path_id", path_id)

	if err := h.pathSvc.UpdatePath(ctx, path_id, req); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "learning path updated successfully", "path_id", path_id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path updated successfully",
		"data": fiber.Map{
			"path_id": path_id,
		},
	})
}

// Delete godoc
// @Summary      Delete a learning path
// @Tags         Learning Paths
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string  true  "Learning path ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      403      {object}  apidoc.ErrorResponse  "Not the owner"
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/{path_id} [delete]
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("path_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "deleting learning path", "path_id", id)

	if err := h.pathSvc.DeletePath(ctx, id); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "learning path deleted successfully", "path_id", id)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path deleted successfully",
		"data": fiber.Map{
			"path_id": id,
		},
	})
}

// Start godoc
// @Summary      Enroll in a learning path
// @Description  Enrolls the authenticated user in the given learning path.
// @Tags         Learning Paths
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string  true  "Learning path ID"
// @Success      201      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Failure      409      {object}  apidoc.ErrorResponse  "Already enrolled"
// @Router       /learningpaths/{path_id}/start [post]
func (h *Handler) Start(c *fiber.Ctx) error {
	pathID := c.Params("path_id")

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.pathSvc.StartPath(ctx, pathID, userID); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "user enrolled in learning path successfully", "user_id", userID, "path_id", pathID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User enrolled in learning path successfully",
		"data": fiber.Map{
			"user_id": userID,
			"path_id": pathID,
		},
	})
}

// GetEnrollmentStatus godoc
// @Summary      Get enrollment status for a path
// @Tags         Learning Paths
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string  true  "Learning path ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /user/learningpaths/{path_id}/status [get]
func (h *Handler) GetEnrollmentStatus(c *fiber.Ctx) error {
	pathID := c.Params("path_id")

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	status, err := h.pathSvc.GetEnrollmentStatus(ctx, pathID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "retrieved enrollment status successfully", "user_id", userID, "path_id", pathID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Enrollment status retrieved successfully",
		"data":    status,
	})
}

// GetPathProgress godoc
// @Summary      Get learning path progress
// @Description  Returns completion percentage and per-node status for the authenticated user.
// @Tags         Learning Paths
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string  true  "Learning path ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /user/learningpaths/{path_id}/progress [get]
func (h *Handler) GetPathProgress(c *fiber.Ctx) error {
	pathID := c.Params("path_id")

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	progress, err := h.pathSvc.GetPathProgress(ctx, pathID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "calculated path progress successfully", "user_id", userID, "path_id", pathID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path progress calculated successfully",
		"data":    progress,
	})
}
// Generate godoc
// @Summary      Generate a learning path with AI
// @Description  Calls the AI service to draft a complete learning path for the given topic.
// @Tags         Learning Paths
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.AIGeneratePathRequest  true  "Topic to generate"
// @Success      200   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Failure      502   {object}  apidoc.ErrorResponse  "AI service error"
// @Router       /learningpaths/generate [post]
func (h *Handler) Generate(c *fiber.Ctx) error {
	var req model.AIGeneratePathRequest

	ctx, cancel := context.WithTimeout(c.UserContext(), 60*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	result, err := h.pathSvc.GeneratePathWithAI(ctx, req.Topic)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "AI learning path generated completed", "topic", req.Topic)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Path generated successfully",
		"data":    result,
	})
}

// UpdateCoverImage godoc
// @Summary      Update learning path cover image URL
// @Tags         Learning Paths
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.UpdateImageRequest  true  "New cover image URL"
// @Success      200   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /learningpaths/uploadimg [put]
func (h *Handler) UpdateCoverImage(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.UpdateImageRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if err := h.pathSvc.UpdatePathCoverImage(ctx, pathID, req.CoverImgURL); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "learning path cover image updated successfully", "path_id", pathID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path cover image updated successfully",
		"data": fiber.Map{
			"path_id":         pathID,
			"cover_image_url": req.CoverImgURL,
		},
	})
}

// GetMyPaths godoc
// @Summary      List paths the authenticated user is enrolled in
// @Tags         Learning Paths
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.SuccessResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Router       /learningpaths/user/enroll [get]
func (h *Handler) GetMyPaths(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	paths, err := h.pathSvc.GetUserEnrolledPaths(ctx, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Retrieved enrolled paths successfully",
		"data":    paths,
	})
}

// SubmitRating godoc
// @Summary      Submit or update a rating for a learning path
// @Tags         Ratings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string                true  "Learning path ID"
// @Param        body     body      model.RatingRequest   true  "Rating value and review"
// @Success      200      {object}  apidoc.MessageResponse
// @Failure      400      {object}  apidoc.ErrorResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/{path_id}/ratings [post]
func (h *Handler) SubmitRating(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req model.RatingRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.ratingSvc.UpsertRating(ctx, pathID, userID, req); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Rating submitted successfully",
	})
}

// GetMyRating godoc
// @Summary      Get the authenticated user's rating for a path
// @Tags         Ratings
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string  true  "Learning path ID"
// @Success      200      {object}  apidoc.SuccessResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/{path_id}/ratings [get]
func (h *Handler) GetMyRating(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	rating, err := h.ratingSvc.GetMyRating(ctx, pathID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    rating,
	})
}

// DeleteRating godoc
// @Summary      Remove the authenticated user's rating from a path
// @Tags         Ratings
// @Produce      json
// @Security     BearerAuth
// @Param        path_id  path      string  true  "Learning path ID"
// @Success      200      {object}  apidoc.MessageResponse
// @Failure      401      {object}  apidoc.ErrorResponse
// @Failure      404      {object}  apidoc.ErrorResponse
// @Router       /learningpaths/{path_id}/ratings [delete]
func (h *Handler) DeleteRating(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.ratingSvc.DeleteRating(ctx, pathID, userID); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Rating deleted successfully",
	})
}
