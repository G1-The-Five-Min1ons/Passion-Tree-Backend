package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

// GetGoogleAuthURL generates the Google OAuth2 authorization URL
func (s *userServiceImpl) GetGoogleAuthURL(state string) string {
	return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetDiscordAuthURL generates the Discord OAuth2 authorization URL
func (s *userServiceImpl) GetDiscordAuthURL(state string) string {
	return s.discordConfig.AuthCodeURL(state)
}

// Private helper function to handle OAuth callbacks for Google and Discord
// handleOAuthCallback is a generic handler for OAuth callbacks
func (s *userServiceImpl) handleOAuthCallback(ctx context.Context, code string, config *oauth2.Config, fetchUserInfo func(context.Context, *oauth2.Token) (*model.OAuthUserInfo, error), providerName string, opts ...oauth2.AuthCodeOption) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	token, err := config.Exchange(ctx, code, opts...)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to exchange oauth code", "provider", providerName, "error", err)
		return nil, "", nil, apperror.NewInternal("failed to authenticate with %s: %w", providerName, err)
	}

	userInfo, err := fetchUserInfo(ctx, token)
	if err != nil {
		return nil, "", nil, err
	}

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

// HandleDiscordCallback processes the Discord OAuth2 callback from web
func (s *userServiceImpl) HandleDiscordCallback(ctx context.Context, code string) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	return s.handleOAuthCallback(ctx, code, s.discordConfig, s.fetchDiscordUserInfo, "discord")
}

// HandleNativeDiscordSignIn processes native Discord Sign-In from mobile apps.
// It returns (user, accessToken, refreshToken, linkConfirm, err). When linkConfirm
// is non-nil, tokens will be empty — the caller is expected to prompt the user
// to confirm account linking via the web OAuth flow.
func (s *userServiceImpl) HandleNativeDiscordSignIn(ctx context.Context, code, deviceInfo, ipAddress, userAgent string) (*model.User, string, string, *model.LinkConfirmationNeeded, error) {
	// For native Discord login, the authorization request used the backend's
	nativeRedirectUri := s.config.DiscordNativeRedirectURL
	nativeRedirectUriOpt := oauth2.SetAuthURLParam("redirect_uri", nativeRedirectUri)

	user, _, linkConfirm, err := s.handleOAuthCallback(ctx, code, s.discordConfig, s.fetchDiscordUserInfo, "discord", nativeRedirectUriOpt)
	if err != nil {
		return nil, "", "", nil, err
	}

	if linkConfirm != nil {
		return nil, "", "", linkConfirm, nil
	}

	accessToken, refreshToken, err := s.issueTokenPair(ctx, user, deviceInfo, ipAddress, userAgent)
	if err != nil {
		return nil, "", "", nil, err
	}

	return user, accessToken, refreshToken, nil, nil
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
func (s *userServiceImpl) fetchDiscordUserInfo(ctx context.Context, token *oauth2.Token) (*model.OAuthUserInfo, error) {
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

	avatarURL := ""
	if discordUser.Avatar != "" {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUser.ID, discordUser.Avatar)
	}

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
		ProviderUserID:   discordUser.ID,
		Email:            discordUser.Email,
		FirstName:        firstName,
		LastName:         lastName,
		AvatarURL:        avatarURL,
		Provider:         model.AuthProviderDiscord,
		ProviderUsername: discordUser.Username,
	}, nil
}

// findOrCreateUser finds an existing user or creates a new one from OAuth info
func (s *userServiceImpl) findOrCreateUser(ctx context.Context, userInfo *model.OAuthUserInfo) (*model.User, string, *model.LinkConfirmationNeeded, error) {
	// Check if user exists with this provider
	user, err := s.repo.GetUserByProvider(ctx, userInfo.Provider, userInfo.ProviderUserID)
	if err != nil {
		return nil, "", nil, apperror.NewInternal("failed to check user existence: %w", err)
	}

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

		user.FirstName = userInfo.FirstName
		user.LastName = userInfo.LastName
		user.Email = userInfo.Email

		jwtToken, err := s.generateJWT(user)
		if err != nil {
			return nil, "", nil, apperror.NewInternal("failed to generate token: %w", err)
		}

		return user, jwtToken, nil, nil
	}

	existingUser, err := s.repo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, "", nil, apperror.NewInternal("failed to check user by email: %w", err)
	}

	if existingUser != nil {
		s.logger.InfoContext(ctx, "user exists with email, need confirmation to link",
			"user_id", existingUser.UserID,
			"email", userInfo.Email,
			"existing_provider", existingUser.AuthProvider,
			"new_provider", userInfo.Provider,
		)
		linkToken, err := s.generateLinkToken(existingUser.UserID, userInfo)
		if err != nil {
			return nil, "", nil, apperror.NewInternal("failed to generate link token: %w", err)
		}

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

	s.logger.InfoContext(ctx, "creating new user from social auth",
		"email", userInfo.Email,
		"provider", userInfo.Provider,
	)

	user, err = s.createUserFromOAuth(ctx, userInfo)
	if err != nil {
		return nil, "", nil, err
	}
	jwtToken, err := s.generateJWT(user)
	if err != nil {
		return nil, "", nil, apperror.NewInternal("failed to generate token: %w", err)
	}

	return user, jwtToken, nil, nil
}

