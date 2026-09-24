package dtorequest

// RegisterRequest represents the registration request data
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents the login request data
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"rememberMe"`
}

// LogoutAllRequest carries the options of POST /auth/logout-all, read from the query string so
// the endpoint still takes no body.
type LogoutAllRequest struct {
	// KeepCurrent spares the session making the request: "sign out every other device".
	KeepCurrent bool `query:"keep_current"`
}

// SessionIDParam is the :id path segment of /users/me/sessions/:id. It is validated as a UUID so
// a malformed value is a 400 here rather than a type error from PostgreSQL.
type SessionIDParam struct {
	ID string `params:"id" validate:"required,uuid"`
}

// ChangePasswordRequest represents a password change by the signed-in user.
// The length ceiling is enforced by pkg/password, which counts bytes rather than characters
// because that is where bcrypt truncates.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}
