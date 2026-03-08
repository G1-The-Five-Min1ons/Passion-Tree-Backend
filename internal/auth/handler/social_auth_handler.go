package handler

import (
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

// handleOAuthResponse handles the common OAuth response logic
func (h *Handler) handleOAuthResponse(c *fiber.Ctx, provider string, user *model.User, token string, linkConfirm *model.LinkConfirmationNeeded, err error) error {
	if err != nil {
		return h.handleError(c, err)
	}

	if linkConfirm != nil && linkConfirm.NeedsConfirm {
        return c.Status(fiber.StatusMultipleChoices).JSON(fiber.Map{
            "success": false,
            "type":    "LINK_CONFIRMATION_REQUIRED",
            "data":    linkConfirm,
        })
    }

	h.logger.Info(provider + " login successful", "user_id", user.UserID)
	return c.JSON(fiber.Map{
		"success": true,
		"message": provider + " Login successful",
		"data": fiber.Map{
			"token": token,
			"user":  user,
		},
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

	// Validate state (CSRF)
	savedState := c.Cookies("oauth_state")
	if state == "" || state != savedState {
		h.logger.Warn(provider+" callback state mismatch", "received", state, "expected", savedState)
		return h.handleError(c, apperror.NewBadRequest("invalid state parameter"))
	}
	c.ClearCookie("oauth_state")

	switch provider {
	case "google":
		h.logger.Info("handling Google OAuth callback", "code", code)
		user, token, linkConfirm, err := h.socialAuthSvc.HandleGoogleCallback(c.UserContext(), code)
		return h.handleOAuthResponse(c, provider, user, token, linkConfirm, err)
	case "discord":
		h.logger.Info("handling Discord OAuth callback", "code", code)
		user, token, linkConfirm, err := h.socialAuthSvc.HandleDiscordCallback(c.UserContext(), code)
		return h.handleOAuthResponse(c, provider, user, token, linkConfirm, err)
	default:
		h.logger.Warn("unsupported provider in handleOAuth", "provider", provider)
		return h.handleError(c, apperror.NewBadRequest("unsupported provider"))
	}
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
	var req model.NativeGoogleSignInRequest

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
		"data": fiber.Map{
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
	var req model.NativeDiscordSignInRequest

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
		"data": fiber.Map{
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

	// Use config values for mobile app scheme and package
	appScheme := h.config.MobileAppScheme + "://auth/callback"
	appPackage := h.config.MobileAppPackage

	// Android intent:// URL natively closes Chrome Custom Tabs
	buildIntentURL := func(queryPart string) string {
		return "intent://auth/callback" + queryPart +
			"#Intent;scheme=" + h.config.MobileAppScheme + ";package=" + appPackage + ";end"
	}

	renderRedirectPage := func(customSchemeURL string, intentURL string, isError bool) error {
		message := "Authentication Successful. Returning to Passion Tree..."
		if isError {
			message = "Authentication failed. Returning to Passion Tree..."
		}

		// Use template if available, otherwise fallback to inline HTML
		if h.oauthRedirectTpl != nil {
			c.Set("Content-Type", "text/html; charset=utf-8")
			return h.oauthRedirectTpl.Execute(c.Response().BodyWriter(), map[string]string{
				"Message":         message,
				"CustomSchemeURL": customSchemeURL,
				"IntentURL":       intentURL,
			})
		}

		// Fallback to inline HTML if template loading failed
		html := `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Passion Tree — Redirecting</title>
	<style>
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body {
			display: flex; align-items: center; justify-content: center;
			min-height: 100vh; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
			background: #1a1a2e; color: #e0e0e0;
		}
		.container { text-align: center; padding: 2rem; }
		.spinner {
			width: 40px; height: 40px; border: 3px solid rgba(255,255,255,0.1);
			border-top-color: #7289da; border-radius: 50%;
			animation: spin 0.8s linear infinite; margin: 0 auto 1.5rem;
		}
		@keyframes spin { to { transform: rotate(360deg); } }
		h2 { font-size: 1rem; font-weight: 500; margin-bottom: 1rem; }
		a { color: #7289da; font-size: 0.875rem; }
	</style>
</head>
<body>
	<div class="container">
		<div class="spinner"></div>
		<h2>` + message + `</h2>
		<a href="` + customSchemeURL + `">Click here if you are not redirected</a>
	</div>
	<script>
		window.location.href = '` + intentURL + `';
		setTimeout(function() {
			window.location.href = '` + customSchemeURL + `';
		}, 500);
	</script>
</body>
</html>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Status(fiber.StatusOK).SendString(html)
	}

	if errorParam != "" {
		h.logger.Warn("discord native callback received error", "error", errorParam)
		return renderRedirectPage(appScheme+"?error="+errorParam, buildIntentURL("?error="+errorParam), true)
	}

	if code == "" {
		h.logger.Warn("discord native callback missing code")
		return renderRedirectPage(appScheme+"?error=missing_code", buildIntentURL("?error=missing_code"), true)
	}

	h.logger.Info("discord native callback redirecting to app", "has_code", true)
	return renderRedirectPage(appScheme+"?code="+code, buildIntentURL("?code="+code), false)
}

func (h *Handler) ConfirmAccountLink(c *fiber.Ctx) error {
	var req model.ConfirmAccountLinkRequest

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
		"data": fiber.Map{
			"user_id":    user.UserID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	})
}
