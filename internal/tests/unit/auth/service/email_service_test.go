package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"passiontree/internal/auth/service"
	"passiontree/internal/config"
)

func TestGenerateVerificationToken(t *testing.T) {
	token, err := service.GenerateVerificationToken()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(token) != 6 {
		t.Errorf("Expected token length to be 6, got %d", len(token))
	}

	// Ensure token is numeric
	for _, char := range token {
		if char < '0' || char > '9' {
			t.Errorf("Expected token to be numeric, got %c in %s", char, token)
		}
	}
}

func TestGetVerificationTokenExpiry(t *testing.T) {
	expiry := service.GetVerificationTokenExpiry()
	expectedTime := time.Now().Add(5 * time.Minute)

	// Since they might not occur exactly simultaneously, let's allow a delta of a second
	diff := expiry.Sub(expectedTime)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Expected expiry to be close to %v, got %v", expectedTime, expiry)
	}
}

// Ensure smtpSendMail mock works
func TestEmailServiceSMTP(t *testing.T) {
	// Setup generic SMTP mock
	mockSMTPError := errors.New("mock smtp connection error")

	originalSMTP := service.ExportSetSMTPSendMail(func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		// Mock behavior to return an error just for evaluation
		return mockSMTPError
	})
	defer service.ExportSetSMTPSendMail(originalSMTP)

	cfg := &config.Config{
		SMTPFromEmail:    "noreply@example.com",
		GmailEmail:       "test@gmail.com",
		GmailAppPassword: "mock-password",
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	emailSvc := service.NewEmailService(cfg, logger)

	t.Run("SendVerificationEmail", func(t *testing.T) {
		err := emailSvc.SendVerificationEmail(context.Background(), "user@example.com", "123456")
		if err == nil {
			t.Fatal("Expected an error due to mock SMTP simulated error, got nil")
		}
		if !strings.Contains(err.Error(), "mock smtp connection error") {
			t.Errorf("Expected mock smtp connection error, got: %v", err)
		}
	})

	t.Run("SendPasswordResetEmail", func(t *testing.T) {
		ctx := context.Background()
		err := emailSvc.SendPasswordResetEmail(ctx, "user@example.com", "reset-token")
		if err == nil {
			t.Fatal("Expected an error due to mock SMTP simulated error, got nil")
		}
		if !strings.Contains(err.Error(), "mock smtp connection error") {
			t.Errorf("Expected mock smtp connection error, got: %v", err)
		}
	})

	t.Run("SendSecurityAlertEmail", func(t *testing.T) {
		err := emailSvc.SendSecurityAlertEmail(context.Background(), "user@example.com", "u-1")
		if err == nil {
			t.Fatal("Expected an error due to mock SMTP simulated error, got nil")
		}
		if !strings.Contains(err.Error(), "mock smtp connection error") {
			t.Errorf("Expected mock smtp connection error, got: %v", err)
		}
	})
}
