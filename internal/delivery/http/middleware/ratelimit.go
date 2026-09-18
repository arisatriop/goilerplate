package middleware

import (
	"goilerplate/config"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/hash"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

type RateLimiter struct {
	// Auth limits unauthenticated endpoints by IP: login, register, and later forgot-password
	// and OTP. These are the ones an attacker can hammer without credentials.
	Auth fiber.Handler
	// Session limits endpoints that already carry a valid token but run before the user
	// limiter, keyed per session rather than per IP. Everyone behind one NAT shares an IP, so
	// keying these by IP would let one user throttle a whole office.
	Session fiber.Handler
	User    fiber.Handler
	Partner fiber.Handler
}

// NewRateLimiter creates rate limiters for each scope.
// Pass a non-nil storage to use Redis (recommended for multi-instance deployments).
// Passing nil falls back to in-memory storage (single-instance / dev only).
func NewRateLimiter(cfg config.RateLimit, storage fiber.Storage) *RateLimiter {
	return &RateLimiter{
		Auth:    newAuthLimiter(cfg.Auth, storage),
		Session: newSessionLimiter(cfg.User, storage),
		User:    newUserLimiter(cfg.User, storage),
		Partner: newPartnerLimiter(cfg.Partner, storage),
	}
}

// newAuthLimiter limits by IP — protects login/register from brute force.
// Only unauthenticated endpoints belong here: an authenticated one keyed by IP punishes every
// user behind the same NAT for one user's traffic.
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

// newSessionLimiter limits refresh and logout per session. It runs after the token has been
// verified, so the session ID is available; falling back to the IP covers the window where an
// early failure means no session was ever established.
func newSessionLimiter(cfg config.RateLimitRule, storage fiber.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		Storage:    storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			if sessionID, ok := c.Locals(string(constants.ContextKeySessionID)).(string); ok && sessionID != "" {
				return "session:" + sessionID
			}
			return "session:" + c.IP()
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
			// Hashed, not raw: the key would otherwise sit in Redis under a readable name and
			// in any dump of the limiter's storage.
			if apiKey := c.Get("x-api-key"); apiKey != "" {
				return "partner:" + hash.Token(apiKey)
			}
			return "partner:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.TooManyRequests(c, "")
		},
	})
}
