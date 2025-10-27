package wire

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/zexample"
	"goilerplate/internal/infrastructure/repository"
)

// Repositories contains all repository implementations
type Repositories struct {
	AuthRepo    auth.Repository
	ExampleRepo zexample.Repository
	// Future repositories will be added here:
	// UserRepo    user.Repository
	// OrderRepo   order.Repository
	// ProductRepo product.Repository
}

// WireRepositories creates all repository implementations
func WireRepositories(app *bootstrap.App) *Repositories {
	return &Repositories{
		AuthRepo:    repository.NewAuth(app.DB.GDB),
		ExampleRepo: repository.NewExample(app.DB.GDB),
		// Future repository wiring:
		// UserRepo:    repository.NewUser(app.DB.GDB),
		// OrderRepo:   repository.NewOrder(app.DB.GDB),
		// ProductRepo: repository.NewProduct(app.DB.GDB),
	}
}
