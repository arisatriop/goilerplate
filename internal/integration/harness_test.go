// Package integration wires the real auth middleware, use case and PostgreSQL repository
// together and drives them over HTTP.
//
// Every layer here is already unit-tested in isolation. What is not covered anywhere else is
// the composition: revoking a session is a write in the repository, an eviction in the cache
// and a rejection in the middleware, and the property that matters — a revoked login stops
// working, and only that login stops working — is a claim about all three agreeing. A stub
// repository cannot make that claim, because its revocation semantics would be the ones the
// test author believed rather than the ones the SQL implements.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"goilerplate/config"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/infrastructure/cache"
	"goilerplate/internal/infrastructure/repository"
	infratx "goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/jwt"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/migration"
	"goilerplate/pkg/password"
	"goilerplate/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	migrationsDir = "../migrations"

	// testAccessExpiry is long enough that no test races the clock. Expiry itself is covered
	// by the jwt package's own tests; what these tests exercise is revocation, which must not
	// wait for a token to expire.
	testAccessExpiry  = 5 * time.Minute
	testSessionExpiry = time.Hour
	testReuseGrace    = 10 * time.Second
	testCacheTTL      = time.Minute
	testMaxAttempts   = 3
)

// openTestDB connects to POSTGRES_TEST_DSN and applies the migrations.
// Database tests are skipped when POSTGRES_TEST_DSN is not set.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set; skipping PostgreSQL integration test")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: utils.Now,
		Logger:  gormlogger.Discard,
	})
	require.NoError(t, err)
	require.NoError(t, migration.NewMigrator(db, nil).Up(context.Background(), migrationsDir))

	return db
}

// cacheMode names one auth.session_cache setting. The lifecycle suite runs under all three
// because the cache is where revocation can silently fail to take effect: with noop there is
// nothing to go stale, so a missing eviction is invisible, and only memory and redis can
// serve a session the database has already revoked.
type cacheMode struct {
	name  string
	build func(t *testing.T) (auth.SessionStore, auth.PermissionCache)
}

func cacheModes() []cacheMode {
	return []cacheMode{
		{
			name: "none",
			build: func(*testing.T) (auth.SessionStore, auth.PermissionCache) {
				return cache.NoopSessionStore{}, cache.NoopPermissionCache{}
			},
		},
		{
			name: "memory",
			build: func(*testing.T) (auth.SessionStore, auth.PermissionCache) {
				return cache.NewMemorySessionStore(testCacheTTL), cache.NewMemoryPermissionCache(testCacheTTL)
			},
		},
		{
			name: "redis",
			build: func(t *testing.T) (auth.SessionStore, auth.PermissionCache) {
				client := newTestRedis(t)
				return cache.NewRedisSessionStore(client, testCacheTTL), cache.NewRedisPermissionCache(client, testCacheTTL)
			},
		},
	}
}

// newTestRedis connects to REDIS_TEST_ADDR (database REDIS_TEST_DB, default 15) and flushes it
// after the test. Redis tests are skipped when REDIS_TEST_ADDR is not set.
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR not set; skipping Redis cache mode")
	}

	db := 15
	if value := os.Getenv("REDIS_TEST_DB"); value != "" {
		parsed, err := strconv.Atoi(value)
		require.NoError(t, err)
		db = parsed
	}

	client := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	require.NoError(t, client.Ping(context.Background()).Err())

	t.Cleanup(func() {
		_ = client.FlushDB(context.Background()).Err()
		_ = client.Close()
	})

	return client
}

// stack is the application under test: real middleware over the real use case over the real
// repository, reachable the way a client reaches it.
type stack struct {
	app  *fiber.App
	db   *gorm.DB
	repo auth.Repository
	uc   auth.Usecase
}

