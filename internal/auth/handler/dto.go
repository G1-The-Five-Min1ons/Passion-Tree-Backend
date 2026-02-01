package handler

import (
	"context"
	"time"

	"passiontree/internal/auth/model"
	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// VerifyEmail - ยืนยันอีเมลด้วยรหัส Code
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

	if err := h.userSvc.VerifyEmail(ctx, req.Code); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Email verified successfully",
	})
}

// ResendVerificationEmail - ส่งอีเมลยืนยันตัวตนอีกครั้ง
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

// ForgotPassword - ร้องขอการรีเซ็ตรหัสผ่าน (Public Route)
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

// ResetPassword - รีเซ็ตรหัสผ่านใหม่โดยใช้ Code
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

// ChangePassword - เปลี่ยนรหัสผ่าน (ต้องผ่าน Auth Middleware)
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
