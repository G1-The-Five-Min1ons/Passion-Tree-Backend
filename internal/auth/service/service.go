package service

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/auth/repository"
	"passiontree/internal/config"
	"passiontree/internal/pkg/jwt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	//go:embed templates/verification.html
	verificationTemplate string
	//go:embed templates/password_reset.html
	passwordResetTemplate string
	//go:embed templates/security_alert.html
	securityAlertTemplate string
	//go:embed templates/notification.html
	notificationTemplate string

	fixedTestOTPPattern = regexp.MustCompile(`^\d{6}$`)
)

// resolveVerificationOTP returns a fixed OTP in test mode when configured,
// otherwise it generates a normal random verification code.
func resolveVerificationOTP(logger *slog.Logger, flow string) (string, error) {
	fixedOTP := strings.TrimSpace(os.Getenv("FIXED_TEST_OTP_CODE"))
	if fixedOTP == "" {
		return GenerateVerificationToken()
	}

	if !fixedTestOTPPattern.MatchString(fixedOTP) {
		return "", fmt.Errorf("FIXED_TEST_OTP_CODE must be a 6-digit numeric code")
	}

	if logger != nil {
		logger.Info("[DEBUG] Fixed OTP configured for test flow",
			"flow", flow,
			"otp_code", fixedOTP,
		)
	}

	return fixedOTP, nil
}

type UserService interface {
	CreateUser(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	CreateUserByAdmin(ctx context.Context, user *model.User, profile *model.Profile) (string, error)
	GetUserByID(ctx context.Context, id string) (*model.User, *model.Profile, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetAllUsers(ctx context.Context) ([]*model.UserWithProfile, error)
	GetDashboardStats(ctx context.Context) (*model.DashboardStats, error)
	UpdateUser(ctx context.Context, id string, username string, firstName string, lastName string, role string) error
	UpdateUserByAdmin(ctx context.Context, id string, firstName string, lastName string, role string) error
	UpdateProfile(ctx context.Context, userID string, profile *model.Profile) error
	DeleteUser(ctx context.Context, id string, password string) error
	DeleteUserByAdmin(ctx context.Context, id string) error
	Login(ctx context.Context, identifier string, password string, deviceInfo, ipAddress, userAgent string) (accessToken, refreshToken string, err error)
	RefreshAccessToken(ctx context.Context, refreshToken string, deviceInfo, ipAddress, userAgent string) (newAccessToken, newRefreshToken string, err error)
	Logout(ctx context.Context, userID string) error
	ValidateToken(ctx context.Context, token string) (*model.User, error)
	VerifyEmail(ctx context.Context, vToken string, deviceInfo, ip, ua string) (accessToken string, refreshToken string, err error)
	ResendVerificationEmail(ctx context.Context, email string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, code string, newPassword string) error
	ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error
	DeactivateAccount(ctx context.Context, userID string, days int) error
	ReactivateAccount(ctx context.Context, userID string) error
	GetAccountDeactivatedUntil(ctx context.Context, userID string) (*time.Time, error)

	// Multi-device Session Management
	GetActiveSessions(ctx context.Context, userID string, currentRefreshToken string) (*model.GetActiveSessionsResponse, error)
	LogoutSession(ctx context.Context, userID string, sessionID string) error
	GetTeacherVerificationStatus(ctx context.Context, userID string) (*model.TeacherVerificationStatus, error)
	ApplyForTeacher(ctx context.Context, userID string, req model.ApplyTeacherRequest) error
	GetTeacherApplications(ctx context.Context, status string) ([]model.TeacherVerificationRequest, error)
	ReviewTeacherApplication(ctx context.Context, requestID, reviewedBy string, req model.ReviewTeacherApplicationRequest) error
}

type EmailService interface {
	SendVerificationEmail(ctx context.Context, to, token string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
	SendSecurityAlertEmail(ctx context.Context, to, userID string) error
	SendNotificationEmail(ctx context.Context, to, subject, headline, message string) error
}

type SocialAuthService interface {
	GetGoogleAuthURL(state string) string
	GetDiscordAuthURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error)
	HandleDiscordCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error)
	HandleNativeGoogleSignIn(ctx context.Context, idToken string) (*model.User, string, error)
	HandleNativeDiscordSignIn(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error)
	ConfirmAccountLink(ctx context.Context, linkToken string, confirm bool) (*model.User, string, error)
}

type userServiceImpl struct {
	repo          repository.Repository
	emailService  EmailService
	jwtService    *jwt.Service
	config        *config.Config
	logger        *slog.Logger
	googleConfig  *oauth2.Config
	discordConfig *oauth2.Config
}

type fetchUserInfoFunc func(context.Context, *oauth2.Token) (*model.OAuthUserInfo, error)

// emailTemplates holds parsed HTML templates for email notifications
type emailTemplates struct {
	verification  *template.Template
	passwordReset *template.Template
	securityAlert *template.Template
	notification  *template.Template
}
type emailServiceImpl struct {
	templates        *emailTemplates
	config           *config.Config
	logger           *slog.Logger
}

// --- Constructors ---

func NewUserService(repo repository.Repository, emailSvc EmailService, cfg *config.Config, jwtSvc *jwt.Service, logger *slog.Logger) UserService {
	return &userServiceImpl{
		repo:         repo,
		config:       cfg,
		emailService: emailSvc,
		jwtService:   jwtSvc,
		logger:       logger,
	}
}

func NewEmailService(cfg *config.Config, logger *slog.Logger) EmailService {
	verificationTmpl := template.Must(template.New("verification").Parse(verificationTemplate))
	passwordResetTmpl := template.Must(template.New("passwordReset").Parse(passwordResetTemplate))
	securityAlertTmpl := template.Must(template.New("securityAlert").Parse(securityAlertTemplate))
	notificationTmpl := template.Must(template.New("notification").Parse(notificationTemplate))

	return &emailServiceImpl{
		templates: &emailTemplates{
			verification:  verificationTmpl,
			passwordReset: passwordResetTmpl,
			securityAlert: securityAlertTmpl,
			notification:  notificationTmpl,
		},
		config: cfg,
		logger: logger,
	}
}

// NewSocialAuthService จะคืนค่าเป็น userServiceImpl ที่มีการตั้งค่า OAuth
func NewSocialAuthService(repo repository.Repository, cfg *config.Config, jwtSvc *jwt.Service, logger *slog.Logger) SocialAuthService {
	return &userServiceImpl{
		repo:       repo,
		config:     cfg,
		jwtService: jwtSvc,
		logger:     logger,
		googleConfig: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
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
	}
}
