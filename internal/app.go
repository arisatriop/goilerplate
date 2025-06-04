package internal

import (
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/delivery/http"
	"golang-clean-architecture/internal/delivery/http/middleware"
	"golang-clean-architecture/internal/delivery/http/route"
	"golang-clean-architecture/internal/repository"
	"golang-clean-architecture/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type BootstrapConfig struct {
	DB       *config.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func Bootstrap(cfg *BootstrapConfig) {
	// setup repositories
	exampleRepository := repository.NewExampleRepository(cfg.Log)

	// setup use cases
	exampleUsecase := usecase.NewExampleUsecase(cfg.Log, cfg.DB, exampleRepository)

	// setup controller
	exampleController := http.NewExampleController(cfg.Log, cfg.Validate, exampleUsecase)
	// setup middleware
	auth := middleware.NewAuth()

	routeConfig := route.RouteConfig{
		App:  cfg.App,
		Auth: auth,

		ExampleController: exampleController,
	}
	routeConfig.Setup()
}