// createUserFromOAuth creates a new user from OAuth provider information
func (s *userServiceImpl) createUserFromOAuth(ctx context.Context, userInfo *model.OAuthUserInfo) (*model.User, error) {
	username := userInfo.ProviderUsername
	if username == "" {
		username = strings.Split(userInfo.Email, "@")[0]
	}

	existingUser, _ := s.repo.GetUserByUsername(ctx, username)
	if existingUser != nil {
		username = fmt.Sprintf("%s_%04d", username, time.Now().UnixNano()%10000)
	}

	user := &model.User{
		Username:        username,
		Email:           userInfo.Email,
		FirstName:       userInfo.FirstName,
		LastName:        userInfo.LastName,
		Role:            model.RolePending,
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

func (s *userServiceImpl) generateJWT(user *model.User) (string, error) {
	return s.jwtService.GenerateAccessToken(user)
}

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
	claims, err := s.jwtService.ValidateCustomToken(linkToken)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to parse link token", "error", err)
		return nil, "", apperror.NewUnauthorized("invalid or expired link token")
	}

	if claims["type"] != "link_confirmation" {
		return nil, "", apperror.NewUnauthorized("invalid token type")
	}

	userID := claims["user_id"].(string)
	provider := claims["provider"].(string)
	providerUserID := claims["provider_user_id"].(string)

	existingUser, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, "", apperror.NewInternal("failed to get user: %w", err)
	}

	if existingUser == nil {
		return nil, "", apperror.NewNotFound("user not found")
	}

	if confirm {
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
			s.logger.ErrorContext(ctx, "failed to update user info after linking",
				"user_id", userID,
				"provider", provider,
				"error", err)
			// Continue - non-critical failure
		}

		// Update profile (avatar)
		if providerInfo.AvatarURL != "" {
			profile := &model.Profile{
				AvatarURL: providerInfo.AvatarURL,
			}
			err = s.repo.UpsertSocialUserProfile(ctx, userID, profile)
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to update profile after linking",
					"user_id", userID,
					"error", err)
				// Continue - non-critical failure
			}
		}

		existingUser.AuthProvider = provider
		existingUser.ProviderUserID = providerUserID

		s.logger.InfoContext(ctx, "account linked successfully",
			"user_id", userID,
			"provider", provider,
		)
	} else {
		s.logger.InfoContext(ctx, "user declined account linking, using existing account",
			"user_id", userID,
			"provider", provider,
		)
	}

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

// HandleNativeGoogleSignIn processes native Google Sign-In from mobile apps.
// Returns (user, accessToken, refreshToken, err). Unlike the web flow, native
// Google Sign-In auto-links to an existing local account if the email already
// exists, so this never returns a LinkConfirmationNeeded.
func (s *userServiceImpl) HandleNativeGoogleSignIn(ctx context.Context, idToken, deviceInfo, ipAddress, userAgent string) (*model.User, string, string, error) {
	userInfo, err := s.verifyGoogleIDToken(ctx, idToken)
	if err != nil {
		return nil, "", "", err
	}

	user, _, linkConfirm, err := s.findOrCreateUser(ctx, userInfo)
	if err != nil {
		return nil, "", "", err
	}

	// Native SSO doesn't support link confirmation dialog - auto-link the account
	if linkConfirm != nil {
		existingUser, _, err := s.repo.GetUserByID(ctx, linkConfirm.ExistingUser.UserID)
		if err != nil || existingUser == nil {
			return nil, "", "", apperror.NewInternal("failed to get existing user: %w", err)
		}

		// Auto-link the social account to the existing user
		err = s.repo.LinkSocialAccount(ctx, existingUser.UserID, userInfo.Provider, userInfo.ProviderUserID)
		if err != nil {
			return nil, "", "", apperror.NewInternal("failed to link account: %w", err)
		}

		err = s.repo.UpdateSocialUserInfo(ctx, existingUser.UserID, userInfo)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to update user info after auto-linking",
				"user_id", existingUser.UserID,
				"provider", userInfo.Provider,
				"error", err)
			// Continue - non-critical failure
		}

		// Update profile (avatar)
		if userInfo.AvatarURL != "" {
			profile := &model.Profile{AvatarURL: userInfo.AvatarURL}
			err = s.repo.UpsertSocialUserProfile(ctx, existingUser.UserID, profile)
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to update profile after auto-linking",
					"user_id", existingUser.UserID,
					"error", err)
				// Continue - non-critical failure
			}
		}

		s.logger.InfoContext(ctx, "native google signin auto-linked account",
			"user_id", existingUser.UserID,
			"provider", userInfo.Provider,
		)

		user = existingUser
	}

	accessToken, refreshToken, err := s.issueTokenPair(ctx, user, deviceInfo, ipAddress, userAgent)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

