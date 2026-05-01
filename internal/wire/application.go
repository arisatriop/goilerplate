package wire

import (
	"goilerplate/internal/application/bar"
	"goilerplate/internal/application/register"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/infrastructure/transaction"
)

// ApplicationServices contains all application services for multi-domain orchestration
type ApplicationServices struct {
	BarSvc      bar.ApplicationService
	RegisterSvc register.ApplicationService
}

func WireApplicationServices(app *bootstrap.App, repos *Repositories, usecases *UseCases, infrastructure *Infrastructure) *ApplicationServices {
	txManager := transaction.NewGormTransaction(app.DB.GDB)

	return &ApplicationServices{
		BarSvc: bar.NewApplicationService(
			txManager,
			usecases.BarUC,
			repos.BarRepo,
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
