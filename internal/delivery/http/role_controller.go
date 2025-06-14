package http

import (
	"errors"
	"fmt"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/internal/model"
	"goilerplate/internal/model/role"
	"goilerplate/internal/usecase"
	"goilerplate/pkg/helper"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type RoleController interface {
	Get(ctx *fiber.Ctx) error
	GetAll(ctx *fiber.Ctx) error
	Create(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
}

type roleController struct {
	Log       *logrus.Logger
	Validator *validator.Validate
	RoleUc    usecase.RoleUseCase
}

func NewRoleController(log *logrus.Logger, validator *validator.Validate, roleUc usecase.RoleUseCase) RoleController {
	return &roleController{
		Log:       log,
		Validator: validator,
		RoleUc:    roleUc,
	}
}

func (c *roleController) Create(ctx *fiber.Ctx) error {
	var req role.CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.BadRequest(ctx, "Invalid request body")
	}

	if err := c.Validator.Struct(req); err != nil {
		if err := c.Validator.Struct(req); err != nil {
			errs := err.(validator.ValidationErrors)[0]
			return model.BadRequest(ctx, strings.ToLower(fmt.Sprintf("field '%s' is %s", errs.Field(), errs.Tag())))
		}
	}

	req.CreatedBy = middleware.GetUser(ctx).ID.String()
	req.UpdatedBy = &req.CreatedBy
	if err := c.RoleUc.Create(ctx.UserContext(), &req); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		c.Log.Error("Failed to create role: ", err)
		return model.InternalServerError(ctx)
	}

	return model.Created(ctx, "Role created successfully")
}

func (c *roleController) Update(ctx *fiber.Ctx) error {

	var req role.UpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		c.Log.Error("Failed to parse request body: ", err)
		return model.BadRequest(ctx, "Invalid request body")
	}

	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return model.BadRequest(ctx, "Invalid user ID")
	}

	req.ID = uuid
	updatedBy := middleware.GetUser(ctx).ID.String()
	req.UpdatedBy = &updatedBy
	if err := c.Validator.Struct(req); err != nil {
		if err := c.Validator.Struct(req); err != nil {
			errs := err.(validator.ValidationErrors)[0]
			return model.BadRequest(ctx, strings.ToLower(fmt.Sprintf("field '%s' is %s", errs.Field(), errs.Tag())))
		}
	}

	if err := c.RoleUc.Update(ctx.UserContext(), req.ID, &req); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		c.Log.Error("Failed to update role: ", err)
		return model.InternalServerError(ctx)
	}

	return model.OK(ctx, "Role updated successfully")
}

func (c *roleController) GetAll(ctx *fiber.Ctx) error {

	params := role.GetParams()
	if err := ctx.QueryParser(params); err != nil {
		return model.BadRequest(ctx, "Malformed JSON payload")
	}

	roles, total, err := c.RoleUc.GetAll(ctx.UserContext(), params)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		return model.InternalServerError(ctx)
	}

	if total == 0 {
		roles = []role.GetAllResponse{}
	}

	return model.OK(ctx, nil, roles, model.NewPagination(params.Limit, params.Offset, int(total)))
}

func (c *roleController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return model.BadRequest(ctx, "Invalid role ID")
	}

	req := role.DeleteRequest{
		ID:        uuid,
		DeletedBy: middleware.GetUser(ctx).ID,
	}

	if err := c.RoleUc.Delete(ctx.UserContext(), &req); err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		c.Log.Error("Failed to delete role: ", err)
		return model.InternalServerError(ctx)
	}

	return model.NoContent(ctx)
}

func (c *roleController) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return model.BadRequest(ctx, "Invalid role ID")
	}

	role, err := c.RoleUc.GetByID(ctx.UserContext(), uuid)
	if err != nil {
		var cerr *helper.ClientError
		if errors.As(err, &cerr) {
			return model.JSON(ctx, cerr.Code, cerr.Message)
		}
		c.Log.Error("Failed to get role by ID: ", err)
		return model.InternalServerError(ctx)
	}

	if role == nil {
		return model.NotFound(ctx, "Role not found")
	}

	return model.OK(ctx, nil, role)
}
