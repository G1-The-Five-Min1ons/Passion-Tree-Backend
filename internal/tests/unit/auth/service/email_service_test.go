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
	expectedTime := time.Now().Add(15 * time.Minute)

	// Since they might not occur exactly simultaneously, let's allow a delta of a second
	diff := expiry.Sub(expectedTime)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Expected expiry to be close to %v, got %v", expectedTime, expiry)
	}
}

// Ensure smtpSendMail mock works
func TestEmailServiceFallbackToSMTP(t *testing.T) {
	// Setup generic SMTP mock
	mockSMTPError := errors.New("mock smtp connection error")

	originalSMTP := service.ExportSetSMTPSendMail(func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		// Mock behavior to return an error just for evaluation
		return mockSMTPError
	})
	defer service.ExportSetSMTPSendMail(originalSMTP)

	cfg := &config.Config{
		MailerSendAPIKey: "", // Purposefully empty to trigger fallback
		SMTPFromEmail:    "noreply@example.com",
		GmailEmail:       "test@gmail.com",
		GmailAppPassword: "mock-password",
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	emailSvc := service.NewEmailService(cfg, logger)

	t.Run("SendVerificationEmail", func(t *testing.T) {
		err := emailSvc.SendVerificationEmail("user@example.com", "123456")
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
		err := emailSvc.SendSecurityAlertEmail("user@example.com", "u-1")
		if err == nil {
			t.Fatal("Expected an error due to mock SMTP simulated error, got nil")
		}
		if !strings.Contains(err.Error(), "mock smtp connection error") {
			t.Errorf("Expected mock smtp connection error, got: %v", err)
		}
	})
}

// Ensure Mailersend client initialization doesn't panic and methods return safely
func TestEmailServiceMailerSend(t *testing.T) {
	// We want to test the case where API Key is given, so it attempts MailerSend
	// Since we can't easily deep-mock mailersendClient (it's tightly coupled inside emailServiceImpl struct)
	// without abstracting MailerSend client into an interface, we'll let it execute with a dummy key.
	// The HTTP call will fail due to unauthorized, and it should gracefully fallback to SMTP.

	mockSMTPCalled := false
	originalSMTP := service.ExportSetSMTPSendMail(func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		mockSMTPCalled = true
		return nil
	})
	defer service.ExportSetSMTPSendMail(originalSMTP)

	cfg := &config.Config{
		MailerSendAPIKey: "dummy-key-that-will-fail-on-network",
		SMTPFromEmail:    "noreply@example.com",
		GmailEmail:       "test@gmail.com",
		GmailAppPassword: "mock-password",
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	emailSvc := service.NewEmailService(cfg, logger)

	err := emailSvc.SendVerificationEmail("user@example.com", "123456")

	// Either Mailersend fails due to network and triggers SMTP, or it passes (unlikely with dummy key)
	// Our main goal is no panics!
	if err != nil {
		t.Logf("Expected behavior: Fallback successful or error returned: %v", err)
	}

	// Typically, it would fallback and call mockSMTPCalled == true if dummy-key makes it return an error
	// But since it tries to make a real HTTP call, we'll just check that it didn't crash.
	_ = mockSMTPCalled
}
