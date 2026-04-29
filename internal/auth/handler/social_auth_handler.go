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

// GoogleLogin godoc
// @Summary      Initiate Google OAuth2 login (web)
// @Description  Returns a Google authorization URL the client should redirect to.
// @Description  Sets a CSRF state cookie that must be echoed back on callback.
// @Tags         Social Auth
// @Produce      json
// @Success      200  {object}  apidoc.AuthURLResponse
// @Router       /auth/google [get]
func (h *Handler) GoogleLogin(c *fiber.Ctx) error {
	// Generate state for CSRF protection
	state := uuid.New().String()

	// Store state in session/cookie for validation
	h.setOAuthStateCookie(c, state)

	authURL := h.socialAuthSvc.GetGoogleAuthURL(state)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Google OAuth URL generated",
		"data": fiber.Map{
			"auth_url": authURL,
		},
	})
}

// DiscordLogin godoc
// @Summary      Initiate Discord OAuth2 login (web)
// @Description  Returns a Discord authorization URL the client should redirect to.
// @Tags         Social Auth
// @Produce      json
// @Success      200  {object}  apidoc.AuthURLResponse
// @Router       /auth/discord [get]
func (h *Handler) DiscordLogin(c *fiber.Ctx) error {
	// Generate state for CSRF protection
	state := uuid.New().String()

	// Store state in session/cookie for validation
	h.setOAuthStateCookie(c, state)

	authURL := h.socialAuthSvc.GetDiscordAuthURL(state)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Discord OAuth URL generated",
		"data": fiber.Map{
			"auth_url": authURL,
		},
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
			"message": "Account with this email already exists, please confirm linking",
            "type":    "LINK_CONFIRMATION_REQUIRED",
            "data":    linkConfirm,
        })
    }

	h.logger.Info(provider + " login successful", "user_id", user.UserID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
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

// GoogleCallback godoc
// @Summary      Google OAuth2 callback (web)
// @Description  Exchanges the OAuth code for tokens and either logs the user in or returns a link-confirmation payload if the email already maps to another account.
// @Tags         Social Auth
// @Produce      json
// @Param        code   query     string  true   "Authorization code from Google"
// @Param        state  query     string  true   "CSRF state echoed from /auth/google"
// @Success      200    {object}  apidoc.SuccessResponse
// @Failure      300    {object}  apidoc.ErrorResponse  "Account linking confirmation required"
// @Failure      400    {object}  apidoc.ErrorResponse
// @Router       /auth/google/callback [get]
func (h *Handler) GoogleCallback(c *fiber.Ctx) error {
	return h.handleOAuth(c, "google")
}

// DiscordCallback godoc
// @Summary      Discord OAuth2 callback (web)
// @Description  Exchanges the OAuth code for tokens and either logs the user in or returns a link-confirmation payload.
// @Tags         Social Auth
// @Produce      json
// @Param        code   query     string  true   "Authorization code from Discord"
// @Param        state  query     string  true   "CSRF state echoed from /auth/discord"
// @Success      200    {object}  apidoc.SuccessResponse
// @Failure      300    {object}  apidoc.ErrorResponse  "Account linking confirmation required"
// @Failure      400    {object}  apidoc.ErrorResponse
// @Router       /auth/discord/callback [get]
func (h *Handler) DiscordCallback(c *fiber.Ctx) error {
	return h.handleOAuth(c, "discord")
}

// NativeGoogleSignIn godoc
// @Summary      Native Google Sign-In (mobile)
// @Description  Verifies a Google ID token coming from a mobile client and returns access & refresh tokens.
// @Tags         Social Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.NativeGoogleSignInRequest  true  "Google ID token"
// @Success      200   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse  "Invalid ID token"
// @Router       /auth/native/google [post]
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

	deviceInfo := c.Get("User-Agent", "Unknown Device")
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "Unknown")

	// Verify and authenticate user
	user, accessToken, refreshToken, err := h.socialAuthSvc.HandleNativeGoogleSignIn(c.UserContext(), req.IDToken, deviceInfo, ipAddress, userAgent)
	if err != nil {
		h.logger.Error("native google signin failed", "error", err)
		return h.handleError(c, err)
	}

	h.logger.Info("native google signin successful", "user_id", user.UserID, "email", user.Email, "name", user.Username)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":       true,
		"message":       "Login successful",
		"token":         accessToken,
		"refresh_token": refreshToken,
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

// NativeDiscordSignIn godoc
// @Summary      Native Discord Sign-In (mobile)
// @Description  Exchanges a Discord OAuth code captured by the mobile client for access & refresh tokens.
// @Tags         Social Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.NativeDiscordSignInRequest  true  "Discord OAuth code"
// @Success      200   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Router       /auth/native/discord [post]
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

	deviceInfo := c.Get("User-Agent", "Unknown Device")
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "Unknown")

	// Verify and authenticate user
	user, accessToken, refreshToken, linkConfirm, err := h.socialAuthSvc.HandleNativeDiscordSignIn(c.UserContext(), req.Code, deviceInfo, ipAddress, userAgent)
	if err != nil {
		h.logger.Error("native discord signin failed", "error", err)
		return h.handleError(c, err)
	}

	if linkConfirm != nil && linkConfirm.NeedsConfirm {
		return h.handleError(c, apperror.NewBadRequest("account with this email already exists, please use web login to link accounts"))
	}

	h.logger.Info("native discord signin successful", "user_id", user.UserID, "email", user.Email, "name", user.Username)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":       true,
		"message":       "Login successful",
		"token":         accessToken,
		"refresh_token": refreshToken,
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

// DiscordNativeCallback godoc
// @Summary      Discord OAuth2 native bridge
// @Description  Discord redirects here with ?code=...; this endpoint serves an HTML page that immediately redirects to the mobile app's custom URL scheme so flutter_web_auth_2 can capture the code.
// @Tags         Social Auth
// @Produce      html
// @Param        code   query     string  false  "Authorization code from Discord"
// @Param        error  query     string  false  "Error code from Discord"
// @Success      200    {string}  string  "HTML redirect page"
// @Router       /auth/discord/native/callback [get]
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

// ConfirmAccountLink godoc
// @Summary      Confirm or decline OAuth account linking
// @Description  Called after a social login returns a link-confirmation requirement. If `confirm=true`, the social identity is linked to the existing local account.
// @Tags         Social Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.ConfirmAccountLinkRequest  true  "Link token and decision"
// @Success      200   {object}  apidoc.SuccessResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Router       /auth/confirm-link [post]
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
