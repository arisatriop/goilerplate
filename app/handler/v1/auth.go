package v1

import (
	"fmt"
	"goilerplate/api/request"
	"goilerplate/api/response"
	"goilerplate/app/logging"
	usecase "goilerplate/app/usecase/v1"
	"goilerplate/config"

	"github.com/gofiber/fiber/v2"
)

func NewAuthHandler(app *config.App, request request.IAuth, response response.IAuth, usecase usecase.IAuth) IAuth {
	return &AuthImpl{
		App:      app,
		Request:  request,
		Response: response,
		Usecase:  usecase,
	}
}

type AuthImpl struct {
	App      *config.App
	Request  request.IAuth
	Response response.IAuth
	Usecase  usecase.IAuth
}

type IAuth interface {
	GenerateToken() fiber.Handler
}

func (h *AuthImpl) GenerateToken() fiber.Handler {
	return func(c *fiber.Ctx) error {

		errLog := logging.NewErrorLog()

		contentType := c.Get("Content-Type")
		if contentType != "application/json" {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
				"code":    4151,
				"result":  false,
				"message": "Unsupported Media Type. Content-Type must be application/json",
				"data":    nil,
			})
		}

		payload, err := h.Request.GetCreatePayload(c)
		if err != nil {
			errLog.Store(c, err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		fmt.Println(payload)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code":    2001,
			"result":  true,
			"message": "Success",
			"data":    nil,
		})
	}
}
