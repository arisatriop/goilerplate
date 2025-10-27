package dtorequest

// RegisterRequest represents the registration request data
type RegisterRequest struct {
	Name       string `json:"name" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Phone      string `json:"phone"`
	Avatar     string `json:"avatar"`
	DeviceName string `json:"deviceName,omitempty"` // Optional: client can provide custom device name
	DeviceType string `json:"deviceType,omitempty"` // Optional: will be auto-detected if not provided
	DeviceID   string `json:"deviceId,omitempty"`   // Optional: will be generated if not provided
	UserAgent  string `json:"-"`                    // Server-side only
	IPAddress  string `json:"-"`                    // Server-side only
}

// LoginRequest represents the login request data
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"rememberMe"`
}

// ForgotPasswordRequest represents the forgot password request
type ForgotPasswordRequest struct {
	Email     string `json:"email" validate:"required,email"`
	UserAgent string `json:"-"`
	IPAddress string `json:"-"`
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=NewPassword"`
}

// ResetPasswordRequest represents the reset password request
type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=NewPassword"`
	UserAgent       string `json:"-"`
	IPAddress       string `json:"-"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
	DeviceID     string `json:"deviceId"`
	UserAgent    string `json:"-"`
	IPAddress    string `json:"-"`
}
