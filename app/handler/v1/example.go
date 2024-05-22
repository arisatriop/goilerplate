package v1

import (
	"goilerplate/api/request"
	usecase "goilerplate/app/usecase/v1"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Res struct {
	Code    int
	Message string
}

type IExample interface {
	Create() fiber.Handler
	Update() fiber.Handler
	Delete() fiber.Handler
	FindAll() fiber.Handler
	FindById() fiber.Handler
}

type ExampleImpl struct {
	Validator *validator.Validate
	Request   request.IExample
	Usecase   usecase.IExample
}

func NewExampleHandler(validator *validator.Validate, request request.IExample, usecase usecase.IExample) IExample {
	return &ExampleImpl{
		Validator: validator,
		Request:   request,
		Usecase:   usecase,
	}
}

func (h *ExampleImpl) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {

		payload, err := h.Request.GetCreatePayload(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"status":  0,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		if err := h.Validator.Struct(payload); err != nil {
			var message string
			for _, err := range err.(validator.ValidationErrors) {
				message = err.Error()
				break // Print the default error message
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    4001,
				"status":  0,
				"message": message,
				"data":    nil,
			})
		}

		err = h.Usecase.Create(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"status":  0,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    2011,
			"status":  1,
			"message": "Sucess",
			"data":    nil,
		})
	}
}

func (h *ExampleImpl) Update() fiber.Handler {
	// panic("Not implement")
	return nil
}

func (h *ExampleImpl) Delete() fiber.Handler {
	// panic("Not implement")
	return nil
}

func (h *ExampleImpl) FindAll() fiber.Handler {
	// panic("Not implement")
	return nil
}

func (h *ExampleImpl) FindById() fiber.Handler {
	// panic("Not implement")
	return nil
}
