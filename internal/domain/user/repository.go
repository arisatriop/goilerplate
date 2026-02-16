package user

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	FindByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
}
