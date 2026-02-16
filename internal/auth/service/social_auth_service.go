package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// GetGoogleAuthURL generates the Google OAuth2 authorization URL
func (s *userServiceImpl) GetGoogleAuthURL(state string) string {
	return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetDiscordAuthURL generates the Discord OAuth2 authorization URL
func (s *userServiceImpl) GetDiscordAuthURL(state string) string {
	return s.discordConfig.AuthCodeURL(state)
}

// handleOAuthCallback is a generic handler for OAuth callbacks
func (s *userServiceImpl) handleOAuthCallback(
	ctx context.Context,
	code string,
	config *oauth2.Config,
	fetchUserInfo func(context.Context, *oauth2.Token) (*model.OAuthUserInfo, error),
	providerName string,
) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	// Exchange code for token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to exchange oauth code", "provider", providerName, "error", err)
		return nil, "", nil, apperror.NewInternal("failed to authenticate with %s: %w", providerName, err)
	}

	// Fetch user info from provider
	userInfo, err := fetchUserInfo(ctx, token)
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

// HandleGoogleCallback processes the Google OAuth2 callback
func (s *userServiceImpl) HandleGoogleCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	return s.handleOAuthCallback(ctx, code, s.googleConfig, s.fetchGoogleUserInfo, "google")
}

// HandleDiscordCallback processes the Discord OAuth2 callback
func (s *userServiceImpl) HandleDiscordCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	return s.handleOAuthCallback(ctx, code, s.discordConfig, s.fetchDiscordUserInfo, "discord")
}

// fetchGoogleUserInfo retrieves user information from Google API
func (s *userServiceImpl) fetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*model.OAuthUserInfo, error) {
	client := s.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get google user info", "error", err)
		return nil, apperror.NewInternal("failed to fetch user info from google: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to read google response", "error", err)
		return nil, apperror.NewInternal("failed to read google response: %w", err)
	}

	var googleUser model.GoogleUserResponse

	if err := json.Unmarshal(body, &googleUser); err != nil {
		s.logger.ErrorContext(ctx, "failed to unmarshal google user", "error", err)
		return nil, apperror.NewInternal("failed to parse google user data: %w", err)
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
func (s *serviceImpl) fetchDiscordUserInfo(ctx context.Context, token *oauth2.Token) (*model.OAuthUserInfo, error) {
	client := s.discordConfig.Client(ctx, token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get discord user info", "error", err)
		return nil, apperror.NewInternal("failed to fetch user info from discord: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to read discord response", "error", err)
		return nil, apperror.NewInternal("failed to read discord response: %w", err)
	}

	var discordUser model.DiscordUserResponse

	if err := json.Unmarshal(body, &discordUser); err != nil {
		s.logger.ErrorContext(ctx, "failed to unmarshal discord user", "error", err)
		return nil, apperror.NewInternal("failed to parse discord user data: %w", err)
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
func (s *userServiceImpl) findOrCreateUser(ctx context.Context, userInfo *model.OAuthUserInfo) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	// Check if user exists with this provider
	user, err := s.repo.GetUserByProvider(ctx, userInfo.Provider, userInfo.ProviderUserID)
	if err != nil {
		return nil, "", nil, apperror.NewInternal("failed to check user existence: %w", err)
	}

	// CASE 1: User exists with this social provider
	if user != nil {
		s.logger.InfoContext(ctx, "user exists with provider, updating info",
			"user_id", user.UserID,
			"provider", userInfo.Provider,
		)

		// Update user info from provider (in case they changed name/email)
		err = s.repo.UpdateSocialUserInfo(ctx, user.UserID, userInfo)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to update user info", "error", err)
			// Continue anyway - update is not critical
		}

		// Update profile (avatar)
		profile := &model.Profile{
			AvatarURL: userInfo.AvatarURL,
		}
		err = s.repo.UpsertSocialUserProfile(ctx, user.UserID, profile)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to update profile", "error", err)
			// Continue anyway - profile update is not critical
		}

		// Refresh user data
		user.FirstName = userInfo.FirstName
		user.LastName = userInfo.LastName
		user.Email = userInfo.Email

		// Generate JWT token
		jwtToken, err := s.generateJWT(user)
		if err != nil {
			return nil, "", nil, apperror.NewInternal("failed to generate token: %w", err)
		}

		return user, jwtToken, nil, nil
	}

	// CASE 2: User doesn't have this provider, check by email
	existingUser, err := s.repo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, "", nil, apperror.NewInternal("failed to check user by email: %w", err)
	}

	if existingUser != nil {
		// CASE 2A: User exists with same email but different provider (or local account)
		// NEED TO ASK USER TO CONFIRM LINKING
		s.logger.InfoContext(ctx, "user exists with email, need confirmation to link",
			"user_id", existingUser.UserID,
			"email", userInfo.Email,
			"existing_provider", existingUser.AuthProvider,
			"new_provider", userInfo.Provider,
		)

		// Create link token with provider info
		linkToken, err := s.generateLinkToken(existingUser.UserID, userInfo)
		if err != nil {
			return nil, "", nil, apperror.NewInternal("failed to generate link token: %w", err)
		}

		// Return confirmation needed
		linkConfirm := &model.LinkConfirmationNeeded{
			LinkToken: linkToken,
			ExistingUser: &model.User{
				UserID:       existingUser.UserID,
				Username:     existingUser.Username,
				Email:        existingUser.Email,
				FirstName:    existingUser.FirstName,
				LastName:     existingUser.LastName,
				AuthProvider: existingUser.AuthProvider,
			},
			ProviderEmail: userInfo.Email,
			ProviderName:  userInfo.Provider,
			Message:       fmt.Sprintf("An account with email %s already exists. Do you want to link your %s account?", userInfo.Email, userInfo.Provider),
			NeedsConfirm:  true,
		}

		return nil, "", linkConfirm, nil
	}

	// CASE 3: User doesn't exist at all - Create new user
	s.logger.InfoContext(ctx, "creating new user from social auth",
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
		return nil, "", nil, apperror.NewInternal("failed to generate token: %w", err)
	}

	return user, jwtToken, nil, nil
}

