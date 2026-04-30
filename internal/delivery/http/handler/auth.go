package handler

import (
	"time"

	"goilerplate/internal/application/register"
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/user"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Auth struct {
	deviceService      auth.DeviceService
	validator          *validator.Validate
	applicationService register.ApplicationService
	usecase            auth.Usecase
}

func NewAuth(deviceService auth.DeviceService, validator *validator.Validate, applicationService register.ApplicationService, usecase auth.Usecase) *Auth {
	return &Auth{
		validator:          validator,
		deviceService:      deviceService,
		usecase:            usecase,
		applicationService: applicationService,
	}
}

// Register handles user registration
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.RegisterRequest  true  "Registration data"
// @Success      201      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Router       /api/v1/auth/register [post]
func (h *Auth) Register(ctx *fiber.Ctx) error {
	var req dtorequest.RegisterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	register := register.Register{
		User: &user.User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: req.Password,
		},
	}

	if err := h.applicationService.Register(ctx.UserContext(), &register); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil, response.WithMessage("User registered successfully"))
}

// Login handles user authentication
// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.LoginRequest  true  "Login credentials"
// @Success      200      {object}  response.BaseResponse{data=dtoresponse.LoginResponse}
// @Failure      400      {object}  response.BaseResponse
// @Failure      401      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Router       /api/v1/auth/login [post]
func (h *Auth) Login(ctx *fiber.Ctx) error {
	var req dtorequest.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	credentials := &auth.LoginCredentials{
		Email:      req.Email,
		Password:   req.Password,
		RememberMe: req.RememberMe,
	}

	deviceInfo := h.deviceService.ExtractDeviceInfo(ctx)

	loginResult, err := h.usecase.Login(ctx.UserContext(), credentials, deviceInfo)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	// Map to response DTO
	responseData := presenter.ToLoginResponse(loginResult)

	return response.Success(ctx, responseData, response.WithMessage("Login successful"))
}

// Logout handles user logout by invalidating the access token
// @Summary      Logout
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.BaseResponse
// @Failure      401  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/auth/logout [post]
func (h *Auth) Logout(ctx *fiber.Ctx) error {
	// Get user ID, token hash, and session ID from context (guaranteed by middleware)
	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)
	tokenHash := ctx.Locals(string(constants.ContextTokenHash)).(string)
	sessionID := ctx.Locals(string(constants.ContextKeySessionID)).(string)

	// Call logout usecase
	if err := h.usecase.Logout(ctx.UserContext(), userID, tokenHash, sessionID); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage("Logout successful"))
}

// LogoutAll handles logout from all devices for a user
// @Summary      Logout from all devices
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.BaseResponse
// @Failure      401  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/auth/logout-all [post]
func (h *Auth) LogoutAll(ctx *fiber.Ctx) error {
	// Get user ID from context (guaranteed by middleware)
	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)

	// Call logout all usecase
	if err := h.usecase.LogoutAll(ctx.UserContext(), userID); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage("Logout from all devices successful"))
}

// RefreshToken handles token refresh using refresh token
// @Summary      Refresh access token
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.BaseResponse{data=dtoresponse.LoginResponse}
// @Failure      401  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/auth/refresh [post]
func (h *Auth) RefreshToken(ctx *fiber.Ctx) error {
	// Get data from context (guaranteed by AuthenticateRefreshToken middleware)
	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)
	sessionID := ctx.Locals(string(constants.ContextKeySessionID)).(string)
	tokenHash := ctx.Locals(string(constants.ContextTokenHash)).(string)
	refreshToken := ctx.Locals("refresh_token").(string)
	refreshTokenExpiresAt := ctx.Locals("refresh_token_expires_at").(time.Time)

	// Extract device information
	deviceInfo := h.deviceService.ExtractDeviceInfo(ctx)

	// Call refresh token usecase
	loginResult, err := h.usecase.RefreshToken(ctx.UserContext(), userID, sessionID, tokenHash, refreshToken, refreshTokenExpiresAt, deviceInfo)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	// Map to response DTO
	responseData := presenter.ToLoginResponse(loginResult)

	return response.Success(ctx, responseData, response.WithMessage("Token refreshed successfully"))
}
