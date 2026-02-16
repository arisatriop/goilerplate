package registration

import (
	"goilerplate/internal/domain/store"
	"goilerplate/internal/domain/user"
)

type Registration struct {
	User  *user.User
	Store *store.Store
}
