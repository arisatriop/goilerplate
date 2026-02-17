package response

import (
	"errors"
	"net/http"

	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// Meta contains metadata about the response

type Meta struct {
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// BaseResponse is the standard structure for all API responses
type BaseResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// PaginatedResponse extends BaseResponse for paginated data
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination interface{} `json:"pagination"`
	Meta       *Meta       `json:"meta,omitempty"`
	Errors     interface{} `json:"errors,omitempty"`
}

// ResponseOption allows customizing the response
type ResponseOption func(*BaseResponse)

// WithMeta adds metadata to the response
func WithMeta(meta *Meta) ResponseOption {
	return func(r *BaseResponse) {
		r.Meta = meta
	}
}

// WithMessage sets a custom message
func WithMessage(message string) ResponseOption {
	return func(r *BaseResponse) {
		r.Message = message
	}
}

// Success sends a successful response
func Success(ctx *fiber.Ctx, data interface{}, options ...ResponseOption) error {
	response := &BaseResponse{
		Success: true,
		Message: constants.MsgSuccess,
		Data:    data,
	}

	for _, opt := range options {
		opt(response)
	}

	return ctx.Status(http.StatusOK).JSON(response)
}

// Created sends a successful creation response
func Created(ctx *fiber.Ctx, data interface{}, options ...ResponseOption) error {
	response := &BaseResponse{
		Success: true,
		Message: constants.MsgResourceCreatedSuccessfully,
		Data:    data,
	}

	for _, opt := range options {
		opt(response)
	}

	return ctx.Status(http.StatusCreated).JSON(response)
}

// NoContent sends a successful response with no content
func NoContent(ctx *fiber.Ctx, options ...ResponseOption) error {
	response := &BaseResponse{
		Success: true,
		Message: constants.MsgOperationCompletedSuccessfully,
	}

	for _, opt := range options {
		opt(response)
	}

	return ctx.Status(http.StatusNoContent).JSON(response)
}

// BadRequest sends a bad request error response
func BadRequest(ctx *fiber.Ctx, message string, errors interface{}) error {
	return ctx.Status(http.StatusBadRequest).JSON(&BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// Unauthorized sends an unauthorized error response
func Unauthorized(ctx *fiber.Ctx, message string) error {
	if message == "" {
		message = constants.MsgUnauthorized
	}
	return ctx.Status(http.StatusUnauthorized).JSON(&BaseResponse{
		Success: false,
		Message: message,
	})
}

// Forbidden sends a forbidden error response
func Forbidden(ctx *fiber.Ctx, message string) error {
	if message == "" {
		message = constants.MsgForbidden
	}
	return ctx.Status(http.StatusForbidden).JSON(&BaseResponse{
		Success: false,
		Message: message,
	})
}

// NotFound sends a not found error response
func NotFound(ctx *fiber.Ctx, message string) error {
	if message == "" {
		message = constants.MsgResourceNotFound
	}
	return ctx.Status(http.StatusNotFound).JSON(&BaseResponse{
		Success: false,
		Message: message,
	})
}

// Conflict sends a conflict error response
func Conflict(ctx *fiber.Ctx, message string, errors interface{}) error {
	return ctx.Status(http.StatusConflict).JSON(&BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// UnprocessableEntity sends an unprocessable entity error response
func UnprocessableEntity(ctx *fiber.Ctx, message string, errors interface{}) error {
	return ctx.Status(http.StatusUnprocessableEntity).JSON(&BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// InternalServerError sends an internal server error response
func InternalServerError(ctx *fiber.Ctx, message string) error {
	if message == "" {
		message = constants.MsgInternalServerError
	}
	return ctx.Status(http.StatusInternalServerError).JSON(&BaseResponse{
		Success: false,
		Message: message,
	})
}

// ValidationError formats validation errors in a standardized way
func ValidationError(ctx *fiber.Ctx, errors interface{}) error {
	return ctx.Status(http.StatusBadRequest).JSON(&BaseResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

// Paginated sends a paginated response
func Paginated(ctx *fiber.Ctx, data interface{}, paginationData interface{}, options ...ResponseOption) error {
	response := &PaginatedResponse{
		Success:    true,
		Message:    constants.MsgSuccess,
		Data:       data,
		Pagination: paginationData,
	}

	// Apply options to the base response fields
	baseResponse := &BaseResponse{
		Success: response.Success,
		Message: response.Message,
		Meta:    response.Meta,
	}

	for _, opt := range options {
		opt(baseResponse)
	}

	// Update the paginated response with modified values
	response.Success = baseResponse.Success
	response.Message = baseResponse.Message
	response.Meta = baseResponse.Meta

	return ctx.Status(http.StatusOK).JSON(response)
}

// CustomError sends a custom error response with specified status code
func CustomError(ctx *fiber.Ctx, statusCode int, message string, errors interface{}) error {
	return ctx.Status(statusCode).JSON(&BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// SuccessWithMeta is a shorthand for Success with metadata
func SuccessWithMeta(ctx *fiber.Ctx, data interface{}, message string, meta *Meta) error {
	return Success(ctx, data, WithMessage(message), WithMeta(meta))
}

// ErrorWithDetails sends an error response with detailed error information
func ErrorWithDetails(ctx *fiber.Ctx, statusCode int, message string, details interface{}, meta *Meta) error {
	response := &BaseResponse{
		Success: false,
		Message: message,
		Errors:  details,
		Meta:    meta,
	}

	return ctx.Status(statusCode).JSON(response)
}

// HandleError handles errors from use case calls with consistent error responses
// It distinguishes between client errors (validation, business logic) and internal errors
// This is a reusable helper for all handlers to maintain consistent error handling
func HandleError(ctx *fiber.Ctx, err error) error {
	var clientError *utils.ClientError
	if errors.As(err, &clientError) {
		return CustomError(ctx, clientError.Code, clientError.Message, nil)
	}

	logger.Error(ctx.UserContext(), err)
	return InternalServerError(ctx, constants.MsgInternalServerError)
}
