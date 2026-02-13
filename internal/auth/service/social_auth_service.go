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
	HandleGoogleCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error)
	HandleDiscordCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error)
	// Native SSO method for Android/mobile apps
	HandleNativeGoogleSignIn(ctx context.Context, idToken string) (*model.User, string, error)
	// Confirm account linking
	ConfirmAccountLink(ctx context.Context, linkToken string, confirm bool) (*model.User, string, error)
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
func (s *socialAuthServiceImpl) HandleGoogleCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	// Exchange code for token
	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		s.logger.Error("failed to exchange google code", "error", err)
		return nil, "", nil, apperror.InternalServerError("Failed to authenticate with Google", err)
	}

	// Fetch user info from Google
	userInfo, err := s.fetchGoogleUserInfo(ctx, token)
	if err != nil {
		return nil, "", nil, err
	}

	// Find or create user
	user, jwtToken, linkConfirm, err := s.findOrCreateUser(ctx, userInfo)
	if err != nil {
		return nil, "", nil, err
	}

	return user, jwtToken, linkConfirm, nil
}

// HandleDiscordCallback processes the Discord OAuth2 callback
func (s *socialAuthServiceImpl) HandleDiscordCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	// Exchange code for token
	token, err := s.discordConfig.Exchange(ctx, code)
	if err != nil {
		s.logger.Error("failed to exchange discord code", "error", err)
		return nil, "", nil, apperror.InternalServerError("Failed to authenticate with Discord", err)
	}

	// Fetch user info from Discord
	userInfo, err := s.fetchDiscordUserInfo(ctx, token)
	if err != nil {
		return nil, "", nil, err
	}

	// Find or create user
	user, jwtToken, linkConfirm, err := s.findOrCreateUser(ctx, userInfo)
	if err != nil {
		return nil, "", nil, err
	}

	return user, jwtToken, linkConfirm, nil
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
// This function implements UPSERT logic:
// - If user exists with provider: Update their info from provider
// - If user exists with email only: Ask user to confirm linking
// - If user doesn't exist: Create new user
func (s *socialAuthServiceImpl) findOrCreateUser(ctx context.Context, userInfo *model.OAuthUserInfo) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	// Check if user exists with this provider
	user, err := s.socialRepo.GetUserByProvider(ctx, userInfo.Provider, userInfo.ProviderUserID)
	if err != nil {
		return nil, "", nil, apperror.InternalServerError("Failed to check user existence", err)
	}

	// CASE 1: User exists with this social provider
	if user != nil {
		s.logger.Info("user exists with provider, updating info",
			"user_id", user.UserID,
			"provider", userInfo.Provider,
		)

		// Update user info from provider (in case they changed name/email)
		err = s.socialRepo.UpdateSocialUserInfo(ctx, user.UserID, userInfo)
		if err != nil {
			s.logger.Error("failed to update user info", "error", err)
			// Continue anyway - update is not critical
		}

		// Update profile (avatar)
		profile := &model.Profile{
			AvatarURL: userInfo.AvatarURL,
		}
		err = s.socialRepo.UpsertSocialUserProfile(ctx, user.UserID, profile)
		if err != nil {
			s.logger.Error("failed to update profile", "error", err)
			// Continue anyway - profile update is not critical
		}

		// Refresh user data
		user.FirstName = userInfo.FirstName
		user.LastName = userInfo.LastName
		user.Email = userInfo.Email

		// Generate JWT token
		jwtToken, err := s.generateJWT(user)
		if err != nil {
			return nil, "", nil, apperror.InternalServerError("Failed to generate token", err)
		}

		return user, jwtToken, nil, nil
	}

	// CASE 2: User doesn't have this provider, check by email
	existingUser, err := s.userRepo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, "", nil, apperror.InternalServerError("Failed to check user by email", err)
	}

	if existingUser != nil {
		// CASE 2A: User exists with same email but different provider (or local account)
		// NEED TO ASK USER TO CONFIRM LINKING
		s.logger.Info("user exists with email, need confirmation to link",
			"user_id", existingUser.UserID,
			"email", userInfo.Email,
			"existing_provider", existingUser.AuthProvider,
			"new_provider", userInfo.Provider,
		)

		// Create link token with provider info
		linkToken, err := s.generateLinkToken(existingUser.UserID, userInfo)
		if err != nil {
			return nil, "", nil, apperror.InternalServerError("Failed to generate link token", err)
		}

		// Return confirmation needed
		linkConfirm := &model.LinkConfirmationNeeded{
			LinkToken: linkToken,
			ExistingUser: &model.User{
				UserID:    existingUser.UserID,
				Username:  existingUser.Username,
				Email:     existingUser.Email,
				FirstName: existingUser.FirstName,
				LastName:  existingUser.LastName,
				AuthProvider: existingUser.AuthProvider,
			},
			ProviderEmail: userInfo.Email,
			ProviderName:  userInfo.Provider,
			Message: fmt.Sprintf("An account with email %s already exists. Do you want to link your %s account?", userInfo.Email, userInfo.Provider),
			NeedsConfirm: true,
		}

		return nil, "", linkConfirm, nil
	}

	// CASE 3: User doesn't exist at all - Create new user
	s.logger.Info("creating new user from social auth",
		"email", userInfo.Email,
		"provider", userInfo.Provider,
	)

	user, err = s.createUserFromOAuth(ctx, userInfo)
	if err != nil {
		return nil, "", nil, err
	}

	// Generate JWT token
	jwtToken, err := s.generateJWT(user)
	if err != nil {
		return nil, "", nil, apperror.InternalServerError("Failed to generate token", err)
	}

	return user, jwtToken, nil, nil
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

