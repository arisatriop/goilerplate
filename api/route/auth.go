package route

import (
	"goilerplate/api/request"
	"goilerplate/api/response"
	handler "goilerplate/app/handler/v1"
	usecase "goilerplate/app/usecase/v1"
	"goilerplate/config"

	"github.com/gofiber/fiber/v2"
)

func Auth(prefix string, r fiber.Router, app *config.App) {
	request := request.NewAuthRequest()
	response := response.NewAuthResponse()
	usecase := usecase.NewAuthUsecase(app)
	handler := handler.NewAuthHandler(app, request, response, usecase)

	auth := r.Group(prefix)
	auth.Post("/", handler.GenerateToken())
}
