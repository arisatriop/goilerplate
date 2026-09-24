package middleware

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"goilerplate/internal/domain/lock"
	pkgcache "goilerplate/pkg/cache"
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeLocker is an in-process lock.Provider that can simulate failures.
type fakeLocker struct {
	mu   sync.Mutex
	held map[string]bool
	err  error
}

func newFakeLocker() *fakeLocker {
	return &fakeLocker{held: make(map[string]bool)}
}

func (l *fakeLocker) TryLock(_ context.Context, key string, _ time.Duration) (lock.Release, bool, error) {
	if l.err != nil {
		return nil, false, l.err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.held[key] {
		return nil, false, nil
	}
	l.held[key] = true

	return func(context.Context) error {
		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.held, key)
		return nil
	}, true, nil
}

// failingStorage returns an error from every Get.
type failingStorage struct{ *pkgcache.MemoryStorage }

func (failingStorage) Get(string) ([]byte, error) { return nil, errors.New("redis down") }

type idempotencyTestApp struct {
	app      *fiber.App
	executed atomic.Int32
	status   atomic.Int32
	block    chan struct{} // when non-nil, the first execution waits for it to be closed
	started  chan struct{} // receives when the first execution starts waiting
}

func newIdempotencyTestApp(storage fiber.Storage, locker lock.Provider) *idempotencyTestApp {
	ta := &idempotencyTestApp{app: fiber.New()}
	ta.status.Store(fiber.StatusCreated)

	ta.app.Use(func(c *fiber.Ctx) error {
		c.Locals(string(constants.ContextKeyUserID), c.Get("X-Test-User"))
		return c.Next()
	})
	ta.app.Post("/orders", NewIdempotency(storage, locker, time.Hour), func(c *fiber.Ctx) error {
		n := ta.executed.Add(1)
		if n == 1 && ta.block != nil {
			ta.started <- struct{}{}
			<-ta.block
		}
		c.Location(fmt.Sprintf("/orders/%d", n))
		return c.Status(int(ta.status.Load())).JSON(fiber.Map{"order": n})
	})

	return ta
}

type testResponse struct {
	status   int
	body     string
	replayed string
	ctype    string
	location string
}

func (ta *idempotencyTestApp) post(t *testing.T, key, user, body string) testResponse {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", user)
	if key != "" {
		req.Header.Set(idempotencyHeader, key)
	}

	resp, err := ta.app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return testResponse{
		status:   resp.StatusCode,
		body:     string(data),
		replayed: resp.Header.Get(idempotencyReplayedHeader),
		ctype:    resp.Header.Get(fiber.HeaderContentType),
		location: resp.Header.Get(fiber.HeaderLocation),
	}
}

func TestIdempotency_WithoutKeyAlwaysExecutes(t *testing.T) {
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), newFakeLocker())

	ta.post(t, "", "u1", `{"amount":"10.00"}`)
	ta.post(t, "", "u1", `{"amount":"10.00"}`)

	assert.Equal(t, int32(2), ta.executed.Load())
}

func TestIdempotency_ReplaysStoredResponseWithoutRedis(t *testing.T) {
	// Arrange
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), newFakeLocker())

	// Act
	first := ta.post(t, "key-1", "u1", `{"amount":"10.00"}`)
	second := ta.post(t, "key-1", "u1", `{"amount":"10.00"}`)

	// Assert
	assert.Equal(t, int32(1), ta.executed.Load())
	assert.Equal(t, fiber.StatusCreated, first.status)
	assert.Equal(t, first.status, second.status)
	assert.Equal(t, first.body, second.body)
	assert.Empty(t, first.replayed)
	assert.Equal(t, "true", second.replayed)
	assert.Equal(t, first.ctype, second.ctype)
}

func TestIdempotency_SameKeyDifferentBodyIsRejected(t *testing.T) {
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), newFakeLocker())

	ta.post(t, "key-1", "u1", `{"amount":"10.00"}`)
	reused := ta.post(t, "key-1", "u1", `{"amount":"99.00"}`)

	assert.Equal(t, fiber.StatusUnprocessableEntity, reused.status)
	assert.Equal(t, int32(1), ta.executed.Load())
}

func TestIdempotency_KeysAreScopedPerUser(t *testing.T) {
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), newFakeLocker())

	alice := ta.post(t, "shared-key", "alice", `{"amount":"10.00"}`)
	bob := ta.post(t, "shared-key", "bob", `{"amount":"10.00"}`)

	assert.Equal(t, int32(2), ta.executed.Load())
	assert.Empty(t, bob.replayed)
	assert.NotEqual(t, alice.body, bob.body)
}

