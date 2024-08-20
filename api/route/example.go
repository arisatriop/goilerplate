package route

import (
	"goilerplate/api/request"
	"goilerplate/api/response"
	"goilerplate/config"

	"github.com/gofiber/fiber/v2"

	handler "goilerplate/app/handler/v1"
	handlerV2 "goilerplate/app/handler/v2"
	repository "goilerplate/app/repository/v1"
	repositoryV2 "goilerplate/app/repository/v2"
	usecase "goilerplate/app/usecase/v1"
	usecaseV2 "goilerplate/app/usecase/v2"
)

func Example(prefix string, r fiber.Router, app *config.App) {

	request := request.NewExampleRequest()
	response := response.NewExampleResponse()
	repository := repository.NewExampleRepository()
	usecase := usecase.NewExampleUsecase(app, repository)
	handler := handler.NewExampleHandler(app, request, response, usecase)

	example := r.Group(prefix)
	example.Post("/", handler.Create())
	example.Put("/:id", handler.Update())
	example.Delete("/:id", handler.Delete())
	example.Get("", handler.FindAll())
	example.Get("/:id", handler.FindById())
}

func ExampleV2(prefix string, r fiber.Router, app *config.App) {

	request := request.NewExampleRequest()
	response := response.NewExampleResponse()
	repository := repositoryV2.NewExampleRepository()
	usecase := usecaseV2.NewExampleUsecase(app, repository)
	handler := handlerV2.NewExampleHandler(app, request, response, usecase)

	example := r.Group(prefix)
	example.Post("/", handler.Create())
	example.Put("/:id", handler.Update())
	example.Delete("/:id", handler.Delete())
	example.Get("", handler.FindAll())
	example.Get("/:id", handler.FindById())
}
