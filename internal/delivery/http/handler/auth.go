package handler

import (
	"goilerplate/internal/application/register"
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/refreshtoken"
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
	// refresh hands out the refresh token: in the body or as an HttpOnly cookie. Nil is body.
	refresh *refreshtoken.Transport
	// registeredMessage answers a successful registration; see WithEmailFlows.
	registeredMessage string
}

// Registration messages. With email on, a taken address is answered like a new one, so the
// message must be true for both: it cannot say an account was created.
const (
	MsgRegistered         = "User registered successfully"
	MsgRegistrationQueued = "Registration received. Check your email to continue"
)

// WithEmailFlows switches registration to the answer that holds whether or not the address was
// already registered. Call it when auth.email.enabled is true.
func (h *Auth) WithEmailFlows() *Auth {
	h.registeredMessage = MsgRegistrationQueued
	return h
}

func NewAuth(deviceService auth.DeviceService, validator *validator.Validate, applicationService register.ApplicationService, usecase auth.Usecase, refresh *refreshtoken.Transport) *Auth {
	return &Auth{
		validator:          validator,
		deviceService:      deviceService,
		usecase:            usecase,
		applicationService: applicationService,
		refresh:            refresh,
	}
}

// tokenResponse shapes a login or refresh result, handing the refresh token over the configured
// transport: in cookie mode it goes into the cookie and never into the body.
func (h *Auth) tokenResponse(ctx *fiber.Ctx, result *auth.LoginResult) *dtoresponse.LoginResponse {
	body := presenter.ToLoginResponse(result)
	body.Tokens.RefreshToken = h.refresh.Issue(ctx, result.Tokens.RefreshToken, result.Tokens.RefreshTokenExpiresAt)
	if body.Tokens.RefreshToken == "" {
		body.Tokens.RefreshTokenType = ""
	}
	return body
}

// Register handles user registration
// @Summary      Register a new user
// @Description  With auth.email.enabled, an address that is already registered gets the same 201 as a new one and its owner is emailed instead, so the response never says which addresses have accounts. Without email, a taken address is 409 email_already_registered.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.RegisterRequest  true  "Registration data"
// @Success      201      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      409      {object}  response.BaseResponse  "email_already_registered, only when auth.email.enabled is false"
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
		Origin:   requestOrigin(ctx),
	}

	if err := h.applicationService.Register(ctx.UserContext(), &register); err != nil {
		return response.HandleError(ctx, err)
	}

	message := h.registeredMessage
	if message == "" {
		message = MsgRegistered
	}
	return response.Created(ctx, nil, response.WithMessage(message))
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

	return response.Success(ctx, h.tokenResponse(ctx, loginResult), response.WithMessage("Login successful"))
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
	h.refresh.Clear(ctx)

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
	// The kept session is this browser's, so its cookie stays; otherwise this device is signed
	// out along with the rest.
	if keepSessionID == "" {
		h.refresh.Clear(ctx)
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

// MsgPasswordResetRequested answers every forgot-password request, whether or not an email was
// sent, so the response never says which addresses are registered.
//
// #nosec G101 -- a message shown to the user, not a credential.
const MsgPasswordResetRequested = "If an account exists for that email, a password reset link has been sent"

// ForgotPassword emails a password reset link
// @Summary      Request a password reset link
// @Description  Always answers 200 with the same message, whether or not the address is registered. Available only when auth.email.enabled is true.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.ForgotPasswordRequest  true  "Account email"
// @Success      200      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      429      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Router       /api/v1/auth/forgot-password [post]
func (h *Auth) ForgotPassword(ctx *fiber.Ctx) error {
	var req dtorequest.ForgotPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	if err := h.usecase.ForgotPassword(ctx.UserContext(), req.Email, requestOrigin(ctx)); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(MsgPasswordResetRequested))
}

// ResetPassword sets a new password using the token from a reset link
// @Summary      Reset the password with an emailed token
// @Description  Signs the account out on every device. Available only when auth.email.enabled is true.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.ResetPasswordRequest  true  "Reset token and new password"
// @Success      200      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse  "validation_failed, a password the policy refuses, or invalid_reset_token"
// @Failure      429      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Router       /api/v1/auth/reset-password [post]
func (h *Auth) ResetPassword(ctx *fiber.Ctx) error {
	var req dtorequest.ResetPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	if err := h.usecase.ResetPassword(ctx.UserContext(), req.Token, req.NewPassword); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage("Password has been reset. Sign in with the new password"))
}

