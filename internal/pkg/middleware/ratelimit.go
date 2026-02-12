package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RateLimitMiddleware() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:               5,                // Login can only be 5 times
        Expiration:        5 * time.Minute,  // 5 minutes
        KeyGenerator: func(c *fiber.Ctx) string {
            clientIP := c.Get("X-Forwarded-For")
            if clientIP == "" {
                clientIP = c.IP()
            }
            return clientIP
        },
        LimitReached: func(c *fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "success": false,
                "message": "Too many login attempts from your IP. Please wait 5 minutes.",
            })
        },
    })
}