package http

import (
	"errors"
	"fmt"
	"golang-clean-architecture/internal/helper"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/model/user"
	"golang-clean-architecture/internal/usecase"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type IUserController interface {
	Create(ctx *fiber.Ctx) error
}

type UserController struct {
	Log         *logrus.Logger
	Validator   *validator.Validate
	UserUsecase usecase.IUserUsecase
}

func NewUserController(log *logrus.Logger, validator *validator.Validate, userUsecase usecase.IUserUsecase) IUserController {
	return &UserController{
		Log:         log,
		Validator:   validator,
		UserUsecase: userUsecase,
	}
}

func (c *UserController) Create(ctx *fiber.Ctx) error {
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
