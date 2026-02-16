package dtorequest

// RegisterRequest represents the registration request data
type RegisterRequest struct {
	Name         string  `json:"name" validate:"required"`
	Phone        string  `json:"phone" validate:"required"`
	Email        string  `json:"email" validate:"required,email"`
	Password     string  `json:"password" validate:"required,min=8"`
	StoreName    string  `json:"storeName" validate:"required"`
	StoreDesc    *string `json:"storeDesc"`
	StoreAddress *string `json:"storeAddress"`
	StorePhone   *string `json:"StorePhone"`
	StoreEmail   *string `json:"StoreEmail"`
}

// LoginRequest represents the login request data
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"rememberMe"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
	DeviceID     string `json:"deviceId"`
	UserAgent    string `json:"-"`
	IPAddress    string `json:"-"`
}
