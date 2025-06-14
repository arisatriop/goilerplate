package handler

import (
	"errors"
	"goilerplate/internal/model"
	"goilerplate/internal/model/menupermission"
	"goilerplate/internal/usecase"
	"goilerplate/pkg/helper"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type MenuPermissionHandler interface {
	GetAll(ctx *fiber.Ctx) error
}

type menuPermissionHandler struct {
	Log                   *logrus.Logger
	Validator             *validator.Validate
	MenuPermissionUsecase usecase.MenuPermissionUsecase
}

func NewMenuPermissionHandler(log *logrus.Logger, validator *validator.Validate, menuPermissionUsecase usecase.MenuPermissionUsecase) MenuPermissionHandler {
	return &menuPermissionHandler{
		Log:                   log,
		Validator:             validator,
		MenuPermissionUsecase: menuPermissionUsecase,
	}
}

func (c *menuPermissionHandler) GetAll(ctx *fiber.Ctx) error {
	params := menupermission.GetParams()
	if err := ctx.QueryParser(params); err != nil {
		return model.BadRequest(ctx, "Invalid query parameters")
	}

	result, total, err := c.MenuPermissionUsecase.GetAll(ctx.UserContext(), params)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	if len(result) == 0 {
		result = []menupermission.GetAllResponse{}
	}

	return model.OK(ctx, nil, result, model.NewPagination(params.Limit, params.Offset, int(total)))

}
