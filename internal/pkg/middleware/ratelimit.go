package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RateLimitMiddleware() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,
		Expiration: 5 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// lock with IP for prevent Bot spam (Dos)
			return "limit:ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Too many requests from this IP. Please wait 5 minutes.",
			})
		},
	})
}

// RateLimitMaintenanceMiddleware is a stricter limiter for heavy admin operations
// (BulkSync, Reconcile) — max 3 triggers per hour per IP.
func RateLimitMaintenanceMiddleware() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        3,
		Expiration: 1 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "limit:maintenance:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Too many maintenance requests from this IP. Please wait 1 hour.",
			})
		},
	})
}
