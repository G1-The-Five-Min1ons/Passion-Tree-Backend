package setting

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"passiontree/internal/connection"
	"passiontree/internal/worker"

	"github.com/gofiber/fiber/v2"
)

type createAnnouncementRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	PublishAt string `json:"publish_at"`
	IsActive  *bool  `json:"is_active"`
}

func registerDebugNotificationRoutes(r fiber.Router, db connection.Database, notificationWorker *worker.EmailNotificationWorker, logger *slog.Logger) {
	if os.Getenv("APP_ENV") == "production" || notificationWorker == nil {
		return
	}

	debug := r.Group("/debug")

	debug.Post("/notifications/daily", func(c *fiber.Ctx) error {
		go notificationWorker.RunDailyNotifications()
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"success": true,
			"message": "daily notifications triggered",
		})
	})

	debug.Post("/notifications/weekly", func(c *fiber.Ctx) error {
		go notificationWorker.RunWeeklyNotifications()
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"success": true,
			"message": "weekly notifications triggered",
		})
	})

	debug.Post("/notifications/announcements", func(c *fiber.Ctx) error {
		var req createAnnouncementRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "invalid request body",
			})
		}

		req.Title = strings.TrimSpace(req.Title)
		req.Content = strings.TrimSpace(req.Content)
		if req.Title == "" || req.Content == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "title and content are required",
			})
		}

		publishAt := time.Now().UTC()
		if strings.TrimSpace(req.PublishAt) != "" {
			parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.PublishAt))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "publish_at must be RFC3339 format",
				})
			}
			publishAt = parsed.UTC()
		}

		isActive := true
		if req.IsActive != nil {
			isActive = *req.IsActive
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
		defer cancel()

		query := `
			INSERT INTO platform_announcements (id, title, content, is_active, publish_at, created_at, updated_at)
			VALUES (NEWID(), @p1, @p2, @p3, @p4, GETDATE(), GETDATE())
		`

		if _, err := db.GetDB().ExecContext(ctx, query, req.Title, req.Content, isActive, publishAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "failed to create platform announcement",
				"error":   err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "platform announcement created",
			"data": fiber.Map{
				"title":      req.Title,
				"is_active":  isActive,
				"publish_at": publishAt,
			},
		})
	})

	logger.Info("debug notification trigger endpoints enabled")
}