// createUserFromOAuth creates a new user from OAuth provider information
func (s *userServiceImpl) createUserFromOAuth(ctx context.Context, userInfo *model.OAuthUserInfo) (*model.User, error) {
	// Generate username from email
	username := strings.Split(userInfo.Email, "@")[0]

	// Check if username exists, append random number if needed
	existingUser, _ := s.repo.GetUserByUsername(ctx, username)
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

	userID, err := s.repo.CreateSocialUser(ctx, user, profile)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create social user", "error", err)
		return nil, apperror.NewInternal("failed to create user account: %w", err)
	}

	user.UserID = userID
	return user, nil
}

// generateJWT creates a JWT token for authenticated user
func (s *userServiceImpl) generateJWT(user *model.User) (string, error) {
	return s.jwtService.GenerateAccessToken(user)
}

// generateLinkToken creates a temporary token for account linking confirmation
func (s *userServiceImpl) generateLinkToken(userID string, providerInfo *model.OAuthUserInfo) (string, error) {
	claims := jwt.MapClaims{
		"user_id":          userID,
		"provider":         providerInfo.Provider,
		"provider_user_id": providerInfo.ProviderUserID,
		"provider_email":   providerInfo.Email,
		"first_name":       providerInfo.FirstName,
		"last_name":        providerInfo.LastName,
		"avatar_url":       providerInfo.AvatarURL,
		"type":             "link_confirmation",
	}

	return s.jwtService.GenerateCustomToken(claims, 10*time.Minute)
}

