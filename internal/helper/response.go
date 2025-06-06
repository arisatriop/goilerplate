package helper

import (
	"github.com/gofiber/fiber/v2"
)

func Res(ctx *fiber.Ctx, status int, message string, data ...any) error {
	type Response struct {
		Message string         `json:"message"`
		Data    map[string]any `json:"data,omitempty"`
	}

	var respData map[string]any
	if len(data) > 0 {
		if d, ok := data[0].(map[string]any); ok {
			respData = d
		}
	}

	return ctx.Status(status).JSON(&Response{
		Message: message,
		Data:    respData,
	})
}

func ResOK(ctx *fiber.Ctx, args ...any) error {
	type SuccessResponse struct {
		Message string `json:"message"`
		Data    any    `json:"data"`
	}

	var message string = "Success"
	var data any

	if len(args) > 0 {
		data = args[0]
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

func ResNoContent(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusNoContent)
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
