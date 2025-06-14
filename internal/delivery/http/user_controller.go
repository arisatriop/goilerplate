package http

import (
	"errors"
	"fmt"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/model"
	"goilerplate/internal/model/user"
	"goilerplate/internal/usecase"
	"goilerplate/pkg/helper"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserController interface {
	Get(ctx *fiber.Ctx) error
	GetAll(ctx *fiber.Ctx) error
	Create(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
}

type userController struct {
	Log         *logrus.Logger
	Validator   *validator.Validate
	UserUsecase usecase.UserUsecase
}

func NewUserController(log *logrus.Logger, validator *validator.Validate, userUsecase usecase.UserUsecase) UserController {
	return &userController{
		Log:         log,
		Validator:   validator,
		UserUsecase: userUsecase,
	}
}

func (c *userController) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return model.BadRequest(ctx, "Invalid user ID")
	}

	user, err := c.UserUsecase.Get(ctx.UserContext(), uuid)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx, "Failed to retrieve user")
	}

	return model.OK(ctx, nil, user)
}

func (c *userController) GetAll(ctx *fiber.Ctx) error {

	params := user.GetParams()
	if err := ctx.QueryParser(params); err != nil {
		return model.BadRequest(ctx, "Malformed query parameters")
	}

	users, total, err := c.UserUsecase.GetAll(ctx.UserContext(), params)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	if total == 0 {
		users = []user.GetAllResponse{}
	}

	return model.OK(ctx, nil, users, model.NewPagination(params.Limit, params.Offset, int(total)))
}

func (c *userController) Create(ctx *fiber.Ctx) error {
	var req user.CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.BadRequest(ctx, "Malformed JSON payload")
	}

	if err := c.Validator.Struct(req); err != nil {
		errs := err.(validator.ValidationErrors)[0]
		return model.BadRequest(ctx, strings.ToLower(fmt.Sprintf("field '%s' is %s", errs.Field(), errs.Tag())))
	}

	if err := c.UserUsecase.Create(ctx.UserContext(), &req); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx, "Failed to create user")
	}

	return model.Created(ctx, "User created successfully")
}

func (c *userController) Update(ctx *fiber.Ctx) error {
	var req user.UpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.BadRequest(ctx, "Malformed JSON payload")
	}

	if err := c.Validator.Struct(req); err != nil {
		errs := err.(validator.ValidationErrors)[0]
		return model.BadRequest(ctx, strings.ToLower(fmt.Sprintf("field '%s' is %s", errs.Field(), errs.Tag())))
	}

	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return model.BadRequest(ctx, "Invalid user ID")
	}

	req.ID = uuid
	req.UpdateBy = middleware.GetUser(ctx).ID.String()
	if err := c.UserUsecase.Update(ctx.UserContext(), &req); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx, "Failed to update user")
	}

	return model.OK(ctx, "User updated successfully")
}

func (c *userController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return model.BadRequest(ctx, "Invalid user ID")
	}

	req := &user.DeleteRequest{
		ID:        uuid,
		DeletedBy: middleware.GetUser(ctx).ID,
	}

	if err := c.UserUsecase.Delete(ctx.UserContext(), req); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx, "Failed to delete user")
	}

	return model.NoContent(ctx)
}
