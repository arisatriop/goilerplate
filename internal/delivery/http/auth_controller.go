package http

import (
	"errors"
	"fmt"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/helper"
	"goilerplate/internal/model"
	"goilerplate/internal/model/auth"
	"goilerplate/internal/usecase"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthController interface {
	Me(ctx *fiber.Ctx) error
	Login(ctx *fiber.Ctx) error
	Logout(ctx *fiber.Ctx) error
	Token(ctx *fiber.Ctx) error
}

type authController struct {
	Log         *logrus.Logger
	Validator   *validator.Validate
	AuthUsecase usecase.AuthUsecase
}

func NewAuthController(log *logrus.Logger, validator *validator.Validate, authUsecase usecase.AuthUsecase) AuthController {
	return &authController{
		Log:         log,
		Validator:   validator,
		AuthUsecase: authUsecase,
	}
}

func (c *authController) Me(ctx *fiber.Ctx) error {
	user := middleware.GetUser(ctx)

	if user == nil {
		return model.Unauthorized(ctx, "Unauthorized access")
	}

	resp, err := c.AuthUsecase.Me(ctx.UserContext(), user)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	return model.OK(ctx, nil, resp)
}

func (c *authController) Login(ctx *fiber.Ctx) error {
	var req auth.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.BadRequest(ctx, "Malformed JSON payload")
	}

	if err := c.Validator.Struct(req); err != nil {
		errs := err.(validator.ValidationErrors)[0]
		return model.BadRequest(ctx, strings.ToLower(fmt.Sprintf("field '%s' is %s", errs.Field(), errs.Tag())))
	}

	resp, err := c.AuthUsecase.Login(ctx.UserContext(), &req)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	return model.OK(ctx, nil, resp)
}

func (c *authController) Logout(ctx *fiber.Ctx) error {

	req := auth.LogoutRequest{
		ID: middleware.GetUser(ctx).ID,
	}

	if err := c.AuthUsecase.Logout(ctx.UserContext(), &req); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	return model.OK(ctx, "Logout successful")
}

func (c *authController) Token(ctx *fiber.Ctx) error {
	var req auth.TokenRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.BadRequest(ctx, "Malformed JSON payload")
	}

	if err := c.Validator.Struct(req); err != nil {
		errs := err.(validator.ValidationErrors)[0]
		return model.BadRequest(ctx, strings.ToLower(fmt.Sprintf("field '%s' is %s", errs.Field(), errs.Tag())))
	}

	resp, err := c.AuthUsecase.Token(ctx.UserContext(), &req)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	return model.OK(ctx, nil, resp)

}
