package wire

import (
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/example"
	"goilerplate/internal/domain/role"
	"goilerplate/internal/domain/template"
	"goilerplate/internal/domain/user"
	"goilerplate/internal/domain/userrole"

	"goilerplate/internal/infrastructure/repository"
)

// Repositories contains all repository implementations
type Repositories struct {
	AuthRepo     auth.Repository
	RoleRepo     role.Repository
	UserRepo     user.Repository
	UserRoleRepo userrole.Repository
	TemplateRepo template.Repository
	ExampleRepo  example.Repository
}

// WireRepositories creates all repository implementations
func WireRepositories(app *bootstrap.App) *Repositories {
	db := app.DB.GDB
	return &Repositories{
		AuthRepo:     repository.NewAuth(db),
		RoleRepo:     repository.NewRole(db),
		UserRepo:     repository.NewUser(db),
		UserRoleRepo: repository.NewUserRole(db),
		TemplateRepo: repository.NewTemplate(db),
		ExampleRepo:  repository.NewExample(db),
	}
}
