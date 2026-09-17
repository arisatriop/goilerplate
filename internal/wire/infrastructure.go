package wire

import (
	"context"
	"fmt"

	"goilerplate/config"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/lock"
	"goilerplate/internal/infrastructure/cache"
	"goilerplate/pkg/filesystem"
	"goilerplate/pkg/jwt"
)

// Infrastructure contains all infrastructure dependencies
type Infrastructure struct {
	FilesystemManager *filesystem.Manager
	JWTService        *jwt.JWTService
	SessionStore      auth.SessionStore
	PermissionCache   auth.PermissionCache
	Locker            lock.Provider
	CacheService      *cache.RedisService
	// Future infrastructure dependencies:
	// EmailService    email.Service
	// SMSService      sms.Service
}

// WireInfrastructure creates all infrastructure dependencies
func WireInfrastructure(app *bootstrap.App) *Infrastructure {
	// Initialize filesystem manager from config
	filesystemMgr, err := filesystem.NewManagerFromConfig(context.Background(), filesystem.Config{
		Driver: filesystem.Driver(app.Config.FileSystem.Driver),
		Local:  app.Config.FileSystem.Local,
		S3:     app.Config.FileSystem.S3,
		Drive:  app.Config.FileSystem.Drive,
	})
	if err != nil {
		panic("Failed to initialize filesystem manager: " + err.Error())
	}

	// Initialize JWT service from config
	jwtService := jwt.NewJWTService(
		app.Config.JWT.SecretKey,
		app.Config.JWT.AccessSecret,
		app.Config.JWT.RefreshSecret,
		app.Config.JWT.Issuer,
		app.Config.JWT.AccessTokenExpiry,
		app.Config.JWT.RefreshTokenExpiry,
	)

	cacheService := cache.NewRedisService(app.Redis)
	sessionStore, permissionCache, locker := wireAuthCaches(app)

	return &Infrastructure{
		JWTService:        jwtService,
		CacheService:      cacheService,
		SessionStore:      sessionStore,
		PermissionCache:   permissionCache,
		Locker:            locker,
		FilesystemManager: filesystemMgr,
		// Future infrastructure wiring:
		// EmailService: email.NewService(...),
		// SMSService:   sms.NewService(...),
	}
}

// wireAuthCaches selects the session and permission cache implementations from
// auth.session_cache, and a lock provider that is shared across instances when Redis is on.
func wireAuthCaches(app *bootstrap.App) (auth.SessionStore, auth.PermissionCache, lock.Provider) {
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
		if app.Redis == nil {
			panic(fmt.Sprintf("auth.session_cache=%s requires redis.enabled=true", mode))
		}
		return cache.NewRedisSessionStore(app.Redis, sessionTTL), cache.NewRedisPermissionCache(app.Redis, permissionTTL), locker
	case config.CacheModeMemory:
		app.Log.Warn("auth.session_cache=memory is per instance: with more than one instance, "+
			"revoked sessions and permission changes may take effect up to the cache TTL later on other instances",
			"session_cache_ttl", sessionTTL.String(),
			"permission_cache_ttl", permissionTTL.String(),
		)
		return cache.NewMemorySessionStore(sessionTTL), cache.NewMemoryPermissionCache(permissionTTL), locker
	default:
		return cache.NoopSessionStore{}, cache.NoopPermissionCache{}, locker
	}
}
