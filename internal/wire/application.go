package wire

import (
	"goilerplate/internal/application/foo"
	"goilerplate/internal/application/registration"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/infrastructure/transaction"
)

// ApplicationServices contains all application services for multi-domain orchestration
type ApplicationServices struct {
	RegistrationService registration.ApplicationService
	FooService          foo.ApplicationService
}

func WireApplicationServices(app *bootstrap.App, repos *Repositories, usecases *UseCases, infrastructure *Infrastructure) *ApplicationServices {
	txManager := transaction.NewGormTransaction(app.DB.GDB)

	return &ApplicationServices{
		RegistrationService: registration.NewApplicationService(
			app.Config,
			txManager,
			repos.UserRepo,
			repos.RoleRepo,
			repos.UserRoleRepo,
		),
		FooService: foo.NewApplicationService(
			txManager,
			usecases.FooUC,
			repos.FooRepo,
			repos.BarRepo,
		),
	}
}
