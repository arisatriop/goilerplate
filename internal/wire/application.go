package wire

import (
	"goilerplate/internal/application/register"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/password"
)

// ApplicationServices contains the services that orchestrate a flow across several domains.
// A flow contained within one domain does not belong here: it goes from the handler straight
// to that domain's Usecase. See docs/guides/architecture.md.
type ApplicationServices struct {
	RegisterSvc register.ApplicationService
}

func WireApplicationServices(app *bootstrap.App, repos *Repositories) *ApplicationServices {
	txManager := transaction.NewGormTransaction(app.DB.GDB)

	return &ApplicationServices{
		RegisterSvc: register.NewApplicationService(
			txManager,
			repos.UserRepo,
			repos.RoleRepo,
			repos.UserRoleRepo,
			// No common-password list ships with the boilerplate; supply a CommonChecker here
			// to enable the NIST 800-63B check (roadmap F1).
			password.NewPolicy(nil),
		),
	}
}
