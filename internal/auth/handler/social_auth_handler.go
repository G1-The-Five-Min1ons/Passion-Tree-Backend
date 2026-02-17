package handler

import (
	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) GoogleLogin(c *fiber.Ctx) error {
	// Generate state for CSRF protection
	state := uuid.New().String()

	// Store state in session/cookie for validation
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   h.isSecureCookie(),
		MaxAge:   300, // 5 minutes
	})

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
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   h.isSecureCookie(),
		MaxAge:   300, // 5 minutes
	})

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
        return c.JSON(fiber.Map{
            "needs_confirmation": true,
            "link_token":         linkConfirm.LinkToken,
            "message":            linkConfirm.Message,
            "existing_user":      linkConfirm.ExistingUser,
            "provider_name":      linkConfirm.ProviderName,
        })
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.IDToken == "" {
		h.logger.Warn("missing id_token in native google signin")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID token is required",
		})
	}

	// Verify and authenticate user
	user, token, err := h.socialAuthSvc.HandleNativeGoogleSignIn(c.UserContext(), req.IDToken)
	if err != nil {
		h.logger.Error("native google signin failed", "error", err)
		return h.handleError(c, err)
	}

	h.logger.Info("native google signin successful",
		"user_id", user.UserID,
		"email", user.Email,
	)

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

func (h *Handler) ConfirmAccountLink(c *fiber.Ctx) error {
	var req struct {
		LinkToken string `json:"link_token"`
		Confirm   bool   `json:"confirm"`
	}

	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("invalid request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.LinkToken == "" {
		h.logger.Warn("missing link_token in confirm request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Link token is required",
		})
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
		"message": "Login successful",
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
