package wire

import (
	"os"

	"goilerplate/config"
	"goilerplate/internal/bootstrap"
	grpcmiddleware "goilerplate/internal/delivery/grpc/middleware"
	"goilerplate/internal/domain/auth"

	"google.golang.org/grpc"
)

// ApplicationContainer holds all wired dependencies
type ApplicationContainer struct {
	Infrastructure      *Infrastructure
	Repositories        *Repositories
	UseCases            *UseCases
	ApplicationServices *ApplicationServices
	Handlers            *Handlers
	GrpcHandlers        *GrpcHandlers // nil when grpc.enabled is false
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

	// The gRPC server is built here rather than in bootstrap: its auth interceptor reuses the
	// same token validator and session store as HTTP, and interceptors can only be handed to
	// grpc.NewServer, so it cannot exist before the infrastructure layer does.
	var grpcHandlers *GrpcHandlers
	if app.Config.GRPC.Enabled {
		grpcHandlers = WireGrpcHandlers(useCases)
		app.GrpcServer = wireGrpcServer(app, repositories, infrastructure)
	}

	return &ApplicationContainer{
		Infrastructure:      infrastructure,
		Repositories:        repositories,
		UseCases:            useCases,
		ApplicationServices: applicationServices,
		Handlers:            handlers,
		GrpcHandlers:        grpcHandlers,
		Middleware:          middleware,
	}
}

// wireGrpcServer builds the gRPC server with its auth interceptor.
func wireGrpcServer(app *bootstrap.App, repos *Repositories, infra *Infrastructure) *grpc.Server {
	cfg := app.Config
	strictRevocation := cfg.Auth.RevocationMode() == config.RevocationStrict
	sessionService := auth.NewSessionService(repos.AuthRepo, infra.SessionStore, strictRevocation)

	server, err := bootstrap.NewGrpcServer(cfg, grpcmiddleware.NewAuth(cfg.GRPC.Auth, infra.JWTService, sessionService))
	if err != nil {
		// Config validation has already checked that the files are named; failing to read them
		// is a deployment fault that must stop startup rather than serve without TLS.
		app.Log.Error("failed to create gRPC server", "error", err)
		os.Exit(1)
	}

	return server
}
