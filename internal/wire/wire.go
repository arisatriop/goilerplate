package wire

import (
	"goilerplate/internal/bootstrap"
)

// ApplicationContainer holds all wired dependencies
type ApplicationContainer struct {
	Repositories *Repositories
	UseCases     *UseCases
	Handlers     *Handlers
	Middleware   *Middleware
}

// Init wires all dependencies following clean architecture layers
func Init(app *bootstrap.App) *ApplicationContainer {
	// Layer 1: Repository Layer (Infrastructure implements Domain interfaces)
	repositories := WireRepositories(app)

	// Layer 2: Use Case Layer (Domain/Business Logic)
	useCases := WireUseCases(app, repositories)

	// Layer 3: Handler Layer (Delivery/Presentation)
	handlers := WireHandlers(app, useCases)

	// Layer 4: Middleware Layer
	middleware := WireMiddleware(app, repositories)

	return &ApplicationContainer{
		Repositories: repositories,
		UseCases:     useCases,
		Handlers:     handlers,
		Middleware:   middleware,
	}
}