// ConfirmAccountLink handles user's decision to link or not link accounts
func (s *userServiceImpl) ConfirmAccountLink(ctx context.Context, linkToken string, confirm bool) (*model.User, string, error) {
	// Parse and validate link token
	claims, err := s.jwtService.ValidateCustomToken(linkToken)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to parse link token", "error", err)
		return nil, "", apperror.NewUnauthorized("invalid or expired link token")
	}

	// Verify token type
	if claims["type"] != "link_confirmation" {
		return nil, "", apperror.NewUnauthorized("invalid token type")
	}

	// Extract claims
	userID := claims["user_id"].(string)
	provider := claims["provider"].(string)
	providerUserID := claims["provider_user_id"].(string)

	// Get existing user
	existingUser, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, "", apperror.NewInternal("failed to get user: %w", err)
	}

	if existingUser == nil {
		return nil, "", apperror.NewNotFound("user not found")
	}

	if confirm {
		// User confirmed - Link the accounts
		s.logger.InfoContext(ctx, "user confirmed account linking",
			"user_id", userID,
			"provider", provider,
		)

		// Link the social account
		err = s.repo.LinkSocialAccount(ctx, userID, provider, providerUserID)
		if err != nil {
			return nil, "", apperror.NewInternal("failed to link account: %w", err)
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

		err = s.repo.UpdateSocialUserInfo(ctx, userID, providerInfo)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to update user info after linking", "error", err)
		}

		// Update profile (avatar)
		if providerInfo.AvatarURL != "" {
			profile := &model.Profile{
				AvatarURL: providerInfo.AvatarURL,
			}
			err = s.repo.UpsertSocialUserProfile(ctx, userID, profile)
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to update profile after linking", "error", err)
			}
		}

		// Update user object
		existingUser.AuthProvider = provider
		existingUser.ProviderUserID = providerUserID

		s.logger.InfoContext(ctx, "account linked successfully",
			"user_id", userID,
			"provider", provider,
		)
	} else {
		// User declined - Just login with existing account
		s.logger.InfoContext(ctx, "user declined account linking, using existing account",
			"user_id", userID,
			"provider", provider,
		)
	}

	// Generate JWT token for login
	jwtToken, err := s.generateJWT(existingUser)
	if err != nil {
		return nil, "", apperror.NewInternal("failed to generate token: %w", err)
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
func (s *userServiceImpl) HandleNativeGoogleSignIn(ctx context.Context, idToken string) (*model.User, string, error) {
	// Verify the ID token with Google
	userInfo, err := s.verifyGoogleIDToken(ctx, idToken)
	if err != nil {
		return nil, "", err
	}

	// Find or create user
	user, jwtToken, linkConfirm, err := s.findOrCreateUser(ctx, userInfo)
	if err != nil {
		return nil, "", err
	}

	// Native SSO doesn't support link confirmation - auto-link if user exists
	if linkConfirm != nil {
		return nil, "", apperror.NewBadRequest("account with this email already exists, please use web login to link accounts")
	}

	return user, jwtToken, nil
}

// verifyGoogleIDToken verifies Google ID token from native apps
func (s *userServiceImpl) verifyGoogleIDToken(ctx context.Context, idToken string) (*model.OAuthUserInfo, error) {
	// Call Google's tokeninfo endpoint to verify the token
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken), nil)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create verification request", "error", err)
		return nil, apperror.NewInternal("failed to create verification request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to verify google id token", "error", err)
		return nil, apperror.NewInternal("failed to verify google id token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.WarnContext(ctx, "invalid google id token", "status", resp.StatusCode)
		return nil, apperror.NewUnauthorized("invalid or expired google id token")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to read token verification response", "error", err)
		return nil, apperror.NewInternal("failed to read verification response: %w", err)
	}

	var tokenInfo model.GoogleTokenInfo

	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		s.logger.ErrorContext(ctx, "failed to parse token info", "error", err)
		return nil, apperror.NewInternal("failed to parse token information: %w", err)
	}

	// Verify the audience matches our client ID
	if tokenInfo.Aud != s.googleConfig.ClientID {
		s.logger.WarnContext(ctx, "token audience mismatch",
			"expected", s.googleConfig.ClientID,
			"got", tokenInfo.Aud,
		)
		return nil, apperror.NewUnauthorized("invalid token audience")
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
