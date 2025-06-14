package main

import (
	"context"
	"fmt"
	"goilerplate/config"
	"goilerplate/internal/delivery/http"
	"goilerplate/pkg"
	"goilerplate/pkg/db"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	validator := pkg.NewValidator()
	vpr := config.NewViper()
	cfg := config.Load(vpr)
	log := pkg.NewLogger(cfg)
	app := pkg.NewFiber(cfg)
	db := db.NewDatabase(cfg, log)

	http.Boot(&http.Bootstrap{
		DB:       db,
		App:      app,
		Log:      log,
		Config:   cfg,
		Validate: validator,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		webPort := cfg.Server.Port
		if err := app.Listen(fmt.Sprintf(":%d", webPort)); err != nil {
			log.Error("Failed to start the server: ", err)
			stop()
		}
	}()

	<-ctx.Done()

	fmt.Println()
	log.Info("Shutting down server...")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gracefulShutdown(timeoutCtx, app, db, log)
}

func gracefulShutdown(ctx context.Context, app *fiber.App, db *db.DB, log *logrus.Logger) {
	message := "Server shutting down gracefully..."

	log.Info("Cleaning up resources...")

	done := make(chan error, 1)
	go func() {
		done <- app.Shutdown()
	}()

	select {
	case err := <-done:
		if err != nil {
			message = fmt.Sprintf("Server forced to shutdown: %v", err)
		}
	case <-ctx.Done():
		message = "Server forced to shutdown: timeout expired during fiber shutdown"
	}

	if db.SqlDB != nil {
		if err := db.SqlDB.Close(); err != nil {
			message = fmt.Sprintf("Server forced to shutdown: %v", err)
		}
	}
	if db.PgxPool != nil {
		db.PgxPool.Close()
	}
	if db.GDB != nil {
		if sqlDB, err := db.GDB.DB(); err != nil {
			message = fmt.Sprintf("Server forced to shutdown: %v", err)
		} else {
			sqlDB.Close()
		}
	}
	if db.Redis != nil {
		if err := db.Redis.Close(); err != nil {
			message = fmt.Sprintf("Server forced to shutdown: %v", err)
		}
	}

	log.Info(message)
}
