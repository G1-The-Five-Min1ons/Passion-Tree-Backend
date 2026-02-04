package handler

import (
	"context"
	"passiontree/internal/learning-path/model"
	"passiontree/internal/pkg/apperror"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "fetching all learning paths")

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

func (h *Handler) GetOne(c *fiber.Ctx) error {
	id := c.Params("path_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	h.logger.InfoContext(ctx, "fetching learning path details", "path_id", id)

	path, err := h.pathSvc.GetPathDetails(ctx, id)
	if err != nil {
		return h.handleError(c, err)
	}
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path retrieved by one successfully",
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
		return h.handleError(c, apperror.NewInternal(err))
	}

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

func (h *Handler) Create(c *fiber.Ctx) error {
	var req model.CreatePathRequest
	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()
	
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	h.logger.InfoContext(ctx, "creating new learning path", "title", req.Title)

	if req.CoverImgURL == "" {
		return h.handleError(c, apperror.NewBadRequest("cover_image_url is required"))
	}

	if req.CoverImgURL != "" {
		if !strings.Contains(req.CoverImgURL, "learning-path") {
			 return h.handleError(c, apperror.NewBadRequest("Invalid image URL source"))
		}

		err := h.storage.ValidateUploadedFile(ctx, req.CoverImgURL, "learning-path")
		if err != nil {
			return h.handleError(c, apperror.NewBadRequest("Image validation failed: %v", err))
		}
	}

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

func (h *Handler) Start(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	var req model.StartPathRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	h.logger.InfoContext(ctx, "user enrolling in learning path", "user_id", req.UserID, "path_id", pathID)

	if err := h.pathSvc.StartPath(ctx, pathID, req.UserID); err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "user enrolled successfully", "user_id", req.UserID, "path_id", pathID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User enrolled in learning path successfully",
		"data": fiber.Map{
			"user_id": req.UserID,
			"path_id": pathID,
		},
	})
}

func (h *Handler) GetEnrollmentStatus(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID := c.Query("user_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if userID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id is required"))
	}

	h.logger.InfoContext(ctx, "checking enrollment status", "user_id", userID, "path_id", pathID)

	status, err := h.pathSvc.GetEnrollmentStatus(ctx, pathID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Enrollment status retrieved successfully",
		"data":    status,
	})
}

func (h *Handler) GetPathProgress(c *fiber.Ctx) error {
	pathID := c.Params("path_id")
	userID := c.Query("user_id")
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if userID == "" {
		return h.handleError(c, apperror.NewBadRequest("user_id is required"))
	}

	h.logger.InfoContext(ctx, "calculating path progress", "user_id", userID, "path_id", pathID)

	progress, err := h.pathSvc.GetPathProgress(ctx, pathID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Learning path progress calculated successfully",
		"data":    progress,
	})
}
func (h *Handler) Generate(c *fiber.Ctx) error {
	var req model.AIGeneratePathRequest

	ctx, cancel := context.WithTimeout(c.UserContext(), 60*time.Second)
	defer cancel()

	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	h.logger.InfoContext(ctx, "requesting AI learning path generation", "topic", req.Topic)

	result, err := h.pathSvc.GeneratePathWithAI(ctx, req.Topic)
	if err != nil {
		return h.handleError(c, err)
	}

	h.logger.InfoContext(ctx, "AI path generation completed", "topic", req.Topic)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Path generated successfully",
		"data":    result,
	})
}
