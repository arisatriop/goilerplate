package helper

import (
	"github.com/gofiber/fiber/v2"
)

func ResOK(ctx *fiber.Ctx, args ...any) error {
	type SuccessResponse struct {
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}

	var message string = "Success"
	var data map[string]interface{} = nil

	if len(args) > 0 {
		d, ok := args[0].(map[string]any)
		if ok {
			data = d
		}
	}
	if len(args) > 1 {
		message = args[1].(string)
	}

	return ctx.Status(fiber.StatusOK).JSON(&SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func ResCreated(ctx *fiber.Ctx, args ...string) error {
	type CreatedResponse struct {
		Message string `json:"message"`
	}

	if len(args) == 0 || args[0] == "" {
		args = append(args, "Data created successfully")
	}

	return ctx.Status(fiber.StatusCreated).JSON(&CreatedResponse{
		Message: args[0],
	})
}

func ResBadRequest(ctx *fiber.Ctx, args ...string) error {
	type BadRequestResponse struct {
		Message string `json:"message"`
	}

	if len(args) == 0 || args[0] == "" {
		args = append(args, "Invalid request payload")
	}

	return ctx.Status(fiber.StatusBadRequest).JSON(&BadRequestResponse{
		Message: args[0],
	})
}

func ResForbidden(ctx *fiber.Ctx, args ...string) error {
	type Forbidden struct {
		Message string `json:"message"`
	}

	if len(args) == 0 || args[0] == "" {
		args = append(args, "Forbidden")
	}

	return ctx.Status(fiber.StatusForbidden).JSON(&Forbidden{
		Message: args[0],
	})
}

func RespNotFound(ctx *fiber.Ctx, args ...string) error {
	type NotFoundResponse struct {
		Message string `json:"message"`
	}

	if len(args) == 0 || args[0] == "" {
		args = append(args, "Data not found")
	}

	return ctx.Status(fiber.StatusNotFound).JSON(&NotFoundResponse{
		Message: args[0],
	})
}

func ResInternalServerError(ctx *fiber.Ctx, args ...string) error {
	type InternalServerErrorResponse struct {
		Message string `json:"message"`
	}

	if len(args) == 0 || args[0] == "" {
		args = append(args, "Whoops, something went wrong")
	}

	return ctx.Status(fiber.StatusInternalServerError).JSON(&InternalServerErrorResponse{
		Message: args[0],
	})
}