func TestIdempotency_ConcurrentDuplicateGetsConflict(t *testing.T) {
	// Arrange: the first request blocks inside the handler
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), newFakeLocker())
	ta.block = make(chan struct{})
	ta.started = make(chan struct{}, 1)

	firstDone := make(chan testResponse, 1)
	go func() { firstDone <- ta.post(t, "key-1", "u1", `{"amount":"10.00"}`) }()
	<-ta.started

	// Act: a duplicate arrives while the first is in flight
	duplicate := ta.post(t, "key-1", "u1", `{"amount":"10.00"}`)
	close(ta.block)
	first := <-firstDone
	retry := ta.post(t, "key-1", "u1", `{"amount":"10.00"}`)

	// Assert
	assert.Equal(t, fiber.StatusConflict, duplicate.status)
	assert.Equal(t, fiber.StatusCreated, first.status)
	assert.Equal(t, "true", retry.replayed, "after completion the stored response is replayed")
	assert.Equal(t, int32(1), ta.executed.Load())
}

func TestIdempotency_ManySimultaneousDuplicatesExecuteOnce(t *testing.T) {
	// Arrange
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), newFakeLocker())
	const requests = 20

	// Act
	var wg sync.WaitGroup
	statuses := make(chan int, requests)
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			statuses <- ta.post(t, "key-1", "u1", `{"amount":"10.00"}`).status
		}()
	}
	wg.Wait()
	close(statuses)

	// Assert: every response is the original, a replay, or a conflict
	assert.Equal(t, int32(1), ta.executed.Load())
	for status := range statuses {
		assert.Contains(t, []int{fiber.StatusCreated, fiber.StatusConflict}, status)
	}
}

func TestIdempotency_FailedResponsesAreNotStored(t *testing.T) {
	// Arrange
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), newFakeLocker())
	ta.status.Store(fiber.StatusBadRequest)

	// Act
	failed := ta.post(t, "key-1", "u1", `{"amount":"10.00"}`)
	ta.status.Store(fiber.StatusCreated)
	retried := ta.post(t, "key-1", "u1", `{"amount":"10.00"}`)

	// Assert
	assert.Equal(t, fiber.StatusBadRequest, failed.status)
	assert.Equal(t, fiber.StatusCreated, retried.status)
	assert.Empty(t, retried.replayed)
	assert.Equal(t, int32(2), ta.executed.Load())
}

func TestIdempotency_KeyTooLong(t *testing.T) {
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), newFakeLocker())

	resp := ta.post(t, strings.Repeat("k", idempotencyMaxKeyLength+1), "u1", `{}`)

	assert.Equal(t, fiber.StatusBadRequest, resp.status)
	assert.Equal(t, int32(0), ta.executed.Load())
}

func TestIdempotency_InfrastructureFailuresDoNotExecute(t *testing.T) {
	t.Run("storage", func(t *testing.T) {
		ta := newIdempotencyTestApp(failingStorage{pkgcache.NewMemoryStorage()}, newFakeLocker())

		resp := ta.post(t, "key-1", "u1", `{}`)

		assert.Equal(t, fiber.StatusInternalServerError, resp.status)
		assert.Equal(t, int32(0), ta.executed.Load())
	})

	t.Run("lock", func(t *testing.T) {
		locker := newFakeLocker()
		locker.err = errors.New("redis down")
		ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), locker)

		resp := ta.post(t, "key-1", "u1", `{}`)

		assert.Equal(t, fiber.StatusInternalServerError, resp.status)
		assert.Equal(t, int32(0), ta.executed.Load())
	})
}

func TestIdempotency_LockReleasedAfterRequest(t *testing.T) {
	locker := newFakeLocker()
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), locker)

	ta.post(t, "key-1", "u1", `{}`)

	locker.mu.Lock()
	defer locker.mu.Unlock()
	assert.Empty(t, locker.held)
}

// A client retrying a create must learn where the first attempt's resource lives. Replaying the
// body without the Location header would hand it a 201 that names nothing.
func TestIdempotency_ReplayKeepsTheLocationHeader(t *testing.T) {
	ta := newIdempotencyTestApp(pkgcache.NewMemoryStorage(), newFakeLocker())

	first := ta.post(t, "key-1", "u1", `{"amount":"10.00"}`)
	replay := ta.post(t, "key-1", "u1", `{"amount":"10.00"}`)

	require.Equal(t, "true", replay.replayed)
	assert.Equal(t, "/orders/1", first.location)
	assert.Equal(t, first.location, replay.location, "the replay must point at the resource the first request created")
	assert.Equal(t, int32(1), ta.executed.Load())
}