// newStack wires one application instance using the given cache mode.
//
// The routes mirror router/public.go — the same middleware in the same order — but call the
// use case directly instead of going through the handlers. The handlers are request decoding
// and response shaping; the behaviour under test is everything behind them.
func newStack(t *testing.T, mode cacheMode) *stack {
	t.Helper()

	db := openTestDB(t)
	repo := repository.NewAuth(db)
	sessionStore, permissionCache := mode.build(t)

	jwtService, err := jwt.NewJWTService(jwt.Config{
		Active: jwt.Key{
			ID:            "test-key",
			AccessSecret:  "integration-test-access-secret-not-a-real-one",
			RefreshSecret: "integration-test-refresh-secret-not-a-real-one",
		},
		Issuer:       "goilerplate-test",
		Audience:     "goilerplate-test",
		AccessExpiry: testAccessExpiry,
	})
	require.NoError(t, err)

	// checkEveryRequest: true is auth.revocation=strict. These tests assert that a revoked
	// session stops working immediately, which is what strict promises; refresh_only trades
	// that for one fewer lookup per request and is covered by the session service's own tests.
	sessionService := auth.NewSessionService(repo, sessionStore, true)
	permissionService := auth.NewPermissionService(repo, permissionCache)

	useCase := auth.NewUseCase(
		repo,
		jwtService,
		sessionService,
		permissionService,
		auth.SessionExpiry{Default: testSessionExpiry, RememberMe: 24 * time.Hour},
		testReuseGrace,
		infratx.NewGormTransaction(db),
		auth.Lockout{MaxAttempts: testMaxAttempts, Duration: time.Minute},
		password.NewPolicy(nil),
	)

	authMiddleware := middleware.NewAuth(jwtService, repo, sessionService, permissionService, nil, config.InternalAuth{})

	return &stack{
		app:  newApp(useCase, authMiddleware),
		db:   db,
		repo: repo,
		uc:   useCase,
	}
}

// loginRequest is the subset of the login body these tests send.
type loginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

// tokens is what a client keeps after a login or a refresh. UserID is not something a client
// would be handed; it is here so a test can act as another instance of the application.
type tokens struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Session string `json:"session_id"`
	UserID  string `json:"user_id"`
}

func newApp(useCase auth.Usecase, authMiddleware *middleware.Auth) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Post("/login", func(ctx *fiber.Ctx) error {
		var body loginRequest
		if err := ctx.BodyParser(&body); err != nil {
			return ctx.SendStatus(http.StatusBadRequest)
		}

		result, err := useCase.Login(ctx.UserContext(), &auth.LoginCredentials{
			Email:      body.Email,
			Password:   body.Password,
			RememberMe: body.RememberMe,
		}, testDevice())
		if err != nil {
			return statusFor(ctx, err)
		}

		return ctx.JSON(asTokens(result))
	})

	// The middleware is here because it is what a client actually goes through: it validates
	// the refresh token, checks the session, and hands the handler the jti. Calling the use
	// case directly would skip the first two and test a path no request takes.
	app.Post("/refresh", authMiddleware.AuthenticateRefreshToken(), func(ctx *fiber.Ctx) error {
		result, err := useCase.RefreshToken(
			ctx.UserContext(),
			ctx.Locals(string(constants.ContextKeyUserID)).(string),
			ctx.Locals(string(constants.ContextKeySessionID)).(string),
			ctx.Locals("refresh_jti").(string),
			testDevice(),
		)
		if err != nil {
			return statusFor(ctx, err)
		}

		return ctx.JSON(asTokens(result))
	})

	app.Post("/logout", authMiddleware.Authenticate(), func(ctx *fiber.Ctx) error {
		err := useCase.Logout(
			ctx.UserContext(),
			ctx.Locals(string(constants.ContextKeyUserID)).(string),
			ctx.Locals(string(constants.ContextKeySessionID)).(string),
		)
		if err != nil {
			return statusFor(ctx, err)
		}

		return ctx.SendStatus(http.StatusOK)
	})

	app.Post("/logout-all", authMiddleware.Authenticate(), func(ctx *fiber.Ctx) error {
		if err := useCase.LogoutAll(ctx.UserContext(), ctx.Locals(string(constants.ContextKeyUserID)).(string)); err != nil {
			return statusFor(ctx, err)
		}

		return ctx.SendStatus(http.StatusOK)
	})

	// Stands in for any authenticated endpoint: reaching it means the access token was
	// accepted and, in strict mode, the session behind it is still live.
	app.Get("/me", authMiddleware.Authenticate(), func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusOK)
	})

	return app
}

func asTokens(result *auth.LoginResult) tokens {
	return tokens{
		Access:  result.Tokens.AccessToken,
		Refresh: result.Tokens.RefreshToken,
		Session: result.Session.ID,
		UserID:  result.User.ID,
	}
}

