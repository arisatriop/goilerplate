package http

import (
	"fmt"
	"golang-clean-architecture/internal/delivery/http/middleware"
	"golang-clean-architecture/internal/helper"
	"golang-clean-architecture/internal/model"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type IExampleController interface {
	Create(ctx *fiber.Ctx) error
	List(ctx *fiber.Ctx) error
	Get(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
}

type ExampleController struct {
	Log       *logrus.Logger
	Validator *validator.Validate
}

func NewExampleController(log *logrus.Logger, validator *validator.Validate) IExampleController {
	return &ExampleController{
		Log:       log,
		Validator: validator,
	}
}

func (c *ExampleController) Create(ctx *fiber.Ctx) error {

	auth := middleware.GetUser(ctx)

	var request model.ExampleCreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		// return helper.BadRequest(ctx, "Invalid request payload")
	}

	if err := c.Validator.Struct(request); err != nil {
		errs := err.(validator.ValidationErrors)[0]
		return helper.BadRequest(ctx, strings.ToLower(fmt.Sprintf("field '%s' is %s", errs.Field(), errs.Tag())))
	}

	request.CreatedBy = auth.ID.String()

	return helper.Created(ctx, "Example created successfully")
}

func (c *ExampleController) List(ctx *fiber.Ctx) error {
	response := map[string]string{
		"message": "List of examples",
	}

	return ctx.JSON(response)
}

func (c *ExampleController) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	response := map[string]string{
		"message": "Example retrieved successfully",
		"id":      id,
	}

	return ctx.JSON(response)
}

func (c *ExampleController) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	response := map[string]string{
		"message": "Example updated successfully",
		"id":      id,
	}

	return ctx.JSON(response)
}

func (c *ExampleController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	response := map[string]string{
		"message": "Example deleted successfully",
		"id":      id,
	}

	return ctx.JSON(response)
}
