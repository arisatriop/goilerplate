package userrole

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository
	CreateUserRole(ctx context.Context, userRole *UserRole) error
}
