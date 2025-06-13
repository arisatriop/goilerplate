package internal

import (
	"goilerplate/internal/config"
	"goilerplate/internal/delivery/http"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/delivery/http/route"
	"goilerplate/internal/repository"
	"goilerplate/internal/usecase"

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
	menuRepository := repository.NewMenuRepository(cfg.Log)
	exampleRepository := repository.NewExampleRepository(cfg.Log)
	userRepository := repository.NewUserRepository(cfg.Log)
	permissionRepository := repository.NewPermissionRepository(cfg.Log)
	menuPermRepo := repository.NewMenuPermissionRepository(cfg.Log)
	roleRepository := repository.NewRoleRepository(cfg.Log)
	menuPermissionRoleRepository := repository.NewMenuPermissionRoleRepository(cfg.Log)

	// setup use cases
	menuUsecase := usecase.NewMenuUsecase(cfg.Log, cfg.DB, menuRepository)
	authUsecase := usecase.NewAuthUsecase(cfg.Config, cfg.Log, cfg.DB, menuUsecase, userRepository, roleRepository, menuRepository, permissionRepository, redisRepository)
	exampleUsecase := usecase.NewExampleUsecase(cfg.Log, cfg.DB, exampleRepository)
	userUsecase := usecase.NewUserUsecase(cfg.Log, cfg.DB, userRepository)
	roleUsecase := usecase.NewRoleUseCase(cfg.Log, cfg.DB, roleRepository, menuPermRepo, menuPermissionRoleRepository)
	menuPermUsecase := usecase.NewMenuPermissionUsecase(cfg.Log, cfg.DB, menuPermRepo)

	// setup controller
	exampleController := http.NewExampleController(cfg.Log, cfg.Validate, exampleUsecase)
	authController := http.NewAuthController(cfg.Log, cfg.Validate, authUsecase)
	userController := http.NewUserController(cfg.Log, cfg.Validate, userUsecase)
	menuController := http.NewMenuController(cfg.Log, cfg.Validate, menuUsecase)
	roleController := http.NewRoleController(cfg.Log, cfg.Validate, roleUsecase)
	menuPermController := http.NewMenuPermissionController(cfg.Log, cfg.Validate, menuPermUsecase)

	// setup middleware
	auth := middleware.NewAuth(cfg.Config, authUsecase, userUsecase)

	routeConfig := route.RouteConfig{
		App:  cfg.App,
		Auth: auth,

		ExampleController:  exampleController,
		AuthController:     authController,
		UserController:     userController,
		MenuController:     menuController,
		RoleController:     roleController,
		MenuPermController: menuPermController,
	}
	routeConfig.Setup()
}
