package routes

import (
	"goilerplate/config"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RouteGroup struct {
	Public *gin.RouterGroup
}

func InitRoute(gin *gin.Engine) {
	con := config.GetDBConnection()
	validator := validator.New()

	example(con, validator)
}

func example(con *config.Con, validator *validator.Validate) {
	// request := api.NewExampleRequest()
	// response := api.NewExampleResponse()
	// repository := repository.NewExampleRepository(db)
	// usecase := usecase.NewExampleUsecase(validator, repository)
	// handler := NewExampleHandler(request, response, usecase)

	// e.POST("api/examples", handler.Create())
	// e.PUT("api/examples/:param", handler.Update())
	// e.DELETE("api/examples/", handler.Delete())
	// e.GET("api/examples/:param", handler.Find())
	// e.GET("api/examples", handler.FindAll())
}
