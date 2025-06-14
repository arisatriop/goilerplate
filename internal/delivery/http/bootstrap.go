package http

import (
	"goilerplate/config"
	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/repository"
	"goilerplate/internal/usecase"
	"goilerplate/pkg/db"
	"goilerplate/pkg/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// import (
// 	"goilerplate/config"
// 	"goilerplate/internal/delivery/http/middleware"
// 	"goilerplate/internal/repository"
// 	"goilerplate/internal/usecase"

// 	"github.com/go-playground/validator/v10"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/sirupsen/logrus"
// 	"github.com/spf13/viper"
// )

type Bootstrap struct {
	DB       *db.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *config.Config
}

func Boot(b *Bootstrap) {
	redisRepository := repository.NewRedisRepository(b.DB.Redis)

	// setup repositories
	menuRepository := repository.NewMenuRepository(b.Log)
	exampleRepository := repository.NewExampleRepository(b.Log)
	userRepository := repository.NewUserRepository(b.Log)
	permissionRepository := repository.NewPermissionRepository(b.Log)
	menuPermRepo := repository.NewMenuPermissionRepository(b.Log)
	roleRepository := repository.NewRoleRepository(b.Log)
	menuPermissionRoleRepository := repository.NewMenuPermissionRoleRepository(b.Log)

	// setup use cases
	menuUsecase := usecase.NewMenuUsecase(b.Log, b.DB, menuRepository)
	authUsecase := usecase.NewAuthUsecase(b.Config, b.Log, b.DB, menuUsecase, userRepository, roleRepository, menuRepository, permissionRepository, redisRepository)
	exampleUsecase := usecase.NewExampleUsecase(b.Log, b.DB, exampleRepository)
	userUsecase := usecase.NewUserUsecase(b.Log, b.DB, userRepository)
	roleUsecase := usecase.NewRoleUseCase(b.Log, b.DB, roleRepository, menuPermRepo, menuPermissionRoleRepository)
	menuPermUsecase := usecase.NewMenuPermissionUsecase(b.Log, b.DB, menuPermRepo)

	// 	// setup Handler
	exampleHandler := handler.NewExampleHandler(b.Log, b.Validate, exampleUsecase)
	authHandler := handler.NewAuthHandler(b.Log, b.Validate, authUsecase)
	userHandler := handler.NewUserHandler(b.Log, b.Validate, userUsecase)
	menuHandler := handler.NewMenuHandler(b.Log, b.Validate, menuUsecase)
	roleHandler := handler.NewRoleHandler(b.Log, b.Validate, roleUsecase)
	menuPermHandler := handler.NewMenuPermissionHandler(b.Log, b.Validate, menuPermUsecase)

	// 	// setup middleware
	auth := middleware.NewAuth(b.Config, authUsecase, userUsecase)

	route := Route{
		App:  b.App,
		Auth: auth,

		ExampleHandler:  exampleHandler,
		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		MenuHandler:     menuHandler,
		RoleHandler:     roleHandler,
		MenuPermHandler: menuPermHandler,
	}

	route.Setup()
}