// statusFor answers with the client error's own status and message.
//
// The message is passed through rather than flattened to a status code because the
// anti-enumeration tests compare whole responses: if the harness invented the body, they would
// be comparing two strings it had just made identical and would pass however the real code
// behaved.
func statusFor(ctx *fiber.Ctx, err error) error {
	var clientErr *utils.ClientError
	if errors.As(err, &clientErr) {
		return ctx.Status(clientErr.Code).JSON(fiber.Map{"message": clientErr.Error()})
	}

	return ctx.SendStatus(http.StatusInternalServerError)
}

func testDevice() *auth.DeviceInfo {
	return &auth.DeviceInfo{
		DeviceID:   "integration-test-device",
		DeviceType: auth.DeviceTypeWeb,
		DeviceName: "integration-test",
		IPAddress:  "203.0.113.7",
		UserAgent:  "integration-test-agent",
	}
}

// createUser registers a user through the use case, so the stored hash is produced by the same
// code login verifies against. The row is removed afterwards; sessions and tokens follow it
// through ON DELETE CASCADE.
func (s *stack) createUser(t *testing.T, password string) string {
	t.Helper()

	email := utils.GenerateUUID() + "@example.test"
	user := &auth.User{Name: "Integration test", Email: email}
	require.NoError(t, s.uc.Register(context.Background(), user, password))

	stored, err := s.repo.GetUserByEmail(context.Background(), email)
	require.NoError(t, err)
	require.NotNil(t, stored)

	t.Cleanup(func() { s.db.Exec("DELETE FROM users WHERE id = ?", stored.ID) })

	return email
}

// do sends one request and returns its status and decoded body. The body is only read when the
// caller asks for it, because most assertions here are about the status alone.
func (s *stack) do(t *testing.T, method, path, bearer string, body any) (int, []byte) {
	t.Helper()

	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		payload = bytes.NewReader(encoded)
	}

	req := httptest.NewRequest(method, path, payload)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	if bearer != "" {
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+bearer)
	}

	resp, err := s.app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	read, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, read
}

// login signs in and returns the tokens, failing the test if the login itself did not succeed.
func (s *stack) login(t *testing.T, email, password string) tokens {
	t.Helper()

	status, body := s.do(t, http.MethodPost, "/login", "", loginRequest{Email: email, Password: password})
	require.Equal(t, http.StatusOK, status, "login: %s", body)

	var issued tokens
	require.NoError(t, json.Unmarshal(body, &issued))
	require.NotEmpty(t, issued.Access)
	require.NotEmpty(t, issued.Refresh)

	return issued
}

// refresh exchanges a refresh token and returns the status alongside whatever was issued.
func (s *stack) refresh(t *testing.T, refreshToken string) (int, tokens) {
	t.Helper()

	status, body := s.do(t, http.MethodPost, "/refresh", refreshToken, nil)
	if status != http.StatusOK {
		return status, tokens{}
	}

	var issued tokens
	require.NoError(t, json.Unmarshal(body, &issued))

	return status, issued
}

// captureSecurityEvents returns only the security-labelled entries fn produced. Ordinary
// application logs are filtered out: an assertion that counted both would pass even if an
// event were downgraded to a plain log line.
func captureSecurityEvents(t *testing.T, fn func()) []map[string]any {
	t.Helper()

	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	fn()

	var events []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal(line, &entry))
		if entry["label"] == logger.SecurityLabel {
			events = append(events, entry)
		}
	}

	return events
}

func actionsOf(events []map[string]any) []string {
	actions := make([]string, 0, len(events))
	for _, event := range events {
		actions = append(actions, event["action"].(string))
	}

	return actions
}

// sessionRow reads a session straight from the database, bypassing both the cache and the
// repository, so an assertion about stored state cannot be satisfied by a cache entry.
func (s *stack) sessionRow(t *testing.T, sessionID string) (isActive bool, revokedReason string) {
	t.Helper()

	var row struct {
		IsActive      bool
		RevokedReason *string
	}
	err := s.db.Raw("SELECT is_active, revoked_reason FROM user_sessions WHERE id = ?", sessionID).Scan(&row).Error
	require.NoError(t, err)

	if row.RevokedReason != nil {
		revokedReason = *row.RevokedReason
	}

	return row.IsActive, revokedReason
}
