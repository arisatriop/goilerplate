package bootstrap

import (
	"errors"
	"fmt"
	"goilerplate/config"
	bootstrap "goilerplate/internal/bootstrap/database"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/response"
	"log/slog"
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

// Init loads and validates the config, sets up logging, and opens the connections the app
// needs. It returns an error rather than exiting, so the caller decides how to report it and
// nothing opened before the failure is left open: a failure after Redis connects closes Redis.
func Init() (*App, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	log := logger.New(loggerOptions(cfg))

	// Fail fast before any connection is opened
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
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
	validator := response.NewValidator()

	redis, err := NewRedis(cfg)
	if err != nil {
		return nil, err
	}

	db, err := initializeDatabase(cfg, log)
	if err != nil {
		if redis != nil {
			_ = redis.Close() // startup is already failing with the more useful error
		}
		return nil, err
	}

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
	}, nil
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
func initializeDatabase(cfg *config.Config, log *slog.Logger) (*bootstrap.DB, error) {
	gdb, err := bootstrap.NewGorm(cfg, log)
	if err != nil {
		return nil, err
	}

	db := bootstrap.NewDB()
	db.GDB = gdb
	return db, nil
}

// LogStartupFailure reports why the process could not start, through the configured logger when
// it got that far and the default one otherwise. A configuration error lists every problem found,
// one per entry, rather than as a single newline-joined string.
func LogStartupFailure(err error) {
	var joined interface{ Unwrap() []error }
	if errors.As(err, &joined) {
		problems := make([]string, 0, len(joined.Unwrap()))
		for _, problem := range joined.Unwrap() {
			problems = append(problems, problem.Error())
		}
		// The joined error's text is the problems again, newline-separated; keep only what wraps it.
		summary := strings.TrimSuffix(err.Error(), ": "+joined.(error).Error())
		slog.Error("startup failed", "error", summary, "problems", problems)
		return
	}
	slog.Error("startup failed", "error", err)
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
