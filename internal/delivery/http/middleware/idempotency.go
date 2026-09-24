package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"goilerplate/internal/domain/lock"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
)

const (
	idempotencyHeader         = "Idempotency-Key"
	idempotencyReplayedHeader = "Idempotency-Replayed"
	idempotencyMaxKeyLength   = 255

	// idempotencyLockTTL bounds how long an in-flight request holds its key. A duplicate
	// that arrives after this while the first request is still running executes again.
	idempotencyLockTTL = time.Minute
)

type idempotencyRecord struct {
	Fingerprint string `json:"fingerprint"`
	StatusCode  int    `json:"status_code"`
	ContentType string `json:"content_type"`
	// Location is kept so a replayed 201 still names the resource the first request created;
	// without it a retrying client gets the body but loses where the resource lives.
	Location string `json:"location,omitempty"`
	Body     []byte `json:"body"`
}

// NewIdempotency returns a middleware that deduplicates requests using the
// Idempotency-Key header. Requests without the header pass through normally.
//
//   - A completed 2xx response is stored for ttl and replayed for the same key
//     (with the Idempotency-Replayed: true header).
//   - A request with the same key while the first one is still running gets 409.
//   - The same key with a different method, URL, or body gets 422.
//   - Non-2xx responses are not stored, so the client may retry with the same key.
//
// Keys are scoped per authenticated user. Storage and locker decide the scope:
// Redis-backed ones are shared by every instance, memory-backed ones only deduplicate
// within one instance. Apply it after authentication and permission checks.
func NewIdempotency(storage fiber.Storage, locker lock.Provider, ttl time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get(idempotencyHeader)
		if key == "" {
			return c.Next()
		}
		if len(key) > idempotencyMaxKeyLength {
			return response.BadRequest(c, fmt.Sprintf("%s must be at most %d characters", idempotencyHeader, idempotencyMaxKeyLength), nil)
		}

		userID, _ := c.Locals(string(constants.ContextKeyUserID)).(string)
		storageKey := idempotencyStorageKey(userID, key)
		fingerprint := requestFingerprint(c)

		if handled, err := replayStoredResponse(c, storage, storageKey, fingerprint); handled {
			return err
		}

		release, acquired, err := locker.TryLock(c.UserContext(), "idempotency:"+storageKey, idempotencyLockTTL)
		if err != nil {
			logger.Error(c.UserContext(), fmt.Errorf("acquiring idempotency lock: %w", err))
			return response.InternalServerError(c, "")
		}
		if !acquired {
			return response.Conflict(c, "A request with this "+idempotencyHeader+" is still being processed", nil)
		}
		defer func() {
			if err := release(context.WithoutCancel(c.UserContext())); err != nil {
				logger.Error(c.UserContext(), fmt.Errorf("releasing idempotency lock: %w", err))
			}
		}()

		// Another request may have completed between the lookup above and acquiring the lock
		if handled, err := replayStoredResponse(c, storage, storageKey, fingerprint); handled {
			return err
		}

		if err := c.Next(); err != nil {
			return err
		}

		storeResponse(c, storage, storageKey, fingerprint, ttl)
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

// replayStoredResponse writes the stored response for storageKey, a 422 when the key was
// used for a different request, or a 500 when storage fails. handled reports whether a
// response was written.
func replayStoredResponse(c *fiber.Ctx, storage fiber.Storage, storageKey, fingerprint string) (handled bool, err error) {
	data, err := storage.Get(storageKey)
	if err != nil {
		logger.Error(c.UserContext(), fmt.Errorf("reading idempotency record: %w", err))
		return true, response.InternalServerError(c, "")
	}
	if data == nil {
		return false, nil
	}

	var record idempotencyRecord
	if err := json.Unmarshal(data, &record); err != nil {
		logger.Error(c.UserContext(), fmt.Errorf("decoding idempotency record: %w", err))
		return true, response.InternalServerError(c, "")
	}

	if record.Fingerprint != fingerprint {
		return true, response.UnprocessableEntity(c, idempotencyHeader+" was already used for a different request", nil)
	}

	c.Set(idempotencyReplayedHeader, "true")
	if record.ContentType != "" {
		c.Set(fiber.HeaderContentType, record.ContentType)
	}
	if record.Location != "" {
		c.Set(fiber.HeaderLocation, record.Location)
	}
	return true, c.Status(record.StatusCode).Send(record.Body)
}

// storeResponse saves a 2xx response. Failures are logged: the request already succeeded.
func storeResponse(c *fiber.Ctx, storage fiber.Storage, storageKey, fingerprint string, ttl time.Duration) {
	statusCode := c.Response().StatusCode()
	if statusCode < fiber.StatusOK || statusCode >= fiber.StatusMultipleChoices {
		return
	}

	data, err := json.Marshal(idempotencyRecord{
		Fingerprint: fingerprint,
		StatusCode:  statusCode,
		ContentType: string(c.Response().Header.ContentType()),
		Location:    string(c.Response().Header.Peek(fiber.HeaderLocation)),
		Body:        append([]byte{}, c.Response().Body()...),
	})
	if err != nil {
		logger.Error(c.UserContext(), fmt.Errorf("encoding idempotency record: %w", err))
		return
	}

	if err := storage.Set(storageKey, data, ttl); err != nil {
		logger.Error(c.UserContext(), fmt.Errorf("storing idempotency record: %w", err))
	}
}

// idempotencyStorageKey scopes a client key to the user and bounds its length.
func idempotencyStorageKey(userID, key string) string {
	sum := sha256.Sum256([]byte(userID + "\x00" + key))
	return hex.EncodeToString(sum[:])
}

// requestFingerprint identifies what a key was used for: method, URL with query, and body.
func requestFingerprint(c *fiber.Ctx) string {
	hash := sha256.New()
	hash.Write([]byte(c.Method()))
	hash.Write([]byte{0})
	hash.Write([]byte(c.OriginalURL()))
	hash.Write([]byte{0})
	hash.Write(c.Body())
	return hex.EncodeToString(hash.Sum(nil))
}