// issueTokenPair generates a new access + refresh token pair for the user and
// persists the refresh token into the Token table (mirrors the behaviour of
// VerifyEmail / RefreshAccessToken). Returned tokens are ready to be handed back
// to the client.
func (s *userServiceImpl) issueTokenPair(ctx context.Context, user *model.User, deviceInfo, ipAddress, userAgent string) (string, string, error) {
	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate tokens for social login", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to generate session")
	}

	now := time.Now()
	refreshTTL := parseDurationOrHours(s.config.JWTRefreshTTL, 720*time.Hour)
	refreshAbsolute := parseDurationOrHours(s.config.JWTRefreshAbsolute, 720*time.Hour)
	maxExpiresAt := now.Add(refreshAbsolute)
	expireAt := clampExpiry(now.Add(refreshTTL), &maxExpiresAt)

	tokenModel := &model.Token{
		UserID:       user.UserID,
		Token:        refreshToken,
		TokenType:    model.TokenTypeRefresh,
		ExpireAt:     expireAt,
		DeviceInfo:   &deviceInfo,
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
		LastUsedAt:   &now,
		MaxExpiresAt: &maxExpiresAt,
	}

	if err := s.repo.CreateToken(ctx, tokenModel); err != nil {
		s.logger.ErrorContext(ctx, "failed to store refresh token for social login", "error", err, "user_id", user.UserID)
		return "", "", apperror.NewInternal("failed to create session")
	}

	return accessToken, refreshToken, nil
}

// verifyGoogleIDToken verifies Google ID token from native apps
// Uses google.golang.org/api/idtoken for secure token verification
func (s *userServiceImpl) verifyGoogleIDToken(ctx context.Context, idTokenString string) (*model.OAuthUserInfo, error) {
	payload, err := idtoken.Validate(ctx, idTokenString, s.googleConfig.ClientID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to validate google id token", "error", err)

		if strings.Contains(err.Error(), "token expired") {
			return nil, apperror.NewUnauthorized("google id token has expired")
		}
		if strings.Contains(err.Error(), "audience") {
			return nil, apperror.NewUnauthorized("invalid token audience - token was not issued for this app")
		}

		return nil, apperror.NewUnauthorized("invalid google id token: %v", err)
	}

	claims := payload.Claims

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		s.logger.ErrorContext(ctx, "missing sub claim in google token")
		return nil, apperror.NewUnauthorized("invalid google id token: missing user ID")
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		s.logger.ErrorContext(ctx, "missing email claim in google token")
		return nil, apperror.NewUnauthorized("invalid google id token: missing email")
	}

	emailVerified, _ := claims["email_verified"].(bool)
	if !emailVerified {
		s.logger.WarnContext(ctx, "google email not verified", "email", email)
		return nil, apperror.NewUnauthorized("google account email is not verified")
	}

	givenName, _ := claims["given_name"].(string)
	familyName, _ := claims["family_name"].(string)
	picture, _ := claims["picture"].(string)

	if givenName == "" {
		if name, ok := claims["name"].(string); ok && name != "" {
			parts := strings.Fields(name)
			if len(parts) > 0 {
				givenName = parts[0]
				if len(parts) > 1 {
					familyName = strings.Join(parts[1:], " ")
				}
			}
		}
	}

	s.logger.InfoContext(ctx, "google id token verified successfully",
		"sub", sub,
		"email", email,
	)

	return &model.OAuthUserInfo{
		ProviderUserID: sub,
		Email:          email,
		FirstName:      givenName,
		LastName:       familyName,
		AvatarURL:      picture,
		Provider:       model.AuthProviderGoogle,
	}, nil
}
