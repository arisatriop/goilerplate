package wire

import (
	"goilerplate/internal/bootstrap"
)

// ApplicationContainer holds all wired dependencies
type ApplicationContainer struct {
	Infrastructure      *Infrastructure
	Repositories        *Repositories
	UseCases            *UseCases
	ApplicationServices *ApplicationServices
	Handlers            *Handlers
	Middleware          *Middleware
}

// Init wires all dependencies following clean architecture layers
func Init(app *bootstrap.App) *ApplicationContainer {
	// Layer 1: Infrastructure Layer (External services, filesystem, etc.)
	infrastructure := WireInfrastructure(app)

	// Layer 2: Repository Layer (Data access)
	repositories := WireRepositories(app)

	// Layer 3: Use Case Layer (Domain/Business Logic)
	useCases := WireUseCases(app, repositories, infrastructure)

	// Layer 4: Application Service Layer (Multi-domain orchestration)
	applicationServices := WireApplicationServices(app, repositories, useCases, infrastructure)

	// Layer 5: Handler Layer (Delivery/Presentation)
	handlers := WireHandlers(app, useCases, applicationServices, infrastructure)

	// Layer 5: Middleware Layer
	middleware := WireMiddleware(app.Config, repositories, infrastructure)

	return &ApplicationContainer{
		Infrastructure:      infrastructure,
		Repositories:        repositories,
		UseCases:            useCases,
		ApplicationServices: applicationServices,
		Handlers:            handlers,
		Middleware:          middleware,
	}
}
