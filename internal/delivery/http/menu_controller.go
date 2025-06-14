package http

import (
	"errors"
	"goilerplate/internal/model"
	"goilerplate/internal/model/menu"
	"goilerplate/internal/usecase"
	"goilerplate/pkg/helper"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type MenuController interface {
	GetAll(ctx *fiber.Ctx) error
}

type menuController struct {
	Log         *logrus.Logger
	Validator   *validator.Validate
	MenuUsecase usecase.MenuUsecase
}

func NewMenuController(log *logrus.Logger, validator *validator.Validate, menuUsecase usecase.MenuUsecase) MenuController {
	return &menuController{
		Log:         log,
		Validator:   validator,
		MenuUsecase: menuUsecase,
	}
}

func (c *menuController) GetAll(ctx *fiber.Ctx) error {
	params := &menu.GetRequest{}
	if err := ctx.QueryParser(params); err != nil {
		return model.BadRequest(ctx, "Malformed JSON payload")
	}

	menus, total, err := c.MenuUsecase.GetAll(ctx.UserContext(), params)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	if total == 0 {
		menus = []menu.GetAllResponse{}
	}

	return model.OK(ctx, nil, menus)
}
