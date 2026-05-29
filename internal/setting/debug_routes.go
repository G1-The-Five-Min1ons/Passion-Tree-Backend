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

type debugNotificationHandler struct {
	db                 connection.Database
	notificationWorker *worker.EmailNotificationWorker
}

// TriggerDailyNotifications godoc
// @Summary      Trigger daily notifications (debug)
// @Description  Triggers the daily notification worker. Available only in non-production.
// @Tags         Debug
// @Produce      json
// @Success      202  {object}  apidoc.MessageResponse
// @Failure      500  {object}  apidoc.ErrorResponse
// @Router       /debug/notifications/daily [post]
func (h *debugNotificationHandler) TriggerDailyNotifications(c *fiber.Ctx) error {
	go h.notificationWorker.RunDailyNotifications()
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": "daily notifications triggered",
	})
}

// TriggerWeeklyNotifications godoc
// @Summary      Trigger weekly notifications (debug)
// @Description  Triggers the weekly notification worker. Available only in non-production.
// @Tags         Debug
// @Produce      json
// @Success      202  {object}  apidoc.MessageResponse
// @Failure      500  {object}  apidoc.ErrorResponse
// @Router       /debug/notifications/weekly [post]
func (h *debugNotificationHandler) TriggerWeeklyNotifications(c *fiber.Ctx) error {
	go h.notificationWorker.RunWeeklyNotifications()
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": "weekly notifications triggered",
	})
}

// CreatePlatformAnnouncement godoc
// @Summary      Create platform announcement (debug)
// @Description  Creates a platform announcement and schedules it for notification delivery. Available only in non-production.
// @Tags         Debug
// @Accept       json
// @Produce      json
// @Param        body  body      createAnnouncementRequest  true  "Announcement payload"
// @Success      201   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      500   {object}  apidoc.ErrorResponse
// @Router       /debug/notifications/announcements [post]
func (h *debugNotificationHandler) CreatePlatformAnnouncement(c *fiber.Ctx) error {
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

	if _, err := h.db.GetDB().ExecContext(ctx, query, req.Title, req.Content, isActive, publishAt); err != nil {
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
}

func registerDebugNotificationRoutes(r fiber.Router, db connection.Database, notificationWorker *worker.EmailNotificationWorker, logger *slog.Logger) {
	if os.Getenv("APP_ENV") == "production" || notificationWorker == nil {
		return
	}

	h := &debugNotificationHandler{
		db:                 db,
		notificationWorker: notificationWorker,
	}

	debug := r.Group("/debug")

	debug.Post("/notifications/daily", h.TriggerDailyNotifications)
	debug.Post("/notifications/weekly", h.TriggerWeeklyNotifications)
	debug.Post("/notifications/announcements", h.CreatePlatformAnnouncement)

	logger.Info("debug notification trigger endpoints enabled")
}
