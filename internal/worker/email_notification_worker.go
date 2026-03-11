package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	authservice "passiontree/internal/auth/service"
	workerModel "passiontree/internal/worker/model"
)

const (
	settingEmailPlatformUpdates      = "email_notify_platform_updates"
	settingEmailWeeklyProgress       = "email_notify_weekly_progress"
	settingEmailDailyReminder        = "email_notify_daily_reminder"
	settingEmailRecommendations      = "email_notify_course_recommendations"
	settingEmailLearningPathComments = "email_notify_learning_path_comments"

	workerTimeout = 10 * time.Minute
)

type EmailNotificationWorker struct {
	provider     EmailNotificationDataProvider
	emailService authservice.EmailService
	userService  authservice.UserService
	logger       *slog.Logger
}

type EmailNotificationDataProvider interface {
	DailyReminderRecipients(ctx context.Context, settingKey string) ([]workerModel.NotificationRecipient, error)
	WeeklyProgressRows(ctx context.Context, settingKey string) ([]workerModel.WeeklyProgressRow, error)
	RecommendationNewPathCount(ctx context.Context) (int, error)
	RecipientsBySetting(ctx context.Context, settingKey string) ([]workerModel.NotificationRecipient, error)
	CommentNotificationRows(ctx context.Context, settingKey string) ([]workerModel.CommentNotificationRow, error)
}

func NewEmailNotificationWorker(provider EmailNotificationDataProvider, emailService authservice.EmailService, userService authservice.UserService, logger *slog.Logger) *EmailNotificationWorker {
	return &EmailNotificationWorker{provider: provider, emailService: emailService, userService: userService, logger: logger}
}

func (w *EmailNotificationWorker) RunDailyNotifications() {
	w.runJob("email_notification_daily", func(ctx context.Context) {
		w.sendDailyReminderEmails(ctx)
		w.sendCommentNotificationEmails(ctx)
		w.sendRecommendationEmails(ctx)
	})
}

func (w *EmailNotificationWorker) RunWeeklyNotifications() {
	w.runJob("email_notification_weekly", func(ctx context.Context) {
		w.sendWeeklyProgressEmails(ctx)
		w.sendPlatformUpdateEmails(ctx)
	})
}

func (w *EmailNotificationWorker) sendDailyReminderEmails(ctx context.Context) {
	if !w.isEmailServiceReady("daily_reminder") {
		return
	}

	recipients, err := w.provider.DailyReminderRecipients(ctx, settingEmailDailyReminder)
	if err != nil {
		w.logger.Error("daily_reminder_query_failed", "error", err)
		return
	}

	for _, recipient := range recipients {
		// Skip email if user account is deactivated
		if isDeactivated, _ := w.isUserDeactivated(ctx, recipient.UserID); isDeactivated {
			w.logger.Info("daily_reminder_skipped_deactivated", "user_id", recipient.UserID)
			continue
		}

		subject := "Daily Reminder · Keep your learning streak alive"
		name := displayName(recipient.FirstName)
		headline := "You missed practice today"
		body := fmt.Sprintf("Hi %s, it looks like you haven’t practiced in the past 24 hours. A short session today can keep your momentum going.", name)
		if err := w.emailService.SendNotificationEmail(ctx, recipient.Email, subject, headline, body); err != nil {
			w.logger.Error("daily_reminder_send_failed", "user_id", recipient.UserID, "error", err)
		}
	}

	w.logger.Info("daily_reminder_completed", "recipient_count", len(recipients))
}

func (w *EmailNotificationWorker) sendWeeklyProgressEmails(ctx context.Context) {
	if !w.isEmailServiceReady("weekly_progress") {
		return
	}

	recipients, err := w.provider.WeeklyProgressRows(ctx, settingEmailWeeklyProgress)
	if err != nil {
		w.logger.Error("weekly_progress_query_failed", "error", err)
		return
	}

	for _, recipient := range recipients {
		// Skip email if user account is deactivated
		if isDeactivated, _ := w.isUserDeactivated(ctx, recipient.UserID); isDeactivated {
			w.logger.Info("weekly_progress_skipped_deactivated", "user_id", recipient.UserID)
			continue
		}

		name := displayName(recipient.FirstName)

		headline := "Your weekly progress report is ready"
		message := fmt.Sprintf("Hi %s! In the last 7 days, you completed %d nodes across %d active days.", name, recipient.CompletedNodes, recipient.ActiveDays)
		if err := w.emailService.SendNotificationEmail(ctx, recipient.Email, "Weekly Progress Report · Passion-Tree", headline, message); err != nil {
			w.logger.Error("weekly_progress_send_failed", "user_id", recipient.UserID, "error", err)
		}
	}

	w.logger.Info("weekly_progress_completed", "recipient_count", len(recipients))
}

