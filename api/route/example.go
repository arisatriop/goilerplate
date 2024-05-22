package route

import (
	"goilerplate/api/request"
	"goilerplate/api/response"
	handler "goilerplate/app/handler/v1"
	repository "goilerplate/app/repository/v1"
	usecase "goilerplate/app/usecase/v1"
	"goilerplate/config"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// func Example(e *gin.Engine, con *config.Con, validator *validator.Validate) {
// 	// request := api.NewExampleRequest()
// 	// response := api.NewExampleResponse()
// 	// repository := repository.NewExampleRepository(db)
// 	// usecase := usecase.NewExampleUsecase(validator, repository)
// 	// handler := NewExampleHandler(request, response, usecase)

// 	// e.POST("api/examples", handler.Create())
// 	// e.PUT("api/examples/:param", handler.Update())
// 	// e.DELETE("api/examples/", handler.Delete())
// 	// e.GET("api/examples/:param", handler.Find())
// 	// e.GET("api/examples", handler.FindAll())
// }

func Example(prefix string, r fiber.Router, con *config.Con, validator *validator.Validate) {

	repository := repository.NewExampleRepository(con)
	usecase := usecase.NewExampleUsecase(repository)
	request := request.NewExampleRequest()
	response := response.NewExampleResponse()
	handler := handler.NewExampleHandler(validator, request, response, usecase)

	example := r.Group(prefix)
	example.Post("/", handler.Create())
	example.Put("/:id", handler.Update())
	example.Delete("/:id", handler.Delete())
	example.Get("", handler.FindAll())
	example.Get("/:id", handler.FindById())
}
