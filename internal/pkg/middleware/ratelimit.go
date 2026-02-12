package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RateLimitMiddleware() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        5,               
        Expiration: 5 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            ip := c.IP()

            var body struct {
                Identifier string `json:"identifier"`
            }
            _ = c.BodyParser(&body)

            if body.Identifier != "" {
                return "limit:auth:" + ip + ":" + body.Identifier
            }
            return "limit:auth:" + ip
        },
        LimitReached: func(c *fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "success": false,
                "error":   "Too many attempts from your IP or account. Please wait 5 minutes.",
            })
        },
    })
}