func (w *EmailNotificationWorker) sendRecommendationEmails(ctx context.Context) {
	if !w.isEmailServiceReady("recommendations") {
		return
	}

	newPathCount, err := w.provider.RecommendationNewPathCount(ctx)
	if err != nil {
		w.logger.Warn("recommendation_count_query_failed", "error", err)
		return
	}

	if newPathCount <= 0 {
		return
	}

	recipients, err := w.provider.RecipientsBySetting(ctx, settingEmailRecommendations)
	if err != nil {
		w.logger.Error("recommendation_recipient_query_failed", "error", err)
		return
	}

	for _, recipient := range recipients {
		// Skip email if user account is deactivated
		if isDeactivated, _ := w.isUserDeactivated(ctx, recipient.UserID); isDeactivated {
			w.logger.Info("recommendation_skipped_deactivated", "user_id", recipient.UserID)
			continue
		}

		message := fmt.Sprintf("There are %d new learning paths this week. Explore your personalized recommendations now.", newPathCount)
		if err := w.emailService.SendNotificationEmail(ctx, recipient.Email, "Course Recommendations · Passion-Tree", "New recommendations available", message); err != nil {
			w.logger.Error("recommendation_send_failed", "user_id", recipient.UserID, "error", err)
		}
	}

	w.logger.Info("recommendation_emails_completed", "recipient_count", len(recipients), "new_paths", newPathCount)
}

func (w *EmailNotificationWorker) sendCommentNotificationEmails(ctx context.Context) {
	if !w.isEmailServiceReady("new_comments") {
		return
	}

	recipients, err := w.provider.CommentNotificationRows(ctx, settingEmailLearningPathComments)
	if err != nil {
		w.logger.Warn("comment_notification_query_failed", "error", err)
		return
	}

	for _, recipient := range recipients {
		// Skip email if user account is deactivated
		if isDeactivated, _ := w.isUserDeactivated(ctx, recipient.UserID); isDeactivated {
			w.logger.Info("comment_notification_skipped_deactivated", "user_id", recipient.UserID)
			continue
		}

		message := fmt.Sprintf("You have %d new comments on your learning paths.", recipient.NewComments)
		if err := w.emailService.SendNotificationEmail(ctx, recipient.Email, "New Comments · Passion-Tree", "New comments on your learning paths", message); err != nil {
			w.logger.Error("comment_notification_send_failed", "user_id", recipient.UserID, "error", err)
		}
	}

	w.logger.Info("comment_notifications_completed", "recipient_count", len(recipients))
}

func (w *EmailNotificationWorker) sendPlatformUpdateEmails(ctx context.Context) {
	if !w.isEmailServiceReady("platform_updates") {
		return
	}

	recipients, err := w.provider.RecipientsBySetting(ctx, settingEmailPlatformUpdates)
	if err != nil {
		w.logger.Error("platform_update_recipient_query_failed", "error", err)
		return
	}

	for _, recipient := range recipients {
		// Skip email if user account is deactivated
		if isDeactivated, _ := w.isUserDeactivated(ctx, recipient.UserID); isDeactivated {
			w.logger.Info("platform_update_skipped_deactivated", "user_id", recipient.UserID)
			continue
		}

		message := "Platform updates are available this week. Visit Passion-Tree to see what’s new."
		if err := w.emailService.SendNotificationEmail(ctx, recipient.Email, "Platform Updates · Passion-Tree", "Platform updates", message); err != nil {
			w.logger.Error("platform_update_send_failed", "user_id", recipient.UserID, "error", err)
		}
	}

	w.logger.Info("platform_update_completed", "recipient_count", len(recipients))
}

func (w *EmailNotificationWorker) runJob(job string, action func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), workerTimeout)
	defer cancel()

	w.logger.Info(job + "_start")
	action(ctx)
	w.logger.Info(job + "_done")
}

func (w *EmailNotificationWorker) isEmailServiceReady(job string) bool {
	if w.emailService != nil {
		return true
	}

	w.logger.Warn("email_service_not_configured", "job", job)
	return false
}

func displayName(firstName string) string {
	name := strings.TrimSpace(firstName)
	if name == "" {
		return "there"
	}

	return name
}

// isUserDeactivated checks if a user's account is currently deactivated
func (w *EmailNotificationWorker) isUserDeactivated(ctx context.Context, userID string) (bool, error) {
	if w.userService == nil {
		return false, nil
	}

	deactivatedUntil, err := w.userService.GetAccountDeactivatedUntil(ctx, userID)
	if err != nil {
		w.logger.Warn("check_deactivation_status_failed", "user_id", userID, "error", err)
		return false, err
	}

	if deactivatedUntil == nil {
		return false, nil
	}

	return deactivatedUntil.After(time.Now().UTC()), nil
}
