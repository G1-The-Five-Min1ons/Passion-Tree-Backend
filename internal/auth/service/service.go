package service

import (
	"context"
	"log/slog"

	"passiontree/internal/auth/model"
	"passiontree/internal/auth/repository"
	"passiontree/internal/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserService interface {
	CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, firstName string, lastName string) error
	UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error
	DeleteUser(ctx context.Context, id string, password string) error
	Login(ctx context.Context, identifier string, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (*model.User, error)
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, email string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, code string, newPassword string) error
	ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error
}

type EmailService interface {
	SendVerificationEmail(to, token string) error
	SendPasswordResetEmail(to, token string) error
}

type SocialAuthService interface {
	GetGoogleAuthURL(state string) string
	GetDiscordAuthURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error)
	HandleDiscordCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error)
	// Native SSO method for Android/mobile apps
	HandleNativeGoogleSignIn(ctx context.Context, idToken string) (*model.User, string, error)
	// Confirm account linking
	ConfirmAccountLink(ctx context.Context, linkToken string, confirm bool) (*model.User, string, error)
}

type userServiceImpl struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.TokenRepository
	emailService EmailService
	logger       *slog.Logger
}

type emailServiceImpl struct {
	config *config.Config
	logger *slog.Logger
}

type socialAuthServiceImpl struct {
	userRepo      repository.UserRepository
	socialRepo    repository.SocialAuthRepository
	googleConfig  *oauth2.Config
	discordConfig *oauth2.Config
	logger        *slog.Logger
}

func NewUserService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, logger *slog.Logger) UserService {
	return &userServiceImpl{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		logger:    logger,
	}
}

func NewEmailService(cfg *config.Config, logger *slog.Logger) EmailService {
	return &emailServiceImpl{
		config: cfg,
		logger: logger,
	}
}

// NewUserServiceWithEmail creates a new UserService with email service configured
func NewUserServiceWithEmail(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, cfg *config.Config, logger *slog.Logger) UserService {
	svc := &userServiceImpl{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		logger:    logger,
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

func NewSocialAuthService(
	userRepo repository.UserRepository,
	socialRepo repository.SocialAuthRepository,
	cfg *config.Config,
	logger *slog.Logger,
) SocialAuthService {
	return &socialAuthServiceImpl{
		userRepo:   userRepo,
		socialRepo: socialRepo,
		googleConfig: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		discordConfig: &oauth2.Config{
			ClientID:     cfg.DiscordClientID,
			ClientSecret: cfg.DiscordClientSecret,
			RedirectURL:  cfg.DiscordRedirectURL,
			Scopes:       []string{"identify", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/api/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
		},
		logger: logger,
	}
}
