package service_test

import "context"

// EmailService is a test double for service.EmailService.
type EmailService struct {
	SendPasswordResetEmailFunc func(ctx context.Context, to, token string) error
	SendVerificationEmailFunc  func(ctx context.Context, to, token string) error
	SendSecurityAlertEmailFunc func(ctx context.Context, to, userID string) error
	SendNotificationEmailFunc  func(ctx context.Context, to, subject, headline, message string) error
}

func (m *EmailService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	if m.SendPasswordResetEmailFunc != nil {
		return m.SendPasswordResetEmailFunc(ctx, to, token)
	}
	return nil
}

func (m *EmailService) SendVerificationEmail(ctx context.Context, to, token string) error {
	if m.SendVerificationEmailFunc != nil {
		return m.SendVerificationEmailFunc(ctx, to, token)
	}
	return nil
}

func (m *EmailService) SendSecurityAlertEmail(ctx context.Context, to, userID string) error {
	if m.SendSecurityAlertEmailFunc != nil {
		return m.SendSecurityAlertEmailFunc(ctx, to, userID)
	}
	return nil
}

func (m *EmailService) SendNotificationEmail(ctx context.Context, to, subject, headline, message string) error {
	if m.SendNotificationEmailFunc != nil {
		return m.SendNotificationEmailFunc(ctx, to, subject, headline, message)
	}
	return nil
}