// MsgVerificationEmailRequested answers every request for a verification code, whether or not
// one was sent.
const MsgVerificationEmailRequested = "If the address needs verifying, a verification code has been sent"

// SendVerificationEmail emails a new verification code
// @Summary      Request an email verification code
// @Description  Always answers 200 with the same message: for an unknown, disabled or already verified address nothing is sent. A code is also sent automatically after registration. Available only when auth.email.enabled is true.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.SendVerificationEmailRequest  true  "Address to verify"
// @Success      200      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse
// @Failure      429      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Router       /api/v1/auth/send-verification-email [post]
func (h *Auth) SendVerificationEmail(ctx *fiber.Ctx) error {
	var req dtorequest.SendVerificationEmailRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	if err := h.usecase.SendEmailVerification(ctx.UserContext(), req.Email, requestOrigin(ctx)); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(MsgVerificationEmailRequested))
}

// VerifyEmail marks the address verified with the emailed code
// @Summary      Verify an email address with the emailed code
// @Description  Unauthenticated, so it works while auth.require_email_verification refuses login. Each code tolerates auth.otp.max_attempts guesses. Available only when auth.email.enabled is true.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.VerifyEmailRequest  true  "Address and 6-digit code"
// @Success      200      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse  "validation_failed or invalid_verification_code"
// @Failure      429      {object}  response.BaseResponse
// @Failure      500      {object}  response.BaseResponse
// @Router       /api/v1/auth/verify-email [post]
func (h *Auth) VerifyEmail(ctx *fiber.Ctx) error {
	var req dtorequest.VerifyEmailRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	if err := h.usecase.VerifyEmail(ctx.UserContext(), req.Email, req.Code); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage("Email address verified"))
}

// MsgEmailChangeRequested answers every email change request that passed the password check,
// whether or not a code was sent.
const MsgEmailChangeRequested = "If the address can be used, a confirmation code has been sent to it"

// RequestEmailChange sends a code to the new address
// @Summary      Start changing the account's email address
// @Description  Requires the current password. Answers 200 with the same message whether or not the address is free; a code is sent only when it is. Available only when auth.email.enabled is true.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.EmailChangeRequest  true  "New address and current password"
// @Success      200      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse  "validation_failed or email_unchanged"
// @Failure      401      {object}  response.BaseResponse  "invalid_credentials"
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/users/me/email-change [post]
func (h *Auth) RequestEmailChange(ctx *fiber.Ctx) error {
	var req dtorequest.EmailChangeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)
	if err := h.usecase.RequestEmailChange(ctx.UserContext(), userID, req.CurrentPassword, req.NewEmail, requestOrigin(ctx)); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage(MsgEmailChangeRequested))
}

// ConfirmEmailChange moves the account to the new address
// @Summary      Confirm an email change with the code sent to the new address
// @Description  The old address is told once the change is made. Available only when auth.email.enabled is true.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtorequest.ConfirmEmailChangeRequest  true  "6-digit code"
// @Success      200      {object}  response.BaseResponse
// @Failure      400      {object}  response.BaseResponse  "validation_failed or invalid_verification_code"
// @Failure      401      {object}  response.BaseResponse
// @Failure      409      {object}  response.BaseResponse  "email_already_registered: the address was taken since the code was sent"
// @Failure      500      {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /api/v1/users/me/email-change/confirm [post]
func (h *Auth) ConfirmEmailChange(ctx *fiber.Ctx) error {
	var req dtorequest.ConfirmEmailChangeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.ValidationError(ctx, response.FormatValidationErrors(err))
	}

	userID := ctx.Locals(string(constants.ContextKeyUserID)).(string)
	if err := h.usecase.ConfirmEmailChange(ctx.UserContext(), userID, req.Code); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage("Email address changed"))
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

	return response.Success(ctx, h.tokenResponse(ctx, loginResult), response.WithMessage("Token refreshed successfully"))
}

// requestOrigin records where a token-issuing request came from. ctx.IP() already applied the
// trusted-proxy check configured in bootstrap.
func requestOrigin(ctx *fiber.Ctx) auth.RequestOrigin {
	return auth.RequestOrigin{IPAddress: ctx.IP(), UserAgent: ctx.Get(fiber.HeaderUserAgent)}
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
