package v1

import (
	"fmt"
	"goilerplate/api/request"
	"goilerplate/api/response"
	"goilerplate/app/logging"
	usecase "goilerplate/app/usecase/v1"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
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
	Response  response.IExample
	Usecase   usecase.IExample
}

func NewExampleHandler(validator *validator.Validate, request request.IExample, response response.IExample, usecase usecase.IExample) IExample {
	return &ExampleImpl{
		Validator: validator,
		Request:   request,
		Response:  response,
		Usecase:   usecase,
	}
}

func (h *ExampleImpl) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {

		errLog := logging.ErrorLog{}

		payload, err := h.Request.GetCreatePayload(c)
		if err != nil {
			go errLog.Store(c, err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		if err := h.Validator.Struct(payload); err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				go errLog.Store(c, err.Error())
				fmt.Println("validation configuration error: ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    5001,
					"result":  false,
					"message": "Whops, something went wrong. Please try again in a moment",
					"data":    nil,
				})
			}

			var message string
			for _, err := range err.(validator.ValidationErrors) {
				message = fmt.Sprintf("fields %s is %s", err.Field(), err.ActualTag())
				break // Print the default error message
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    4001,
				"result":  false,
				"message": message,
				"data":    nil,
			})
		}

		err = h.Usecase.Create(c)
		if err != nil {
			go errLog.Store(c, err.Error())
			fmt.Println("ERROR: handler (create example): ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"code":    2011,
			"result":  true,
			"message": "Sucess",
			"data":    nil,
		})
	}
}

func (h *ExampleImpl) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {

		payload, err := h.Request.GetUpdatePayload(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		if err := h.Validator.Struct(payload); err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				fmt.Println("validation configuration error: ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    5001,
					"result":  false,
					"message": "Whops, something went wrong. Please try again in a moment",
					"data":    nil,
				})
			}

			var message string
			for _, err := range err.(validator.ValidationErrors) {
				message = fmt.Sprintf("fields %s is %s", err.Field(), err.ActualTag())
				break // Print the default error message
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    4001,
				"result":  false,
				"message": message,
				"data":    nil,
			})
		}

		err = h.Usecase.Update(c)
		if err != nil {
			if err == pgx.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"code":    4041,
					"result":  false,
					"message": "Data not found",
					"data":    nil,
				})
			}

			fmt.Println("ERROR: handler (update example): ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code":    2001,
			"result":  true,
			"message": "Sucess",
			"data":    nil,
		})

	}
}

func (h *ExampleImpl) Delete() fiber.Handler {
	return func(c *fiber.Ctx) error {

		err := h.Usecase.Delete(c)
		if err != nil {
			if err == pgx.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"code":    4041,
					"result":  false,
					"message": "Data not found",
					"data":    nil,
				})
			}

			fmt.Println("ERROR: handler (delete example): ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code":    2001,
			"result":  true,
			"message": "Sucess",
			"data":    nil,
		})
	}
}

func (h *ExampleImpl) FindAll() fiber.Handler {
	return func(c *fiber.Ctx) error {

		examples, err := h.Usecase.FindAll(c)
		if err != nil {
			fmt.Println("ERROR: handler (find all example): ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    []string{},
			})
		}

		response, err := h.Response.FindAll(examples)
		if err != nil {
			fmt.Println("ERROR: handler (find all example): ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    []string{},
			})
		}

		var data interface{}
		data = response
		if len(response) == 0 {
			data = []string{}
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code":    2001,
			"result":  true,
			"message": "Sucess",
			"data":    data,
		})
	}
}

func (h *ExampleImpl) FindById() fiber.Handler {
	return func(c *fiber.Ctx) error {

		example, err := h.Usecase.FindById(c)
		if err != nil {
			if err == pgx.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"code":    4041,
					"result":  false,
					"message": "Data not found",
					"data":    nil,
				})
			}

			fmt.Println("ERROR: handler (find by id example): ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		response, err := h.Response.FindById(example)
		if err != nil {
			fmt.Println("ERROR: handler (find by id example): ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    5001,
				"result":  false,
				"message": "Whops, something went wrong. Please try again in a moment",
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code":    2001,
			"result":  true,
			"message": "Sucess",
			"data":    response,
		})
	}
}
