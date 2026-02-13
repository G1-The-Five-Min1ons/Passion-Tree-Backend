package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/auth/repository"
	"passiontree/internal/config"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	JWTSecret     = "passion-tree-secret-key-2024" // TODO: Move to config
	JWTExpiration = 24 * 7 * time.Hour
)

type SocialAuthService interface {
	GetGoogleAuthURL(state string) string
	GetDiscordAuthURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*model.User, string, error)
	HandleDiscordCallback(ctx context.Context, code string) (*model.User, string, error)
}

type socialAuthServiceImpl struct {
	userRepo       repository.UserRepository
	socialRepo     repository.SocialAuthRepository
	googleConfig   *oauth2.Config
	discordConfig  *oauth2.Config
	logger         *slog.Logger
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

// GetGoogleAuthURL generates the Google OAuth2 authorization URL
func (s *socialAuthServiceImpl) GetGoogleAuthURL(state string) string {
	return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetDiscordAuthURL generates the Discord OAuth2 authorization URL
func (s *socialAuthServiceImpl) GetDiscordAuthURL(state string) string {
	return s.discordConfig.AuthCodeURL(state)
}

// HandleGoogleCallback processes the Google OAuth2 callback
func (s *socialAuthServiceImpl) HandleGoogleCallback(ctx context.Context, code string) (*model.User, string, error) {
	// Exchange code for token
	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		s.logger.Error("failed to exchange google code", "error", err)
		return nil, "", apperror.InternalServerError("Failed to authenticate with Google", err)
	}

	// Fetch user info from Google
	userInfo, err := s.fetchGoogleUserInfo(ctx, token)
	if err != nil {
		return nil, "", err
	}

	// Find or create user
	user, jwtToken, err := s.findOrCreateUser(ctx, userInfo)
	if err != nil {
		return nil, "", err
	}

	return user, jwtToken, nil
}

// HandleDiscordCallback processes the Discord OAuth2 callback
func (s *socialAuthServiceImpl) HandleDiscordCallback(ctx context.Context, code string) (*model.User, string, error) {
	// Exchange code for token
	token, err := s.discordConfig.Exchange(ctx, code)
	if err != nil {
		s.logger.Error("failed to exchange discord code", "error", err)
		return nil, "", apperror.InternalServerError("Failed to authenticate with Discord", err)
	}

	// Fetch user info from Discord
	userInfo, err := s.fetchDiscordUserInfo(ctx, token)
	if err != nil {
		return nil, "", err
	}

	// Find or create user
	user, jwtToken, err := s.findOrCreateUser(ctx, userInfo)
	if err != nil {
		return nil, "", err
	}

	return user, jwtToken, nil
}

// fetchGoogleUserInfo retrieves user information from Google API
func (s *socialAuthServiceImpl) fetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*model.OAuthUserInfo, error) {
	client := s.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		s.logger.Error("failed to get google user info", "error", err)
		return nil, apperror.InternalServerError("Failed to fetch user info from Google", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed to read google response", "error", err)
		return nil, apperror.InternalServerError("Failed to read Google response", err)
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		s.logger.Error("failed to unmarshal google user", "error", err)
		return nil, apperror.InternalServerError("Failed to parse Google user data", err)
	}

	return &model.OAuthUserInfo{
		ProviderUserID: googleUser.ID,
		Email:          googleUser.Email,
		FirstName:      googleUser.GivenName,
		LastName:       googleUser.FamilyName,
		AvatarURL:      googleUser.Picture,
		Provider:       model.AuthProviderGoogle,
	}, nil
}

