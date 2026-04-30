package middleware

import (
	"goilerplate/config"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

type RateLimiter struct {
	Auth    fiber.Handler
	User    fiber.Handler
	Partner fiber.Handler
}

// NewRateLimiter creates rate limiters for each scope.
// Storage defaults to in-memory — swap to gofiber/storage/redis for multi-instance deployments.
func NewRateLimiter(cfg config.RateLimit) *RateLimiter {
	return &RateLimiter{
		Auth:    newAuthLimiter(cfg.Auth),
		User:    newUserLimiter(cfg.User),
		Partner: newPartnerLimiter(cfg.Partner),
	}
}

// newAuthLimiter limits by IP — protects login/register from brute force.
func newAuthLimiter(cfg config.RateLimitRule) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.TooManyRequests(c, "")
		},
	})
}

// newUserLimiter limits by authenticated user ID — protects authenticated routes from abuse.
func newUserLimiter(cfg config.RateLimitRule) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			if userID, ok := c.Locals(string(constants.ContextKeyUserID)).(string); ok && userID != "" {
				return userID
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.TooManyRequests(c, "")
		},
	})
}

// newPartnerLimiter limits by API key — controls partner consumption.
func newPartnerLimiter(cfg config.RateLimitRule) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			if apiKey := c.Get("x-api-key"); apiKey != "" {
				return apiKey
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.TooManyRequests(c, "")
		},
	})
}
