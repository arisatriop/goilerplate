package http

import (
	"errors"
	"fmt"
	"golang-clean-architecture/internal/delivery/http/middleware"
	"golang-clean-architecture/internal/helper"
	"golang-clean-architecture/internal/model/zexample"
	"golang-clean-architecture/internal/usecase"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type IExampleController interface {
	Get(ctx *fiber.Ctx) error
	Create(ctx *fiber.Ctx) error
	List(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
}

type ExampleController struct {
	Log            *logrus.Logger
	Validator      *validator.Validate
	ExampleUsecase usecase.IExampleUsecase
}

func NewExampleController(log *logrus.Logger, validator *validator.Validate, exampleUsecase usecase.IExampleUsecase) IExampleController {
	return &ExampleController{
		Log:            log,
		Validator:      validator,
		ExampleUsecase: exampleUsecase,
	}
}

func (c *ExampleController) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return helper.ResBadRequest(ctx, "Invalid UUID format")
	}

	example, err := c.ExampleUsecase.FindByID(ctx.UserContext(), uuid)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return helper.Res(ctx, cerr.Code, cerr.Message)
		}
		return helper.ResInternalServerError(ctx, "Failed to retrieve example")
	}

	return helper.ResOK(ctx, example, "Example retrieved successfully")
}

func (c *ExampleController) Create(ctx *fiber.Ctx) error {

	var request zexample.CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return helper.ResBadRequest(ctx, "Invalid request payload")
	}

	if err := c.Validator.Struct(request); err != nil {
		errs := err.(validator.ValidationErrors)[0]
		return helper.ResBadRequest(ctx, strings.ToLower(fmt.Sprintf("field '%s' is %s", errs.Field(), errs.Tag())))
	}

	request.CreatedBy = middleware.GetUser(ctx).ID

	if err := c.ExampleUsecase.Create(ctx.UserContext(), &request); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return helper.Res(ctx, cerr.Code, cerr.Message)
		}
		return helper.ResInternalServerError(ctx, "Failed to create example")
	}

	return helper.ResCreated(ctx, "Example created successfully")
}

func (c *ExampleController) Update(ctx *fiber.Ctx) error {
	uuid, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return helper.ResBadRequest(ctx, "Invalid UUID format")
	}

	var request zexample.UpdateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return helper.ResBadRequest(ctx, "Invalid request payload")
	}

	request.UpdatedBy = middleware.GetUser(ctx).ID

	if err := c.ExampleUsecase.Update(ctx.UserContext(), uuid, &request); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return helper.Res(ctx, cerr.Code, cerr.Message)
		}
		return helper.ResInternalServerError(ctx, "Failed to update example")
	}

	return helper.ResOK(ctx, nil, "Example updated successfully")
}

func (c *ExampleController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return helper.ResBadRequest(ctx, "Invalid UUID format")
	}

	request := zexample.DeleteRequest{
		DeletedBy: middleware.GetUser(ctx).ID,
	}

	if err := c.ExampleUsecase.Delete(ctx.UserContext(), uuid, &request); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return helper.Res(ctx, cerr.Code, cerr.Message)
		}
		return helper.ResInternalServerError(ctx, "Failed to delete example")
	}

	return helper.ResOK(ctx, nil, "Example deleted successfully")
}

func (c *ExampleController) List(ctx *fiber.Ctx) error {
	response := map[string]string{
		"message": "List of examples",
	}

	return ctx.JSON(response)
}
