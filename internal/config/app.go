package config

import (
	"golang-clean-architecture/internal/delivery/http/route"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type BootstrapConfig struct {
	DB       *DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
	Producer *kafka.Producer
}

func Bootstrap(config *BootstrapConfig) {
	// setup repositories

	// setup use cases

	// setup controller

	// setup middleware
	// authMiddleware := middleware.NewAuth(userUseCase)

	routeConfig := route.RouteConfig{
		App: config.App,
		// UserController:    userController,
		// ContactController: contactController,
		// AddressController: addressController,
		// AuthMiddleware:    authMiddleware,
	}
	routeConfig.Setup()
}
