package handler

import "github.com/gofiber/fiber/v2"

// GetAdminDashboard godoc
// @Summary      Admin dashboard greeting
// @Description  Returns a simple admin dashboard greeting (admin only).
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.MessageResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      403  {object}  apidoc.ErrorResponse
// @Router       /auth/admin/dashboard [get]
func (h *Handler) GetAdminDashboard(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Welcome to Admin Dashboard",
	})
}

// GetTeacherDashboard godoc
// @Summary      Teacher dashboard greeting
// @Description  Returns a simple teacher dashboard greeting (teacher only).
// @Tags         Teacher
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  apidoc.MessageResponse
// @Failure      401  {object}  apidoc.ErrorResponse
// @Failure      403  {object}  apidoc.ErrorResponse
// @Router       /auth/teacher/dashboard [get]
func (h *Handler) GetTeacherDashboard(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Welcome Teacher",
	})
}
