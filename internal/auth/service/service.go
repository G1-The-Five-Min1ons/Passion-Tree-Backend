package service

import (
	"context"
	"log/slog"

	"passiontree/internal/auth/model"
	"passiontree/internal/auth/repository"
	"passiontree/internal/config"
)

type UserService interface {
	CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, firstName string, lastName string) error
	UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error
	DeleteUser(ctx context.Context, id string, password string) error
	Login(ctx context.Context, identifier string, password string, deviceInfo, ipAddress, userAgent string) (accessToken, refreshToken string, err error)
	RefreshAccessToken(ctx context.Context, refreshToken string, deviceInfo, ipAddress, userAgent string) (newAccessToken, newRefreshToken string, err error)
	Logout(ctx context.Context, userID string) error
	ValidateToken(ctx context.Context, token string) (*model.User, error)
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, email string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, code string, newPassword string) error
	ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error

	// Multi-device Session Management
	GetActiveSessions(ctx context.Context, userID string, currentRefreshToken string) (*model.GetActiveSessionsResponse, error)
	LogoutSession(ctx context.Context, userID string, sessionID string) error
}

type EmailService interface {
	SendVerificationEmail(to, token string) error
	SendPasswordResetEmail(to, token string) error
	SendSecurityAlertEmail(to, userID string) error
}

type userServiceImpl struct {
	repo         repository.Repository
	emailService EmailService
	config       *config.Config
	logger       *slog.Logger
}

type emailServiceImpl struct {
	config *config.Config
	logger *slog.Logger
}

func NewUserService(repo repository.Repository, cfg *config.Config, logger *slog.Logger) UserService {
	return &userServiceImpl{
		repo:   repo,
		config: cfg,
		logger: logger,
	}
}

func NewEmailService(cfg *config.Config, logger *slog.Logger) EmailService {
	return &emailServiceImpl{
		config: cfg,
		logger: logger,
	}
}

// NewUserServiceWithEmail creates a new UserService with email service configured
func NewUserServiceWithEmail(repo repository.Repository, cfg *config.Config, logger *slog.Logger) UserService {
	svc := &userServiceImpl{
		repo:   repo,
		config: cfg,
		logger: logger,
	}

	// Initialize email service if SMTP or MailerSend is configured
	if cfg.SMTPHost != "" {
		svc.emailService = NewEmailService(cfg, logger)
		svc.logger.Info("Email service initialized (SMTP)")
	} else if cfg.MailerSendAPIKey != "" {
		svc.emailService = NewEmailService(cfg, logger)
		svc.logger.Info("Email service initialized (MailerSend API)")
	} else {
		svc.logger.Warn("Email service NOT initialized - no email configuration found")
	}

	return svc
}

// SetEmailService sets the email service (used for dependency injection)
func (s *userServiceImpl) SetEmailService(emailService EmailService) {
	s.emailService = emailService
}
