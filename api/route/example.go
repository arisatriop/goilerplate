package route

import (
	"goilerplate/api/request"
	"goilerplate/api/response"
	"goilerplate/config"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	handler "goilerplate/app/handler/v1"
	repository "goilerplate/app/repository/v1"
	repositoryV2 "goilerplate/app/repository/v2"
	usecase "goilerplate/app/usecase/v1"
	usecaseV2 "goilerplate/app/usecase/v2"
)

func Example(prefix string, r fiber.Router, con *config.Con, validator *validator.Validate) {

	request := request.NewExampleRequest()
	response := response.NewExampleResponse()
	repository := repository.NewExampleRepository()
	usecase := usecase.NewExampleUsecase(con, repository)
	handler := handler.NewExampleHandler(validator, request, response, usecase)

	example := r.Group(prefix)
	example.Post("/", handler.Create())
	example.Put("/:id", handler.Update())
	example.Delete("/:id", handler.Delete())
	example.Get("", handler.FindAll())
	example.Get("/:id", handler.FindById())
}

func ExampleV2(prefix string, r fiber.Router, con *config.Con, validator *validator.Validate) {

	request := request.NewExampleRequest()
	response := response.NewExampleResponse()
	repository := repositoryV2.NewExampleRepository()
	usecase := usecaseV2.NewExampleUsecase(con, repository)
	handler := handler.NewExampleHandler(validator, request, response, usecase)

	example := r.Group(prefix)
	example.Post("/", handler.Create())
	example.Put("/:id", handler.Update())
	example.Delete("/:id", handler.Delete())
	example.Get("", handler.FindAll())
	example.Get("/:id", handler.FindById())
}
