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
)

// App holds only infrastructure dependencies (Clean Architecture compliant)
type App struct {
	DB        *bootstrap.DB
	Log       *slog.Logger
	Redis     *redis.Client
	Config    *config.Config
	WebServer *fiber.App
	Validator *validator.Validate
}

func Init() *App {
	cfg := Load()
	log := logger.NewSlog(cfg)
	fiber := NewFiber(cfg)
	redis := NewRedis(cfg, log)
	validator := validator.New()

	db := initializeDatabase(cfg, log)

	return &App{
		Config:    cfg,
		Log:       log,
		WebServer: fiber,
		DB:        db,
		Redis:     redis,
		Validator: validator,
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
