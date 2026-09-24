package wire

import (
	"context"
	"fmt"

	"goilerplate/config"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/lock"
	"goilerplate/internal/infrastructure/cache"
	pkgcache "goilerplate/pkg/cache"
	"goilerplate/pkg/filesystem"
	"goilerplate/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// Infrastructure contains all infrastructure dependencies
type Infrastructure struct {
	FilesystemManager *filesystem.Manager
	JWTService        *jwt.JWTService
	SessionStore      auth.SessionStore
	PermissionCache   auth.PermissionCache
	Locker            lock.Provider
	IdempotencyStore  fiber.Storage
	RateLimitStore    fiber.Storage
}

// WireInfrastructure creates all infrastructure dependencies
func WireInfrastructure(app *bootstrap.App) (*Infrastructure, error) {
	// Initialize filesystem manager from config
	filesystemMgr, err := filesystem.NewManagerFromConfig(context.Background(), filesystem.Config{
		Driver: filesystem.Driver(app.Config.FileSystem.Driver),
		Local:  app.Config.FileSystem.Local,
		S3:     app.Config.FileSystem.S3,
		Drive:  app.Config.FileSystem.Drive,
	})
	if err != nil {
		return nil, fmt.Errorf("initializing filesystem: %w", err)
	}

	// Initialize JWT service from config. config.Validate() has already run, so anything
	// rejected here is a bug in the mapping rather than bad user input.
	jwtService, err := newJWTService(app.Config)
	if err != nil {
		return nil, fmt.Errorf("initializing JWT service: %w", err)
	}

	sessionStore, permissionCache, locker, err := wireAuthCaches(app)
	if err != nil {
		return nil, err
	}

	return &Infrastructure{
		JWTService: jwtService,
		// nil without Redis, which the limiter reads as "count in memory".
		RateLimitStore:    pkgcache.NewFiberStorage(app.Redis, "rl:"),
		SessionStore:      sessionStore,
		PermissionCache:   permissionCache,
		Locker:            locker,
		IdempotencyStore:  wireIdempotencyStore(app),
		FilesystemManager: filesystemMgr,
	}, nil
}

// wireAuthCaches selects the session and permission cache implementations from
// auth.session_cache, and a lock provider that is shared across instances when Redis is on.
func wireAuthCaches(app *bootstrap.App) (auth.SessionStore, auth.PermissionCache, lock.Provider, error) {
	cfg := app.Config
	mode := cfg.Auth.CacheMode(cfg.Redis.Enabled)
	sessionTTL := cfg.Auth.SessionCacheTTLOrDefault()
	permissionTTL := cfg.Auth.PermissionCacheTTLOrDefault()

	var locker lock.Provider = cache.NewMemoryLocker()
	if app.Redis != nil {
		locker = cache.NewRedisLocker(app.Redis)
	}

	app.Log.Info("auth cache configured",
		"mode", mode,
		"session_cache_ttl", sessionTTL.String(),
		"permission_cache_ttl", permissionTTL.String(),
	)

	switch mode {
	case config.CacheModeRedis:
		// config.Validate refuses this combination, so reaching it means the two disagree.
		if app.Redis == nil {
			return nil, nil, nil, fmt.Errorf("auth.session_cache=%s requires redis.enabled=true", mode)
		}
		return cache.NewRedisSessionStore(app.Redis, sessionTTL), cache.NewRedisPermissionCache(app.Redis, permissionTTL), locker, nil
	case config.CacheModeMemory:
		app.Log.Warn("auth.session_cache=memory is per instance: with more than one instance, "+
			"revoked sessions and permission changes may take effect up to the cache TTL later on other instances",
			"session_cache_ttl", sessionTTL.String(),
			"permission_cache_ttl", permissionTTL.String(),
		)
		return cache.NewMemorySessionStore(sessionTTL), cache.NewMemoryPermissionCache(permissionTTL), locker, nil
	default:
		return cache.NoopSessionStore{}, cache.NoopPermissionCache{}, locker, nil
	}
}

// wireIdempotencyStore stores idempotent responses in Redis when enabled, otherwise in memory.
func wireIdempotencyStore(app *bootstrap.App) fiber.Storage {
	if app.Redis != nil {
		return pkgcache.NewFiberStorage(app.Redis, "idem:")
	}

	app.Log.Warn("redis is disabled: Idempotency-Key deduplication is per instance; " +
		"with more than one instance, duplicates that reach different instances are processed again")
	return pkgcache.NewMemoryStorage()
}

// newJWTService maps the JWT config onto the signing keys the service verifies with:
// one active key that signs, plus any retired keys that are still accepted.
func newJWTService(cfg *config.Config) (*jwt.JWTService, error) {
	previous := make([]jwt.Key, 0, len(cfg.JWT.PreviousKeys))
	for _, key := range cfg.JWT.PreviousKeys {
		previous = append(previous, jwt.Key{
			ID:            key.KeyID,
			AccessSecret:  key.AccessSecret,
			RefreshSecret: key.RefreshSecret,
		})
	}

	return jwt.NewJWTService(jwt.Config{
		Active: jwt.Key{
			ID:            cfg.JWT.KeyID,
			AccessSecret:  cfg.JWT.AccessSecret,
			RefreshSecret: cfg.JWT.RefreshSecret,
		},
		Previous:     previous,
		Issuer:       cfg.JWT.Issuer,
		Audience:     cfg.JWT.Audience,
		AccessExpiry: cfg.JWT.AccessTokenExpiry,
		Leeway:       cfg.JWT.Leeway,
	})
}
