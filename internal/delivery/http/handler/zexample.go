package handler

import (
	"errors"
	"fmt"
	"goilerplate/internal/model"
	"goilerplate/internal/model/zexample"
	"goilerplate/internal/usecase"
	"goilerplate/pkg/helper"
	"goilerplate/pkg/middleware"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ExampleHandler interface {
	Get(ctx *fiber.Ctx) error
	GetAll(ctx *fiber.Ctx) error
	Create(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
}

type exampleHandler struct {
	Log            *logrus.Logger
	Validator      *validator.Validate
	ExampleUsecase usecase.ExampleUsecase
}

func NewExampleHandler(log *logrus.Logger, validator *validator.Validate, exampleUsecase usecase.ExampleUsecase) ExampleHandler {
	return &exampleHandler{
		Log:            log,
		Validator:      validator,
		ExampleUsecase: exampleUsecase,
	}
}

func (c *exampleHandler) Get(ctx *fiber.Ctx) error {

	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return model.BadRequest(ctx, "Invalid UUID format")
	}

	example, err := c.ExampleUsecase.Get(ctx.UserContext(), uuid)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	return model.OK(ctx, nil, example)
}

func (c *exampleHandler) GetAll(ctx *fiber.Ctx) error {

	params := zexample.GetParams()
	if err := ctx.QueryParser(params); err != nil {
		return model.BadRequest(ctx, "Malformed JSON payload")
	}

	examples, total, err := c.ExampleUsecase.GetAll(ctx.UserContext(), params)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	if total == 0 {
		examples = []zexample.GetAllResponse{}
	}

	return model.OK(ctx, nil, examples, model.NewPagination(params.Limit, params.Offset, int(total)))
}

func (c *exampleHandler) Create(ctx *fiber.Ctx) error {

	var request zexample.CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return model.BadRequest(ctx, "Malformed JSON payload")
	}

	if err := c.Validator.Struct(request); err != nil {
		errs := err.(validator.ValidationErrors)[0]
		return model.BadRequest(ctx, strings.ToLower(fmt.Sprintf("field '%s' is %s", errs.Field(), errs.Tag())))
	}

	request.CreatedBy = middleware.GetUser(ctx).ID

	if err := c.ExampleUsecase.Create(ctx.UserContext(), &request); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	return model.Created(ctx, "Example created successfully")
}

func (c *exampleHandler) Update(ctx *fiber.Ctx) error {
	uuid, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return model.BadRequest(ctx, "Invalid UUID format")
	}

	var request zexample.UpdateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return model.BadRequest(ctx, "Malformed JSON payload")
	}

	request.UpdatedBy = middleware.GetUser(ctx).ID

	if err := c.ExampleUsecase.Update(ctx.UserContext(), uuid, &request); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	return model.Created(ctx, "Example updated successfully")
}

func (c *exampleHandler) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return model.BadRequest(ctx, "Invalid UUID format")
	}

	request := zexample.DeleteRequest{
		DeletedBy: middleware.GetUser(ctx).ID,
	}

	if err := c.ExampleUsecase.Delete(ctx.UserContext(), uuid, &request); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	return model.NoContent(ctx)
}
