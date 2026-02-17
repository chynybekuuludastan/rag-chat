package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/google/uuid"
)

func RateLimiter(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			if userID, ok := c.Locals("userID").(uuid.UUID); ok {
				return userID.String()
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
		},
	})
}

func AuthRateLimiter() fiber.Handler {
	return RateLimiter(10, 1*time.Minute)
}

func UploadRateLimiter() fiber.Handler {
	return RateLimiter(5, 1*time.Minute)
}

func ChatRateLimiter() fiber.Handler {
	return RateLimiter(20, 1*time.Minute)
}

func DefaultRateLimiter() fiber.Handler {
	return RateLimiter(60, 1*time.Minute)
}
