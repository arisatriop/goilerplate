package register

import (
	"goilerplate/internal/domain/user"
)

// Register is the input for the registration flow.
//
// Password is the plaintext as the client sent it. It is kept separate from User so that no
// field named PasswordHash ever holds an unhashed value: the service validates this field
// against the password policy and then hands it to User.SetPassword.
type Register struct {
	User     *user.User
	Password string
}
