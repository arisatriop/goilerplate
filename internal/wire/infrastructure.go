package wire

import (
	"context"

	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/infrastructure/cache"
	"goilerplate/pkg/filesystem"
	"goilerplate/pkg/jwt"
)

// Infrastructure contains all infrastructure dependencies
type Infrastructure struct {
	FilesystemManager *filesystem.Manager
	JWTService        *jwt.JWTService
	AuthCacheService  *auth.CacheService
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

	// Initialize cache service (will be nil if Redis is disabled)
	authCacheService := auth.NewCacheService(app.Redis)

	cacheService := cache.NewRedisService(app.Redis)

	return &Infrastructure{
		JWTService:        jwtService,
		CacheService:      cacheService,
		AuthCacheService:  authCacheService,
		FilesystemManager: filesystemMgr,
		// Future infrastructure wiring:
		// EmailService: email.NewService(...),
		// SMSService:   sms.NewService(...),
	}
}
