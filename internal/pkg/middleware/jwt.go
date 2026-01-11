package middleware

import (
	"strings"

	"passiontree/internal/pkg/apperror"
	"passiontree/internal/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware validates JWT token from Authorization header
func JWTMiddleware() fiber.Handler {
	jwtService := jwt.NewService()

	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "missing authorization header",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "invalid or expired token",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// GetUserIDFromContext extracts user_id from fiber context
func GetUserIDFromContext(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return "", apperror.NewUnauthorized("user not authenticated")
	}
	return userID, nil
}

// GetUsernameFromContext extracts username from fiber context
func GetUsernameFromContext(c *fiber.Ctx) (string, error) {
	username, ok := c.Locals("username").(string)
	if !ok || username == "" {
		return "", apperror.NewUnauthorized("user not authenticated")
	}
	return username, nil
}

// GetRoleFromContext extracts role from fiber context
func GetRoleFromContext(c *fiber.Ctx) (string, error) {
	role, ok := c.Locals("role").(string)
	if !ok || role == "" {
		return "", apperror.NewUnauthorized("user not authenticated")
	}
	return role, nil
}
