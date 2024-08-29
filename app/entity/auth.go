package entity

type Auth struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
	Scope       string
	Abilities   string
}
