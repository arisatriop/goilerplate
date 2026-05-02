package middleware

import (
	"encoding/json"
	"time"

	"goilerplate/pkg/constants"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
)

const idempotencyHeader = "Idempotency-Key"

type idempotencyCached struct {
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
}

// NewIdempotency returns a middleware that deduplicates POST requests using
// the Idempotency-Key header. Requests without the header pass through normally.
// When storage is nil (Redis disabled), the middleware is a no-op.
// Apply per-route on sensitive endpoints (payments, transfers, etc.).
func NewIdempotency(storage fiber.Storage, ttl time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if storage == nil {
			return c.Next()
		}

		key := c.Get(idempotencyHeader)
		if key == "" {
			return c.Next()
		}

		userID, _ := c.Locals(string(constants.ContextKeyUserID)).(string)
		cacheKey := userID + ":" + key

		// Return cached response for duplicate request
		if cached, err := storage.Get(cacheKey); err == nil && cached != nil {
			var resp idempotencyCached
			if err := json.Unmarshal(cached, &resp); err == nil {
				c.Set("Idempotency-Replayed", "true")
				return c.Status(resp.StatusCode).Send(resp.Body)
			}
		}

		if err := c.Next(); err != nil {
			return err
		}

		// Only cache successful responses (2xx)
		statusCode := c.Response().StatusCode()
		if statusCode < 200 || statusCode >= 300 {
			return nil
		}

		resp := idempotencyCached{
			StatusCode: statusCode,
			Body:       append([]byte{}, c.Response().Body()...),
		}
		data, err := json.Marshal(resp)
		if err != nil {
			return nil
		}

		_ = storage.Set(cacheKey, data, ttl)

		return nil
	}
}

// RequireIdempotencyKey returns a middleware that rejects requests missing
// the Idempotency-Key header with 400. Use before NewIdempotency on
// endpoints where the key is mandatory.
func RequireIdempotencyKey() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get(idempotencyHeader) == "" {
			return response.BadRequest(c, "Idempotency-Key header is required", nil)
		}
		return c.Next()
	}
}
