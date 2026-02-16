package handler

import (
	"goilerplate/internal/application/expapp"
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/domain/zexample"
	"goilerplate/internal/domain/zexample2"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Example2 struct {
	Validator           *validator.Validate
	AppplicationService expapp.ApplicationService
}

func NewExample2(validator *validator.Validate, applicationService expapp.ApplicationService) *Example2 {
	return &Example2{
		Validator:           validator,
		AppplicationService: applicationService,
	}
}

func (h *Example2) CreateSomething(ctx *fiber.Ctx) error {
	var req dtorequest.Example2CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	exampleEntity := &zexample.Example{
		Code:    req.SomethingField1,
		Example: req.SomethingField2,
	}

	example2 := make([]*zexample2.Example2, len(req.SomethingElseField))
	for i, v := range req.SomethingElseField {
		example2[i] = &zexample2.Example2{
			Example: v,
		}
	}

	exp := expapp.Exp{
		Example:  exampleEntity,
		Example2: example2,
	}

	err := h.AppplicationService.CreateSomething(ctx.UserContext(), &exp)
	if err != nil {
		return response.HandleError(ctx, err)

	}

	return response.Success(ctx, nil, response.WithMessage(constants.MsgSuccess))
}
