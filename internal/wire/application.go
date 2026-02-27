package wire

import (
	"goilerplate/internal/application/example"
	"goilerplate/internal/application/register"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/infrastructure/transaction"
)

// ApplicationServices contains all application services for multi-domain orchestration
type ApplicationServices struct {
	ExpSvc      example.ApplicationService
	RegisterSvc register.ApplicationService
}

func WireApplicationServices(app *bootstrap.App, repos *Repositories, usecases *UseCases, infrastructure *Infrastructure) *ApplicationServices {
	txManager := transaction.NewGormTransaction(app.DB.GDB)

	return &ApplicationServices{
		ExpSvc: example.NewApplicationService(
			txManager,
			usecases.ZexampleUC,
			repos.ZexampleRepo,
		),
		RegisterSvc: register.NewApplicationService(
			app.Config,
			txManager,
			repos.UserRepo,
			repos.RoleRepo,
			repos.UserRoleRepo,
		),
	}
}
