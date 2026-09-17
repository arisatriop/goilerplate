package bootstrap

import (
	"goilerplate/config"
	bootstrap "goilerplate/internal/bootstrap/database"
	"goilerplate/pkg/logger"
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
	log := logger.NewSlog(cfg)

	// Fail fast before any connection is opened
	if err := cfg.Validate(); err != nil {
		log.Error("invalid configuration", "errors", strings.Split(err.Error(), "\n"))
		os.Exit(1)
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
	validator := validator.New()

	db := initializeDatabase(cfg, log)

	var grpcServer *grpc.Server
	if cfg.GRPC.Enabled {
		grpcServer = NewGrpcServer(cfg)
	}

	logComponents(cfg, log)

	return &App{
		Config:         cfg,
		Log:            log,
		WebServer:      fiber,
		GrpcServer:     grpcServer,
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

// initializeDatabase opens the PostgreSQL connections (GORM and pgx)
func initializeDatabase(cfg *config.Config, log *slog.Logger) *bootstrap.DB {
	db := bootstrap.NewDB()
	db.GDB = bootstrap.NewGorm(cfg, log)
	db.PgxDB = bootstrap.NewPostgres(cfg, log)

	return db
}
