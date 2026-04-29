package handler

import (
	"context"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// VerifyEmail godoc
// @Summary      Verify email and auto-login
// @Description  Confirms a user's email using the one-time code sent at registration and immediately issues access & refresh tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.VerifyEmailRequest  true  "Verification code"
// @Success      200   {object}  apidoc.TokenPairResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse
// @Router       /auth/verify-email [post]
func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	var req model.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	if req.Code == "" {
		return h.handleError(c, apperror.NewBadRequest("verification code is required"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	deviceInfo := c.Get("User-Agent", "Unknown Device")
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "Unknown")

	accessToken, refreshToken, err := h.userSvc.VerifyEmail(ctx, req.Code, deviceInfo, ipAddress, userAgent)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Email verified and logged in successfully",
		"data": fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

// ResendVerificationEmail godoc
// @Summary      Resend the email-verification code
// @Description  Sends a fresh verification code to the supplied email if the account exists and is unverified.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.ResendVerificationRequest  true  "Email to resend to"
// @Success      200   {object}  apidoc.MessageResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Router       /auth/resend-verification [post]
func (h *Handler) ResendVerificationEmail(c *fiber.Ctx) error {
	var req model.ResendVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.ResendVerificationEmail(ctx, req.Email); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Verification email sent successfully",
	})
}

// ForgotPassword godoc
// @Summary      Request password reset email
// @Description  Always returns 200 to avoid leaking whether an email exists in the system.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.ForgotPasswordRequest  true  "Email to send reset code to"
// @Success      200   {object}  apidoc.MessageResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Router       /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req model.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.ForgotPassword(ctx, req.Email); err != nil {
		return h.handleError(c, err)
	}

	// Security Best Practice: ไม่บอกว่าอีเมลมีในระบบหรือไม่ เพื่อป้องกันการเดา User
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "If the email exists, a password reset code has been sent",
	})
}

// ResetPassword godoc
// @Summary      Reset password using a reset code
// @Description  Updates the user's password if the supplied reset code is valid.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.ResetPasswordRequest  true  "Reset code and new password"
// @Success      200   {object}  apidoc.MessageResponse
// @Failure      400   {object}  apidoc.ErrorResponse  "Invalid code or password"
// @Router       /auth/reset-password [post]
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var req model.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.ResetPassword(ctx, req.Code, req.NewPassword); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Password reset successfully",
	})
}

// ChangePassword godoc
// @Summary      Change password (authenticated)
// @Description  Updates the password for the JWT-authenticated user after verifying the old password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.ChangePasswordRequest  true  "Old and new passwords"
// @Success      200   {object}  apidoc.MessageResponse
// @Failure      400   {object}  apidoc.ErrorResponse
// @Failure      401   {object}  apidoc.ErrorResponse  "Wrong old password or invalid session"
// @Router       /auth/change-password [put]
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	// ดึง userID จาก JWT (Middleware)
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return h.handleError(c, err)
	}

	var req model.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, apperror.NewBadRequest("invalid request body"))
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	if err := h.userSvc.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Password changed successfully",
	})
}
