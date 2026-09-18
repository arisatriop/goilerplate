package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"goilerplate/config"
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const limitRule = 3

func newLimiterApp(t *testing.T, limiterOf func(*RateLimiter) fiber.Handler, before fiber.Handler) *fiber.App {
	t.Helper()

	rl := NewRateLimiter(config.RateLimit{
		Auth:    config.RateLimitRule{Max: limitRule, Expiration: time.Minute},
		User:    config.RateLimitRule{Max: limitRule, Expiration: time.Minute},
		Partner: config.RateLimitRule{Max: limitRule, Expiration: time.Minute},
	}, nil)

	app := fiber.New()
	if before != nil {
		app.Use(before)
	}
	app.Use(limiterOf(rl))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	return app
}

func status(t *testing.T, app *fiber.App, headers map[string]string) int {
	t.Helper()

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp.StatusCode
}

// Two sessions from the same IP must not share a budget. Keyed by IP, one user refreshing in a
// few tabs would lock out everyone else behind the same NAT.
func TestSessionLimiter_BudgetIsPerSessionNotPerIP(t *testing.T) {
	// Arrange: the session ID comes from a header only so the test can vary it; in production
	// the auth middleware sets this local.
	app := newLimiterApp(t,
		func(rl *RateLimiter) fiber.Handler { return rl.Session },
		func(c *fiber.Ctx) error {
			c.Locals(string(constants.ContextKeySessionID), c.Get("X-Test-Session"))
			return c.Next()
		})

	// Act: exhaust session A
	for i := range limitRule {
		require.Equal(t, fiber.StatusNoContent, status(t, app, map[string]string{"X-Test-Session": "session-a"}),
			"request %d of session A should pass", i+1)
	}

	// Assert
	assert.Equal(t, fiber.StatusTooManyRequests, status(t, app, map[string]string{"X-Test-Session": "session-a"}),
		"session A is over its budget")
	assert.Equal(t, fiber.StatusNoContent, status(t, app, map[string]string{"X-Test-Session": "session-b"}),
		"session B has its own budget from the same IP")
}

func TestAuthLimiter_BudgetIsPerIP(t *testing.T) {
	app := newLimiterApp(t, func(rl *RateLimiter) fiber.Handler { return rl.Auth }, nil)

	for range limitRule {
		require.Equal(t, fiber.StatusNoContent, status(t, app, nil))
	}

	assert.Equal(t, fiber.StatusTooManyRequests, status(t, app, nil))
}

// Different keys get different budgets, and the raw key must not be what identifies them.
func TestPartnerLimiter_KeyedByHashedAPIKey(t *testing.T) {
	app := newLimiterApp(t, func(rl *RateLimiter) fiber.Handler { return rl.Partner }, nil)

	for range limitRule {
		require.Equal(t, fiber.StatusNoContent, status(t, app, map[string]string{"x-api-key": "partner-one-secret"}))
	}

	assert.Equal(t, fiber.StatusTooManyRequests, status(t, app, map[string]string{"x-api-key": "partner-one-secret"}))
	assert.Equal(t, fiber.StatusNoContent, status(t, app, map[string]string{"x-api-key": "partner-two-secret"}),
		"a different key has its own budget")
}

// The stored key is a hash, so a dump of the limiter's storage does not hand over API keys.
func TestPartnerLimiter_StorageNeverHoldsTheRawKey(t *testing.T) {
	const apiKey = "partner-one-secret"

	storage := &recordingStorage{keys: map[string]bool{}}
	rl := NewRateLimiter(config.RateLimit{
		Partner: config.RateLimitRule{Max: limitRule, Expiration: time.Minute},
	}, storage)

	app := fiber.New()
	app.Use(rl.Partner)
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	status(t, app, map[string]string{"x-api-key": apiKey})

	require.NotEmpty(t, storage.keys, "the limiter stored something to check against")
	for key := range storage.keys {
		assert.NotContains(t, key, apiKey, "the raw API key must never become a storage key")
	}
}

// recordingStorage is a fiber.Storage that remembers the keys it was asked about.
type recordingStorage struct {
	keys map[string]bool
}

func (s *recordingStorage) Get(key string) ([]byte, error) {
	s.keys[key] = true
	return nil, nil
}

func (s *recordingStorage) Set(key string, _ []byte, _ time.Duration) error {
	s.keys[key] = true
	return nil
}

func (s *recordingStorage) Delete(key string) error {
	s.keys[key] = true
	return nil
}

func (s *recordingStorage) Reset() error { return nil }
func (s *recordingStorage) Close() error { return nil }

var _ fiber.Storage = (*recordingStorage)(nil)
