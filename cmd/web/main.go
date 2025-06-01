package main

import (
	"context"
	"fmt"
	"golang-clean-architecture/internal/config"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	viperConfig := config.NewViper()
	log := config.NewLogger(viperConfig)
	app := config.NewFiber(viperConfig)
	db := config.NewDatabase(viperConfig, log)
	validate := config.NewValidator(viperConfig)

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   viperConfig,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		webPort := viperConfig.GetInt("server.port")
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

func gracefulShutdown(ctx context.Context, app *fiber.App, db *config.DB, log *logrus.Logger) {
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

	log.Info(message)
}
