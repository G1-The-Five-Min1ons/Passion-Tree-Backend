package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GoogleLogin initiates Google OAuth2 flow
// @route GET /auth/google
func (h *Handler) GoogleLogin(c *fiber.Ctx) error {
	// Generate state for CSRF protection
	state := uuid.New().String()
	
	// Store state in session/cookie for validation
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		MaxAge:   300,   // 5 minutes
	})

	authURL := h.socialAuthSvc.GetGoogleAuthURL(state)
	
	return c.JSON(fiber.Map{
		"auth_url": authURL,
	})
}

// GoogleCallback handles Google OAuth2 callback
// @route GET /auth/google/callback
func (h *Handler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" {
		h.logger.Warn("google callback missing code")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing authorization code",
		})
	}

	// Validate state (CSRF protection)
	savedState := c.Cookies("oauth_state")
	if state != savedState {
		h.logger.Warn("google callback state mismatch",
			"received", state,
			"expected", savedState,
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state parameter",
		})
	}

	// Clear state cookie
	c.ClearCookie("oauth_state")

	// Process callback
	user, token, err := h.socialAuthSvc.HandleGoogleCallback(c.UserContext(), code)
	if err != nil {
		h.logger.Error("google callback failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	}

	h.logger.Info("google login successful",
		"user_id", user.UserID,
		"email", user.Email,
	)

	return c.JSON(fiber.Map{
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
		Secure:   false, // Set to true in production with HTTPS
		MaxAge:   300,   // 5 minutes
	})

	authURL := h.socialAuthSvc.GetDiscordAuthURL(state)
	
	return c.JSON(fiber.Map{
		"auth_url": authURL,
	})
}

// DiscordCallback handles Discord OAuth2 callback
// @route GET /auth/discord/callback
func (h *Handler) DiscordCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" {
		h.logger.Warn("discord callback missing code")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing authorization code",
		})
	}

	// Validate state (CSRF protection)
	savedState := c.Cookies("oauth_state")
	if state != savedState {
		h.logger.Warn("discord callback state mismatch",
			"received", state,
			"expected", savedState,
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state parameter",
		})
	}

	// Clear state cookie
	c.ClearCookie("oauth_state")

	// Process callback
	user, token, err := h.socialAuthSvc.HandleDiscordCallback(c.UserContext(), code)
	if err != nil {
		h.logger.Error("discord callback failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	}

	h.logger.Info("discord login successful",
		"user_id", user.UserID,
		"email", user.Email,
	)

	return c.JSON(fiber.Map{
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
