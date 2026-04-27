package bootstrap

import (
	"goilerplate/config"
	bootstrap "goilerplate/internal/bootstrap/database"
	"goilerplate/pkg/logger"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

// App holds only infrastructure dependencies (Clean Architecture compliant)
type App struct {
	DB             *bootstrap.DB
	Log            *slog.Logger
	Redis          *redis.Client
	Config         *config.Config
	WebServer      *fiber.App
	GrpcServer     *grpc.Server
	Validator      *validator.Validate
	TracerProvider *sdktrace.TracerProvider
}

func Init() *App {
	cfg := Load()
	log := logger.NewSlog(cfg)

	tp, err := NewTracerProvider(cfg)
	if err != nil {
		log.Error("failed to initialize tracer provider", "error", err)
	}

	fiber := NewFiber(cfg)
	redis := NewRedis(cfg, log)
	validator := validator.New()

	db := initializeDatabase(cfg, log)

	return &App{
		Config:         cfg,
		Log:            log,
		WebServer:      fiber,
		GrpcServer:     NewGrpcServer(cfg),
		DB:             db,
		Redis:          redis,
		Validator:      validator,
		TracerProvider: tp,
	}
}

// initializeDatabase sets up your multi-database configuration
func initializeDatabase(cfg *config.Config, log *slog.Logger) *bootstrap.DB {
	db := bootstrap.NewDB()
	db.GDB = bootstrap.NewGorm(cfg, log)

	switch strings.ToLower(cfg.DB.Driver) {
	case bootstrap.Postgres:
		db.PgxDB = bootstrap.NewPostgres(cfg, log)
	case bootstrap.Mysql:
		db.MysqlDB = bootstrap.NewMysql(cfg, log)
	}

	return db
}
