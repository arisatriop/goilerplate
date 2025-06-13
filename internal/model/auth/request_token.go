package auth

type TokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}
