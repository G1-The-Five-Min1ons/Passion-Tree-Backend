package handler

import (
	"fmt"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) setOAuthStateCookie(c *fiber.Ctx, state string) {
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   h.isSecureCookie(),
		SameSite: "Lax",
		MaxAge:   300,
		Path:     "/",
	})
}

func (h *Handler) GoogleLogin(c *fiber.Ctx) error {
	// Generate state for CSRF protection
	state := uuid.New().String()

	// Store state in session/cookie for validation
	h.setOAuthStateCookie(c, state)

	authURL := h.socialAuthSvc.GetGoogleAuthURL(state)

	return c.JSON(fiber.Map{
		"auth_url": authURL,
	})
}

// DiscordLogin initiates Discord OAuth2 flow
// @route GET /auth/discord
func (h *Handler) DiscordLogin(c *fiber.Ctx) error {
	// Generate state for CSRF protection
	state := uuid.New().String()

	// Store state in session/cookie for validation
	h.setOAuthStateCookie(c, state)

	authURL := h.socialAuthSvc.GetDiscordAuthURL(state)

	return c.JSON(fiber.Map{
		"auth_url": authURL,
	})
}

// handleOAuth เป็นฟังก์ชันกลางที่รวม Logic การจัดการ HTTP ไว้ที่เดียว
func (h *Handler) handleOAuth(c *fiber.Ctx, provider string) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		h.logger.Warn(provider + " callback missing code")
		return h.handleError(c, apperror.NewBadRequest("missing authorization code"))
	}

	// 1. Validate state (CSRF)
	savedState := c.Cookies("oauth_state")
	if state == "" || state != savedState {
		h.logger.Warn(provider+" callback state mismatch", "received", state, "expected", savedState)
		return h.handleError(c, apperror.NewBadRequest("invalid state parameter"))
	}
	c.ClearCookie("oauth_state")

	var (
		user        *model.User
		token       string
		linkConfirm *model.LinkConfirmationNeeded
		err         error
	)

	switch provider {
	case "google":
		h.logger.Info("handling Google OAuth callback", "code", code)
		user, token, linkConfirm, err = h.socialAuthSvc.HandleGoogleCallback(c.UserContext(), code)
	case "discord":
		h.logger.Info("handling Discord OAuth callback", "code", code)
		user, token, linkConfirm, err = h.socialAuthSvc.HandleDiscordCallback(c.UserContext(), code)
	default:
		h.logger.Warn("unsupported provider in handleOAuth", "provider", provider)
		return h.handleError(c, apperror.NewBadRequest("unsupported provider"))
	}

	if err != nil {
		return h.handleError(c, err)
	}

	// 3. Manage Link Confirmation
	if linkConfirm != nil && linkConfirm.NeedsConfirm {
		return c.JSON(linkConfirm)
	}

	h.logger.Info(provider+" login successful", "user_id", user.UserID)
	return c.JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}

// GoogleCallback handles Google OAuth2 callback
// @route GET /auth/google/callback
func (h *Handler) GoogleCallback(c *fiber.Ctx) error {
	return h.handleOAuth(c, "google")
}

// DiscordCallback handles Discord OAuth2 callback
// @route GET /auth/discord/callback
func (h *Handler) DiscordCallback(c *fiber.Ctx) error {
	return h.handleOAuth(c, "discord")
}

// NativeGoogleSignIn handles native Google Sign-In from mobile apps
// @route POST /auth/native/google
func (h *Handler) NativeGoogleSignIn(c *fiber.Ctx) error {
	var req struct {
		IDToken string `json:"id_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("invalid request body", "error", err)
		return h.handleError(c, apperror.NewBadRequest("Invalid request body"))
	}

	if req.IDToken == "" {
		h.logger.Warn("missing id_token in native google signin")
		return h.handleError(c, apperror.NewBadRequest("ID token is required"))
	}

	// Verify and authenticate user
	user, token, err := h.socialAuthSvc.HandleNativeGoogleSignIn(c.UserContext(), req.IDToken)
	if err != nil {
		h.logger.Error("native google signin failed", "error", err)
		return h.handleError(c, err)
	}

	h.logger.Info("native google signin successful", "user_id", user.UserID, "email", user.Email, "name", user.Username)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"token":   token,
		"user": fiber.Map{
			"user_id":    user.UserID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	})
}

// NativeDiscordSignIn handles native Discord Sign-In from mobile apps
// @route POST /auth/native/discord
func (h *Handler) NativeDiscordSignIn(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}

	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("invalid request body", "error", err)
		return h.handleError(c, apperror.NewBadRequest("Invalid request body"))
	}

	if req.Code == "" {
		h.logger.Warn("missing code in native discord signin")
		return h.handleError(c, apperror.NewBadRequest("Authorization code is required"))
	}

	// Verify and authenticate user
	user, token, linkConfirm, err := h.socialAuthSvc.HandleNativeDiscordSignIn(c.UserContext(), req.Code)
	if err != nil {
		h.logger.Error("native discord signin failed", "error", err)
		return h.handleError(c, err)
	}

	if linkConfirm != nil && linkConfirm.NeedsConfirm {
		return h.handleError(c, apperror.NewBadRequest("account with this email already exists, please use web login to link accounts"))
	}

	h.logger.Info("native discord signin successful", "user_id", user.UserID, "email", user.Email, "name", user.Username)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"token":   token,
		"user": fiber.Map{
			"user_id":    user.UserID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	})
}

// DiscordNativeCallback handles the Discord OAuth2 redirect for native mobile apps.
// Discord redirects here with ?code=..., and this handler redirects to the
// app's custom URL scheme so flutter_web_auth_2 can capture the code.
// @route GET /auth/discord/native/callback
func (h *Handler) DiscordNativeCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	errorParam := c.Query("error")

	// Build the deep link redirect URL
	const appScheme = "passiontree://auth/callback"

	if errorParam != "" {
		redirectURL := fmt.Sprintf("%s?error=%s", appScheme, errorParam)
		h.logger.Warn("discord native callback received error", "error", errorParam)
		return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
	}

	if code == "" {
		redirectURL := fmt.Sprintf("%s?error=missing_code", appScheme)
		h.logger.Warn("discord native callback missing code")
		return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
	}

	// Redirect to app with the authorization code
	redirectURL := fmt.Sprintf("%s?code=%s", appScheme, code)
	h.logger.Info("discord native callback redirecting to app", "has_code", true)
	return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}

func (h *Handler) ConfirmAccountLink(c *fiber.Ctx) error {
	var req struct {
		LinkToken string `json:"link_token"`
		Confirm   bool   `json:"confirm"`
	}

	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("invalid request body", "error", err)
		return h.handleError(c, apperror.NewBadRequest("Invalid request body"))
	}

	if req.LinkToken == "" {
		h.logger.Warn("missing link_token in confirm request")
		return h.handleError(c, apperror.NewBadRequest("Link token is required"))
	}

	// Process confirmation
	user, token, err := h.socialAuthSvc.ConfirmAccountLink(c.UserContext(), req.LinkToken, req.Confirm)
	if err != nil {
		h.logger.Error("account link confirmation failed", "error", err)
		return h.handleError(c, err)
	}

	action := "declined"
	if req.Confirm {
		action = "confirmed and linked"
	}

	h.logger.Info("account link decision processed",
		"user_id", user.UserID,
		"action", action,
	)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Account link confirmation processed",
		"token":   token,
		"linked":  req.Confirm,
		"user": fiber.Map{
			"user_id":    user.UserID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	})
}
