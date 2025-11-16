package role

import "context"

type Repository interface {
	WithTx(ctx context.Context) Repository

	GetRoleBySlug(ctx context.Context, slug string) (*Role, error)
}
