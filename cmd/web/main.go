package main

import (
	"fmt"
	"golang-clean-architecture/internal/config"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	viperConfig := config.NewViper()
	log := config.NewLogger(viperConfig)
	app := config.NewFiber(viperConfig)
	db := config.NewDatabase(viperConfig, log)
	validate := config.NewValidator(viperConfig)
	// producer := config.NewKafkaProducer(viperConfig, log)

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   viperConfig,
		// Producer: producer,
	})

	done := make(chan string)
	go waitForShutdown(app, db, log, done)

	webPort := viperConfig.GetInt("server.port")
	if err := app.Listen(fmt.Sprintf(":%d", webPort)); err != nil {
		select {
		case <-done:
		default:
			done <- "App failed to start"
		}
	}

	log.Info(<-done)
}

func waitForShutdown(app *fiber.App, db *config.DB, log *logrus.Logger, done chan string) {
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, syscall.SIGINT, syscall.SIGTERM)
	signame := <-quitSignal

	fmt.Println()
	log.Infof("Shutdown signal received: %s", signame.String())
	log.Info("Cleaning up resources...")

	msg := "App shutdown gracefully..."

	if err := app.Shutdown(); err != nil {
		msg = "App force to shutdown due to fiber app stop failed"
	}
	if db.SqlDB != nil {
		if err := db.SqlDB.Close(); err != nil {
			msg = "App force to shutdown due to database connection close failed"
		}
	}
	if db.PgxPool != nil {
		db.PgxPool.Close()
	}
	if db.GDB != nil {
		sqlDB, err := db.GDB.DB()
		if err != nil {
			msg = "App force to shutdown due to gorm database connection close failed"
		} else {
			sqlDB.Close()
		}
	}

	select {
	case done <- msg:
	default:
	}
}
