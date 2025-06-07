package auth

type TokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
