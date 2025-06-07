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
	redisRepository := repository.NewRedisRepository(cfg.DB.Redis)

	// setup repositories
	exampleRepository := repository.NewExampleRepository(cfg.Log)
	userRepository := repository.NewUserRepository(cfg.Log)
	permissionRepository := repository.NewPermissionRepository(cfg.Log)

	// setup use cases
	exampleUsecase := usecase.NewExampleUsecase(cfg.Log, cfg.DB, exampleRepository)
	authUsecase := usecase.NewAuthUsecase(cfg.Config, cfg.Log, cfg.DB, userRepository, permissionRepository, redisRepository)
	userUsecase := usecase.NewUserUsecase(cfg.Log, cfg.DB, userRepository)

	// setup controller
	exampleController := http.NewExampleController(cfg.Log, cfg.Validate, exampleUsecase)
	authController := http.NewAuthController(cfg.Log, cfg.Validate, authUsecase)
	userController := http.NewUserController(cfg.Log, cfg.Validate, userUsecase)
	// setup middleware
	auth := middleware.NewAuth(cfg.Config, authUsecase, userUsecase)

	routeConfig := route.RouteConfig{
		App:  cfg.App,
		Auth: auth,

		ExampleController: exampleController,
		AuthController:    authController,
		UserController:    userController,
	}
	routeConfig.Setup()
}
