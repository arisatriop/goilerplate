package constants

// Success Messages
const (
	MsgSuccess                        = "Success"
	MsgResourceCreatedSuccessfully    = "Resource created successfully"
	MsgOperationCompletedSuccessfully = "Operation completed successfully"
)

// Error Messages
const (
	MsgInternalServerError = "Internal Server Error"
	MsgInvalidRequestBody  = "Invalid request body"
	MsgUnauthorized        = "Unauthorized"
	MsgForbidden           = "Forbidden"
	MsgPermissionDenied    = "Permission denied"
	MsgValidationFailed    = "Validation failed"
)

// Session Messages
const (
	MsgSessionNotFound = "Session not found"
	MsgSessionExpired  = "Session has expired"
	MsgSessionInvalid  = "Invalid session"
)

// Token Messages
const (
	MsgTokenNotFound    = "Token not found"
	MsgTokenExpired     = "Token has expired"
	MsgTokenAlreadyUsed = "Token has already been used"
	MsgTokenInvalid     = "Invalid token"
	MsgTokenEmpty       = "Token is empty"
)

// Other Messages
const (
	MsgEmailNotRegistered         = "Email address is not registered"
	MsgInvalidCredentials         = "Invalid email or password"
	MsgAccountLocked              = "Account is temporarily locked due to multiple failed login attempts"
	MsgAccountDisabled            = "Account is disabled"
	MsgEmailNotVerified           = "Email address is not verified"
	MsgEmailAlreadyExists         = "Email address is already registered"
	MsgUserNotFound               = "User not found"
	MsgPasswordTooWeak            = "Password does not meet security requirements"
	MsgCurrentPasswordWrong       = "Current password is incorrect"
	MsgNotImplemented             = "Feature not implemented yet"
	MsgAuthorizationHeaderMissing = "Authorization header is required"
	MsgInvalidAuthorizationFormat = "Invalid authorization header format. Expected: Bearer <token>"
	MsgResourceNotFound           = "Resource not found"
)
