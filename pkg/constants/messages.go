package constants

// Success Messages
const (
	MsgSuccess                        = "Success"
	MsgResourceCreatedSuccessfully    = "Resource created successfully"
	MsgOperationCompletedSuccessfully = "Operation completed successfully"
)

// Error Messages
const (
	MsgInvalidRequestBody    = "Invalid request body"
	MsgUnauthorized          = "Unauthorized"
	MsgForbidden             = "Forbidden"
	MsgResourceNotFound      = "Not found"
	MsgInternalServerError   = "Whoops! Something went wrong"
	MsgAccountLocked         = "Too many failed attempts, please try again later"
	MsgAccountDisabled       = "Account is disabled"
	MsgInvalidCredential     = "Invalid credential" // #nosec G101 -- a message shown to users.
	MsgFeatureNotImplemented = "Feature not implemented"
	MsgUnauthorizedAccess    = "Anda tidak memiliki akses untuk data tersebut"
)
