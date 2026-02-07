package service

import (
	"context"
	"log"

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
	Login(ctx context.Context, identifier string, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (*model.User, error)
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, email string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, code string, newPassword string) error
	ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error
}

type userServiceImpl struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.TokenRepository
	emailService EmailService
}

func NewUserService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository) UserService {
	return &userServiceImpl{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

// NewUserServiceWithEmail creates a new UserService with email service configured
func NewUserServiceWithEmail(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, cfg *config.Config) UserService {
	svc := &userServiceImpl{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}

	// Initialize email service if SMTP or MailerSend is configured
	if cfg.SMTPHost != "" {
		svc.emailService = NewEmailService(cfg)
		log.Println("Email service initialized (SMTP)")
	} else if cfg.MailerSendAPIKey != "" {
		svc.emailService = NewEmailService(cfg)
		log.Println("Email service initialized (MailerSend API)")
	} else {
		log.Println("Warning: Email service NOT initialized - no email configuration found")
	}

	return svc
}

// SetEmailService sets the email service (used for dependency injection)
func (s *userServiceImpl) SetEmailService(emailService EmailService) {
	s.emailService = emailService
}
