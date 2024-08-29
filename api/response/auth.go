package response

import "goilerplate/app/entity"

func NewAuthResponse() IAuth {
	return &AuthImpl{}
}

type AuthImpl struct{}

type IAuth interface {
	Login(*entity.Auth) (*AuthLogin, error)
}

type AuthLogin struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	Abilities   string `json:"abilities"`
}

func (r *AuthImpl) Login(auth *entity.Auth) (*AuthLogin, error) {
	return &AuthLogin{
		AccessToken: auth.AccessToken,
		TokenType:   auth.TokenType,
		ExpiresIn:   auth.ExpiresIn,
		Scope:       auth.Scope,
		Abilities:   auth.Abilities,
	}, nil
}
