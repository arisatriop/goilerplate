// @title           Goilerplate API
// @version         1.0
// @description     Go backend boilerplate using Clean Architecture. Provides a ready-to-use foundation for REST APIs with auth, RBAC, file uploads, and PostgreSQL.
// @host            localhost:3000
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description Partner API key.

package main

import (
	"context"
	"fmt"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/delivery/http/router"
	"goilerplate/internal/wire"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Stamped by the linker at build time (see the Dockerfile's -ldflags). They stay "dev" and
// "unknown" for a `go build` or `go run`, so a locally built binary is distinguishable from a
// released one rather than pretending to be version 0.0.0.
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	// All times are UTC regardless of the host time zone
	time.Local = time.UTC

	app := bootstrap.Init()

	// First line out of the process. When a deploy misbehaves, the question is always which
	// commit is actually running, and a running container is the only thing that can answer it.
	app.Log.Info("starting",
		"service", app.Config.App.Name,
		"version", version,
		"commit", commit,
		"buildDate", buildDate,
		"env", app.Config.App.Env,
	)

	// 2. Wire all dependencies in dedicated wire package
	wired := wire.Init(app)

	// 3. Setup HTTP routes
	router.NewRouteRegistry(app, wired).Register()

	// 4. Register gRPC services (only when grpc.enabled)
	if app.GrpcServer != nil {
		wired.GrpcHandlers.ServiceRegistry.Register(app.GrpcServer)
	}

	// 5. Start the servers
	start(app, wired)
}

// shutdownTimeout bounds the whole drain. Every step below shares it, so the process cannot
// outlive it and be SIGKILLed by the orchestrator mid-cleanup.
const shutdownTimeout = 10 * time.Second

func start(app *bootstrap.App, wired *wire.ApplicationContainer) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Background jobs take the signal context directly, so a shutdown stops them at the same
	// moment it stops accepting requests rather than after the HTTP drain.
	if wired.CleanupJob != nil {
		go wired.CleanupJob.Run(ctx)
	}

	go serveHTTP(app, stop)
	if app.GrpcServer != nil {
		go serveGRPC(ctx, app, stop)
	}

	<-ctx.Done()
	app.Log.Info("Shutdown signal received, draining")

	// Not derived from ctx: ctx is already cancelled by the signal, so a child of it would
	// expire immediately and every step below would report a timeout it never had.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	shutdown(timeoutCtx, app)
}

func serveHTTP(app *bootstrap.App, stop context.CancelFunc) {
	addr := fmt.Sprintf(":%d", app.Config.Server.Port)
	app.Log.Info("HTTP server listening", "addr", addr)
	if err := app.WebServer.Listen(addr); err != nil {
		app.Log.Error("HTTP server stopped", "error", err)
		stop()
	}
}

func serveGRPC(ctx context.Context, app *bootstrap.App, stop context.CancelFunc) {
	addr := fmt.Sprintf(":%d", app.Config.GRPC.Port)
	// ListenConfig rather than net.Listen so a signal arriving mid-startup aborts the bind
	// instead of opening a port the process is about to abandon.
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		app.Log.Error("Failed to listen for gRPC", "addr", addr, "error", err)
		stop()
		return
	}

	app.Log.Info("gRPC server listening", "addr", addr)
	if err := app.GrpcServer.Serve(listener); err != nil {
		app.Log.Error("gRPC server stopped", "error", err)
		stop()
	}
}

// shutdown releases things in dependency order: the servers that accept work stop first, then
// telemetry is flushed, and only then are the database and cache they were using closed.
//
// The previous order closed the pools before stopping gRPC, so any call still in flight during
// a rolling deploy failed against a closed pool instead of finishing.
func shutdown(ctx context.Context, app *bootstrap.App) {
	drainServers(ctx, app)
	shutdownTelemetry(ctx, app)
	closeDependencies(app)
	app.Log.Info("Shutdown complete")
}

// drainServers stops HTTP and gRPC at the same time. Sequentially, a slow HTTP drain would eat
// the budget gRPC needs, and the two do not depend on each other.
func drainServers(ctx context.Context, app *bootstrap.App) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.WebServer.ShutdownWithContext(ctx); err != nil {
			app.Log.Error("HTTP drain did not finish cleanly", "error", err)
			return
		}
		app.Log.Info("HTTP server drained")
	}()

	if app.GrpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			drainGRPC(ctx, app)
		}()
	}

	wg.Wait()
}

// drainGRPC bounds GracefulStop, which otherwise waits for every active RPC without a deadline —
// one long-lived stream would keep the process alive past terminationGracePeriodSeconds and get
// it SIGKILLed rather than letting it exit cleanly.
func drainGRPC(ctx context.Context, app *bootstrap.App) {
	stopped := make(chan struct{})
	go func() {
		app.GrpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		app.Log.Info("gRPC server drained")
	case <-ctx.Done():
		app.Log.Warn("gRPC drain timed out, abandoning it")
		// Stop cannot rescue a stuck GracefulStop, and calling it here in the foreground would
		// hang this function instead. grpc-go's stop() takes s.mu and holds it via a deferred
		// unlock across handlersWG.Wait() (server.go:1966-1989, v1.83.1), so while a handler
		// refuses to return, the graceful call keeps that mutex and any concurrent Stop blocks
		// on s.mu.Lock() for just as long.
		//
		// So fire it and do not wait. Connections that can still be closed will be; the ones
		// held by a stuck handler are released when the process exits, moments from now. The
		// alternative is outliving terminationGracePeriodSeconds and being SIGKILLed, which
		// closes them no more gently and skips the rest of the shutdown.
		go app.GrpcServer.Stop()
	}
}

// shutdownTelemetry runs after the servers so that spans and metrics produced while draining
// are still exported.
func shutdownTelemetry(ctx context.Context, app *bootstrap.App) {
	if app.TracerProvider != nil {
		if err := app.TracerProvider.Shutdown(ctx); err != nil {
			app.Log.Error("Error shutting down tracer provider", "error", err)
		}
	}

	if app.MeterProvider != nil {
		if err := app.MeterProvider.Shutdown(ctx); err != nil {
			app.Log.Error("Error shutting down meter provider", "error", err)
		}
	}
}

// closeDependencies releases the connection pools. Called last: nothing is serving by now, so
// no request can find a closed pool.
func closeDependencies(app *bootstrap.App) {
	if app.DB.GDB != nil {
		if gdb, err := app.DB.GDB.DB(); err != nil {
			app.Log.Error("Error getting underlying sql.DB from GORM", "error", err)
		} else if err := gdb.Close(); err != nil {
			app.Log.Error("Error closing GORM connection", "error", err)
		} else {
			app.Log.Info("PostgreSQL connection pool closed")
		}
	}

	if app.Redis != nil {
		if err := app.Redis.Close(); err != nil {
			app.Log.Error("Error closing Redis connection", "error", err)
		} else {
			app.Log.Info("Redis connection closed")
		}
	}
}