// fetchDiscordUserInfo retrieves user information from Discord API
func (s *socialAuthServiceImpl) fetchDiscordUserInfo(ctx context.Context, token *oauth2.Token) (*model.OAuthUserInfo, error) {
	client := s.discordConfig.Client(ctx, token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		s.logger.Error("failed to get discord user info", "error", err)
		return nil, apperror.InternalServerError("Failed to fetch user info from Discord", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed to read discord response", "error", err)
		return nil, apperror.InternalServerError("Failed to read Discord response", err)
	}

	var discordUser struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Discriminator string `json:"discriminator"`
		Email         string `json:"email"`
		Verified      bool   `json:"verified"`
		Avatar        string `json:"avatar"`
		GlobalName    string `json:"global_name"`
	}

	if err := json.Unmarshal(body, &discordUser); err != nil {
		s.logger.Error("failed to unmarshal discord user", "error", err)
		return nil, apperror.InternalServerError("Failed to parse Discord user data", err)
	}

	// Build avatar URL
	avatarURL := ""
	if discordUser.Avatar != "" {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUser.ID, discordUser.Avatar)
	}

	// Parse name
	firstName := discordUser.GlobalName
	lastName := ""
	if firstName == "" {
		firstName = discordUser.Username
	}
	nameParts := strings.Fields(discordUser.GlobalName)
	if len(nameParts) > 1 {
		firstName = nameParts[0]
		lastName = strings.Join(nameParts[1:], " ")
	}

	return &model.OAuthUserInfo{
		ProviderUserID: discordUser.ID,
		Email:          discordUser.Email,
		FirstName:      firstName,
		LastName:       lastName,
		AvatarURL:      avatarURL,
		Provider:       model.AuthProviderDiscord,
	}, nil
}

// findOrCreateUser finds an existing user or creates a new one from OAuth info
func (s *socialAuthServiceImpl) findOrCreateUser(ctx context.Context, userInfo *model.OAuthUserInfo) (*model.User, string, error) {
	// Check if user exists with this provider
	user, err := s.socialRepo.GetUserByProvider(ctx, userInfo.Provider, userInfo.ProviderUserID)
	if err != nil {
		return nil, "", apperror.InternalServerError("Failed to check user existence", err)
	}

	// If user doesn't exist, check by email
	if user == nil {
		existingUser, err := s.userRepo.GetUserByEmail(ctx, userInfo.Email)
		if err != nil {
			return nil, "", apperror.InternalServerError("Failed to check user by email", err)
		}

		if existingUser != nil {
			// User exists with same email but different provider
			// Link the social account to existing user
			err = s.socialRepo.LinkSocialAccount(ctx, existingUser.UserID, userInfo.Provider, userInfo.ProviderUserID)
			if err != nil {
				return nil, "", apperror.InternalServerError("Failed to link social account", err)
			}
			user = existingUser
			user.AuthProvider = userInfo.Provider
			user.ProviderUserID = userInfo.ProviderUserID
		} else {
			// Create new user
			user, err = s.createUserFromOAuth(ctx, userInfo)
			if err != nil {
				return nil, "", err
			}
		}
	}

	// Generate JWT token
	jwtToken, err := s.generateJWT(user)
	if err != nil {
		return nil, "", apperror.InternalServerError("Failed to generate token", err)
	}

	return user, jwtToken, nil
}

// createUserFromOAuth creates a new user from OAuth provider information
func (s *socialAuthServiceImpl) createUserFromOAuth(ctx context.Context, userInfo *model.OAuthUserInfo) (*model.User, error) {
	// Generate username from email
	username := strings.Split(userInfo.Email, "@")[0]
	
	// Check if username exists, append random number if needed
	existingUser, _ := s.userRepo.GetUserByUsername(ctx, username)
	if existingUser != nil {
		username = fmt.Sprintf("%s_%d", username, time.Now().Unix()%10000)
	}

	user := &model.User{
		Username:        username,
		Email:           userInfo.Email,
		FirstName:       userInfo.FirstName,
		LastName:        userInfo.LastName,
		Role:            model.RoleStudent,
		HeartCount:      0,
		IsEmailVerified: true,
		AuthProvider:    userInfo.Provider,
		ProviderUserID:  userInfo.ProviderUserID,
	}

	profile := &model.Profile{
		AvatarURL: userInfo.AvatarURL,
		Bio:       "",
		Location:  "",
	}

	userID, err := s.socialRepo.CreateSocialUser(ctx, user, profile)
	if err != nil {
		s.logger.Error("failed to create social user", "error", err)
		return nil, apperror.InternalServerError("Failed to create user account", err)
	}

	user.UserID = userID
	return user, nil
}

// generateJWT creates a JWT token for authenticated user
func (s *socialAuthServiceImpl) generateJWT(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.UserID,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(JWTExpiration).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
