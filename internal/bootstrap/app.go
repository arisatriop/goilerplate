package bootstrap

import (
	"goilerplate/config"
	bootstrap "goilerplate/internal/bootstrap/database"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/response"
	"log/slog"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

// App holds only infrastructure dependencies (Clean Architecture compliant).
// Optional components are nil when disabled: GrpcServer (grpc.enabled), Redis (redis.enabled),
// TracerProvider and MeterProvider (otel.enabled).
//
// GrpcServer is set by the wire layer rather than here: its auth interceptor needs the token
// validator and session store, which do not exist until the infrastructure layer is wired.
type App struct {
	DB             *bootstrap.DB
	Log            *slog.Logger
	Redis          *redis.Client
	Config         *config.Config
	WebServer      *fiber.App
	GrpcServer     *grpc.Server
	Validator      *validator.Validate
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
}

func Init() *App {
	cfg := Load()
	log := logger.New(loggerOptions(cfg))

	// Fail fast before any connection is opened
	if err := cfg.Validate(); err != nil {
		log.Error("invalid configuration", "errors", strings.Split(err.Error(), "\n"))
		os.Exit(1)
	}
	for _, warning := range cfg.Warnings() {
		log.Warn("configuration", "warning", warning)
	}

	var tp *sdktrace.TracerProvider
	var mp *sdkmetric.MeterProvider
	if cfg.OTel.Enabled {
		var err error
		if tp, err = NewTracerProvider(cfg); err != nil {
			log.Error("failed to initialize tracer provider", "error", err)
		}
		if mp, err = NewMeterProvider(cfg); err != nil {
			log.Error("failed to initialize meter provider", "error", err)
		}
	}

	fiber := NewFiber(cfg)
	redis := NewRedis(cfg, log)
	validator := response.NewValidator()

	db := initializeDatabase(cfg, log)

	logComponents(cfg, log)

	return &App{
		Config:         cfg,
		Log:            log,
		WebServer:      fiber,
		DB:             db,
		Redis:          redis,
		Validator:      validator,
		TracerProvider: tp,
		MeterProvider:  mp,
	}
}

// logComponents prints which optional components are enabled for this run.
func logComponents(cfg *config.Config, log *slog.Logger) {
	log.Info("components",
		"database", "postgres",
		"redis", cfg.Redis.Enabled,
		"grpc", cfg.GRPC.Enabled,
		"otel", cfg.OTel.Enabled,
		"storage", strings.ToLower(cfg.FileSystem.Driver),
		"auth_cache", cfg.Auth.CacheMode(cfg.Redis.Enabled),
		"partner_routes", PartnerRoutesEnabled(cfg),
	)
}

// PartnerRoutesEnabled reports whether partner routes are registered: only when at least
// one partner API key is configured.
func PartnerRoutesEnabled(cfg *config.Config) bool {
	return len(cfg.Apikeys) > 0
}

// initializeDatabase opens the PostgreSQL connection pool.
func initializeDatabase(cfg *config.Config, log *slog.Logger) *bootstrap.DB {
	db := bootstrap.NewDB()
	db.GDB = bootstrap.NewGorm(cfg, log)

	return db
}

// loggerOptions translates this application's config into pkg/logger's own options, so the
// logger package does not have to know about config.Config.
func loggerOptions(cfg *config.Config) logger.Options {
	if cfg == nil || cfg.Log == nil {
		return logger.Options{}
	}
	return logger.Options{
		Level:        cfg.Log.Level,
		AddSource:    cfg.Log.Source,
		RedactFields: cfg.Log.RedactFields,
	}
}
