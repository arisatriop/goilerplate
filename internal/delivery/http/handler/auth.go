package handler

import (
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
			Name:  req.Name,
			Email: req.Email,
		},
		Password: req.Password,
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

	deviceInfo := h.deviceService.ExtractDeviceInfo(newDeviceRequest(ctx))

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
	// Get user ID and session ID from context (guaranteed by middleware)
	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)
	sessionID := ctx.Locals(string(constants.ContextKeySessionID)).(string)

	// Call logout usecase
	if err := h.usecase.Logout(ctx.UserContext(), userID, sessionID); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage("Logout successful"))
}

// LogoutAll handles logout from all devices for a user
// @Summary      Logout from all devices
// @Description  Revokes every session of the caller. With keep_current=true the session making the request is spared.
// @Tags         auth
// @Produce      json
// @Param        keep_current  query     bool  false  "Keep the current session signed in"
// @Success      200  {object}  response.BaseResponse
// @Failure      400  {object}  response.BaseResponse
// @Failure      401  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/auth/logout-all [post]
func (h *Auth) LogoutAll(ctx *fiber.Ctx) error {
	var req dtorequest.LogoutAllRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, "Invalid query parameters", nil)
	}

	// Both come from the middleware, so the only session that can be kept is the caller's own.
	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)
	keepSessionID := ""
	message := "Logout from all devices successful"
	if req.KeepCurrent {
		keepSessionID = ctx.Locals(string(constants.ContextKeySessionID)).(string)
		message = "Logout from all other devices successful"
	}

	if err := h.usecase.LogoutAll(ctx.UserContext(), userID, keepSessionID); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(message))
}

// ListSessions lists the devices the signed-in user is signed in on
// @Summary      List my sessions
// @Description  Active, unexpired sessions of the caller, most recently used first. The session making the request has current=true.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.BaseResponse{data=[]dtoresponse.ActiveSessionResponse}
// @Failure      401  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/users/me/sessions [get]
func (h *Auth) ListSessions(ctx *fiber.Ctx) error {
	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)
	sessionID := ctx.Locals(string(constants.ContextKeySessionID)).(string)

	sessions, err := h.usecase.ListSessions(ctx.UserContext(), userID)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	// Not paginated: this is one user's own devices, a handful of rows bounded by how many
	// places they sign in from, and a device list split across pages would be worse to use.
	return response.Success(ctx, presenter.ToActiveSessionsResponse(sessions, sessionID),
		response.WithMessage("Sessions fetched successfully"))
}

// RevokeSession signs one of the caller's own sessions out
// @Summary      Revoke one of my sessions
// @Description  Signs out one device. A session that does not exist, is already revoked, or belongs to someone else is 404 — the three are indistinguishable on purpose.
// @Tags         auth
// @Produce      json
// @Param        id   path      string  true  "Session ID"
// @Success      200  {object}  response.BaseResponse
// @Failure      400  {object}  response.BaseResponse
// @Failure      401  {object}  response.BaseResponse
// @Failure      404  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/users/me/sessions/{id} [delete]
func (h *Auth) RevokeSession(ctx *fiber.Ctx) error {
	var param dtorequest.SessionIDParam
	if err := ctx.ParamsParser(&param); err != nil {
		return response.BadRequest(ctx, "Invalid path parameters", nil)
	}
	if err := h.validator.Struct(&param); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	// The user ID comes from the token, never the request, so only the caller's own sessions
	// can match.
	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)

	if err := h.usecase.RevokeSession(ctx.UserContext(), userID, param.ID); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage("Session revoked successfully"))
}

// ChangePassword changes the signed-in user's password
// @Summary      Change password
// @Description  Requires the current password. Revokes every other session; the caller stays signed in.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.ChangePasswordRequest  true  "Password change payload"
// @Success      200  {object}  response.BaseResponse
// @Failure      400  {object}  response.BaseResponse
// @Failure      401  {object}  response.BaseResponse
// @Failure      500  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/users/me/password [put]
func (h *Auth) ChangePassword(ctx *fiber.Ctx) error {
	var req dtorequest.ChangePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	// Both come from the middleware, so the caller can only ever change their own password and
	// can only keep the session they are actually using.
	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)
	sessionID := ctx.Locals(string(constants.ContextKeySessionID)).(string)

	if err := h.usecase.ChangePassword(ctx.UserContext(), userID, sessionID, req.CurrentPassword, req.NewPassword); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage("Password changed successfully"))
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
	refreshJTI := ctx.Locals("refresh_jti").(string)

	// Extract device information
	deviceInfo := h.deviceService.ExtractDeviceInfo(newDeviceRequest(ctx))

	// Call refresh token usecase
	loginResult, err := h.usecase.RefreshToken(ctx.UserContext(), userID, sessionID, refreshJTI, deviceInfo)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	// Map to response DTO
	responseData := presenter.ToLoginResponse(loginResult)

	return response.Success(ctx, responseData, response.WithMessage("Token refreshed successfully"))
}

// newDeviceRequest collects the request attributes the auth domain uses to identify a device.
func newDeviceRequest(ctx *fiber.Ctx) auth.DeviceRequest {
	return auth.DeviceRequest{
		UserAgent:      ctx.Get(fiber.HeaderUserAgent),
		AcceptLanguage: ctx.Get(fiber.HeaderAcceptLanguage),
		AcceptEncoding: ctx.Get(fiber.HeaderAcceptEncoding),
		// ctx.IP() already applied the trusted-proxy check configured in bootstrap, so a
		// forwarding header from an untrusted peer has been ignored by this point.
		ClientIP: ctx.IP(),
	}
}
