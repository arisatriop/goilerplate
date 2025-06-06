package model

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type jsonResponse struct {
	Message    string      `json:"message"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page      int   `json:"page"`
	Size      int   `json:"size"`
	TotalItem int64 `json:"total_item"`
	TotalPage int64 `json:"total_page"`
}

func NewPagination(limit, offset, total int) *Pagination {
	if limit <= 0 {
		limit = 10
	}
	return &Pagination{
		Page:      offset/limit + 1,
		Size:      limit,
		TotalItem: int64(total),
		TotalPage: int64((total + limit - 1) / limit),
	}
}

func newJsonResponse(args ...any) jsonResponse {
	var data any
	var message string
	var pagination *Pagination

	if len(args) > 0 {
		if args[0] != nil && args[0].(string) != "" {
			message = args[0].(string)
		}
	}
	if len(args) > 1 {
		data = args[1]
	}
	if len(args) > 2 {
		pagination = args[2].(*Pagination)
	}
	return jsonResponse{
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

// JSON contain mandatory code parameter and can contain three optional parameters:
// First: a message string,
// Second: data to be returned,
// Third: a Pagination object for paginated responses.
// If no message is provided, the default depends on the status code:
// - 200: "Success"
// - 201: "Created"
// - 204: "No Content"
// - 400: "Bad Request"
// - 401: "Unauthorized"
// - 403: "Forbidden"
// - 404: "Not Found"
// - 500: "Internal Server Error"
// If the message is not provided, it will be set based on the status code.
// If the status code is not recognized, it defaults to "Unknown Status".
// The function returns a JSON response with the specified status code.
func JSON(ctx *fiber.Ctx, code int, args ...any) error {
	response := newJsonResponse(args...)
	if response.Message == "" {
		switch code {
		case 200:
			response.Message = "Success"
		case 201:
			response.Message = "Created"
		case 204:
			response.Message = "No Content"
		case 400:
			response.Message = "Bad Request"
		case 401:
			response.Message = "Unauthorized"
		case 403:
			response.Message = "Forbidden"
		case 404:
			response.Message = "Not Found"
		case 500:
			response.Message = "Internal Server Error"
		default:
			response.Message = "Unknown Status"
		}
	}
	return ctx.Status(code).JSON(response)
}

// OK can contain three optional parameters:
// First: a message string,
// Second: data to be returned,
// Third: a Pagination object for paginated responses.
// If no message is provided, it defaults to "Success".
func OK(ctx *fiber.Ctx, args ...any) error {
	response := newJsonResponse(args...)
	fmt.Printf("message: %v\n", response.Message)
	if response.Message == "" {
		response.Message = "Success"
	}
	return ctx.Status(200).JSON(response)
}

// Created contain optional parameters message
// If no message is provided, it defaults to "Created".
func Created(ctx *fiber.Ctx, args ...any) error {
	response := newJsonResponse(args...)
	if response.Message == "" {
		response.Message = "Created"
	}
	return ctx.Status(201).JSON(response)
}

func NoContent(ctx *fiber.Ctx, args ...any) error {
	return ctx.Status(204).JSON(nil)
}

func BadRequest(ctx *fiber.Ctx, args ...any) error {
	response := newJsonResponse(args...)
	if response.Message == "" {
		response.Message = "Bad request"
	}
	return ctx.Status(400).JSON(response)
}

func Unauthorized(ctx *fiber.Ctx, args ...any) error {
	response := newJsonResponse(args...)
	if response.Message == "" {
		response.Message = "Unauthorized"
	}
	return ctx.Status(401).JSON(response)
}

func Forbidden(ctx *fiber.Ctx, args ...any) error {
	response := newJsonResponse(args...)
	if response.Message == "" {
		response.Message = "Forbidden"
	}
	return ctx.Status(403).JSON(response)
}

func NotFound(ctx *fiber.Ctx, args ...any) error {
	response := newJsonResponse(args...)
	if response.Message == "" {
		response.Message = "Not found"
	}
	return ctx.Status(404).JSON(response)
}

// InternalServerError can contain optional parameters message.
// If no message is provided, it will returned default message.
func InternalServerError(ctx *fiber.Ctx, args ...any) error {
	response := newJsonResponse(args...)
	if response.Message == "" {
		response.Message = "Whoops, something went wrong"
	}
	return ctx.Status(500).JSON(response)
}
