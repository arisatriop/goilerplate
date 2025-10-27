package constants

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

// Context Keys
const (
	ContextKeyRequestID ContextKey = "request_id"
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyUserName  ContextKey = "user_name"
	ContextTokenHash    ContextKey = "token_hash"
	ContextKeySessionID ContextKey = "session_id"
)

// HTTP Status Code Constants
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusPermissionDenied    = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusInternalServerError = 500
)

const (
	StatusEmailNotRegistered         = 404
	StatusInvalidCredentials         = 401
	StatusAccountLocked              = 423
	StatusAccountDisabled            = 403
	StatusEmailNotVerified           = 403
	StatusEmailAlreadyExists         = 409
	StatusUserNotFound               = 404
	StatusSessionNotFound            = 404
	StatusSessionExpired             = 401
	StatusSessionInvalid             = 401
	StatusTokenNotFound              = 404
	StatusTokenExpired               = 410
	StatusTokenAlreadyUsed           = 410
	StatusTokenInvalid               = 400
	StatusPasswordTooWeak            = 400
	StatusCurrentPasswordWrong       = 400
	StatusNotImplemented             = 501
	StatusAuthorizationHeaderMissing = 401
	StatusInvalidAuthorizationFormat = 401
	StatusTokenEmpty                 = 401
)
