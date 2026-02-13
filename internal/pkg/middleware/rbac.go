package middleware

import (
	"log/slog"

	"passiontree/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

// RฺBAC Middleware checks if the user has one of the allowed roles
func RbacMiddleware(logger *slog.Logger, allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, err := GetRoleFromContext(c)
		userID, _ := GetUserIDFromContext(c)

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "unauthorized: " + err.Error(),
			})
		}

		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		logger.WarnContext(c.UserContext(), "rbac_forbidden_access",
            "user_id", userID,
            "user_role", userRole,
            "allowed_roles", allowedRoles,
            "path", c.Path(),
            "method", c.Method(),
            "ip", c.IP(),
        )

		errForbidden := apperror.NewForbidden("insufficient permissions")
		return c.Status(errForbidden.Code).JSON(fiber.Map{
			"success": false,
			"error":   errForbidden.Message,
		})
	}
}
