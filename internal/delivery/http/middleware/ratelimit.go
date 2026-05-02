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
// Pass a non-nil storage to use Redis (recommended for multi-instance deployments).
// Passing nil falls back to in-memory storage (single-instance / dev only).
func NewRateLimiter(cfg config.RateLimit, storage fiber.Storage) *RateLimiter {
	return &RateLimiter{
		Auth:    newAuthLimiter(cfg.Auth, storage),
		User:    newUserLimiter(cfg.User, storage),
		Partner: newPartnerLimiter(cfg.Partner, storage),
	}
}

// newAuthLimiter limits by IP — protects login/register from brute force.
func newAuthLimiter(cfg config.RateLimitRule, storage fiber.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		Storage:    storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "auth:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.TooManyRequests(c, "")
		},
	})
}

// newUserLimiter limits by authenticated user ID — protects authenticated routes from abuse.
func newUserLimiter(cfg config.RateLimitRule, storage fiber.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		Storage:    storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			if userID, ok := c.Locals(string(constants.ContextKeyUserID)).(string); ok && userID != "" {
				return "user:" + userID
			}
			return "user:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.TooManyRequests(c, "")
		},
	})
}

// newPartnerLimiter limits by API key — controls partner consumption.
func newPartnerLimiter(cfg config.RateLimitRule, storage fiber.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		Storage:    storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			if apiKey := c.Get("x-api-key"); apiKey != "" {
				return "partner:" + apiKey
			}
			return "partner:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.TooManyRequests(c, "")
		},
	})
}