// generateLinkToken creates a temporary token for account linking confirmation
func (s *socialAuthServiceImpl) generateLinkToken(userID string, providerInfo *model.OAuthUserInfo) (string, error) {
	claims := jwt.MapClaims{
		"user_id":          userID,
		"provider":         providerInfo.Provider,
		"provider_user_id": providerInfo.ProviderUserID,
		"provider_email":   providerInfo.Email,
		"first_name":       providerInfo.FirstName,
		"last_name":        providerInfo.LastName,
		"avatar_url":       providerInfo.AvatarURL,
		"type":             "link_confirmation",
		"exp":              time.Now().Add(10 * time.Minute).Unix(), // Short expiration
		"iat":              time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ConfirmAccountLink handles user's decision to link or not link accounts
func (s *socialAuthServiceImpl) ConfirmAccountLink(ctx context.Context, linkToken string, confirm bool) (*model.User, string, error) {
	// Parse and validate link token
	token, err := jwt.Parse(linkToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		s.logger.Error("failed to parse link token", "error", err)
		return nil, "", apperror.NewAppError(fiber.StatusUnauthorized, "Invalid or expired link token", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, "", apperror.NewAppError(fiber.StatusUnauthorized, "Invalid link token", nil)
	}

	// Verify token type
	if claims["type"] != "link_confirmation" {
		return nil, "", apperror.NewAppError(fiber.StatusUnauthorized, "Invalid token type", nil)
	}

	// Extract claims
	userID := claims["user_id"].(string)
	provider := claims["provider"].(string)
	providerUserID := claims["provider_user_id"].(string)

	// Get existing user
	existingUser, _, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, "", apperror.InternalServerError("Failed to get user", err)
	}

	if existingUser == nil {
		return nil, "", apperror.NewAppError(fiber.StatusNotFound, "User not found", nil)
	}

	if confirm {
		// User confirmed - Link the accounts
		s.logger.Info("user confirmed account linking",
			"user_id", userID,
			"provider", provider,
		)

		// Link the social account
		err = s.socialRepo.LinkSocialAccount(ctx, userID, provider, providerUserID)
		if err != nil {
			return nil, "", apperror.InternalServerError("Failed to link account", err)
		}

		// Update user info from provider
		providerInfo := &model.OAuthUserInfo{
			Provider:       provider,
			ProviderUserID: providerUserID,
			Email:          claims["provider_email"].(string),
			FirstName:      getStringClaim(claims, "first_name"),
			LastName:       getStringClaim(claims, "last_name"),
			AvatarURL:      getStringClaim(claims, "avatar_url"),
		}

		err = s.socialRepo.UpdateSocialUserInfo(ctx, userID, providerInfo)
		if err != nil {
			s.logger.Error("failed to update user info after linking", "error", err)
		}

		// Update profile (avatar)
		if providerInfo.AvatarURL != "" {
			profile := &model.Profile{
				AvatarURL: providerInfo.AvatarURL,
			}
			err = s.socialRepo.UpsertSocialUserProfile(ctx, userID, profile)
			if err != nil {
				s.logger.Error("failed to update profile after linking", "error", err)
			}
		}

		// Update user object
		existingUser.AuthProvider = provider
		existingUser.ProviderUserID = providerUserID

		s.logger.Info("account linked successfully",
			"user_id", userID,
			"provider", provider,
		)
	} else {
		// User declined - Just login with existing account
		s.logger.Info("user declined account linking, using existing account",
			"user_id", userID,
			"provider", provider,
		)
	}

	// Generate JWT token for login
	jwtToken, err := s.generateJWT(existingUser)
	if err != nil {
		return nil, "", apperror.InternalServerError("Failed to generate token", err)
	}

	return existingUser, jwtToken, nil
}

// Helper function to safely get string from JWT claims
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

// HandleNativeGoogleSignIn processes native Google Sign-In from mobile apps
// This method verifies the Google ID token and authenticates the user
func (s *socialAuthServiceImpl) HandleNativeGoogleSignIn(ctx context.Context, idToken string) (*model.User, string, error) {
	// Verify the ID token with Google
	userInfo, err := s.verifyGoogleIDToken(ctx, idToken)
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

// verifyGoogleIDToken verifies Google ID token from native apps
func (s *socialAuthServiceImpl) verifyGoogleIDToken(ctx context.Context, idToken string) (*model.OAuthUserInfo, error) {
	// Call Google's tokeninfo endpoint to verify the token
	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken))
	if err != nil {
		s.logger.Error("failed to verify google id token", "error", err)
		return nil, apperror.InternalServerError("Failed to verify Google ID token", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("invalid google id token", "status", resp.StatusCode)
		return nil, apperror.NewAppError(fiber.StatusUnauthorized, "Invalid or expired Google ID token", nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed to read token verification response", "error", err)
		return nil, apperror.InternalServerError("Failed to read verification response", err)
	}

	var tokenInfo struct {
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
	}

	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		s.logger.Error("failed to parse token info", "error", err)
		return nil, apperror.InternalServerError("Failed to parse token information", err)
	}

	// Verify the audience matches our client ID
	if tokenInfo.Aud != s.googleConfig.ClientID {
		s.logger.Warn("token audience mismatch",
			"expected", s.googleConfig.ClientID,
			"got", tokenInfo.Aud,
		)
		return nil, apperror.NewAppError(fiber.StatusUnauthorized, "Invalid token audience", nil)
	}

	return &model.OAuthUserInfo{
		ProviderUserID: tokenInfo.Sub,
		Email:          tokenInfo.Email,
		FirstName:      tokenInfo.GivenName,
		LastName:       tokenInfo.FamilyName,
		AvatarURL:      tokenInfo.Picture,
		Provider:       model.AuthProviderGoogle,
	}, nil
}
