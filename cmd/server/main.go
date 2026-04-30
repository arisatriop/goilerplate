// @title           Goilerplate API
// @version         1.0
// @description     Go backend boilerplate using Clean Architecture. Provides a ready-to-use foundation for REST APIs with auth, RBAC, file uploads, and multi-database support.
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
	_ "goilerplate/docs"
	"net"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	app := bootstrap.Init()

	// 2. Wire all dependencies in dedicated wire package
	wired := wire.Init(app)

	// 3. Setup HTTP routes
	router.NewRouteRegistry(app, wired).Register()

	// 4. Register gRPC services
	wired.GrpcHandlers.ServiceRegistry.Register(app.GrpcServer)

	// 5. Start the servers
	start(app)
}

func start(app *bootstrap.App) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		webPort := app.Config.Server.Port
		if err := app.WebServer.Listen(fmt.Sprintf(":%d", webPort)); err != nil {
			fmt.Printf("Failed to start HTTP server: %v\n", err)
			stop()
		}
	}()

	if app.Config.GRPC.Enabled {
		go func() {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", app.Config.GRPC.Port))
			if err != nil {
				fmt.Printf("Failed to listen for gRPC: %v\n", err)
				stop()
				return
			}
			fmt.Printf("gRPC server listening on :%d\n", app.Config.GRPC.Port)
			if err := app.GrpcServer.Serve(lis); err != nil {
				fmt.Printf("Failed to start gRPC server: %v\n", err)
				stop()
			}
		}()
	}

	<-ctx.Done()

	fmt.Printf("\n\nShutting down server...\n\n")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if app.TracerProvider != nil {
		if err := app.TracerProvider.Shutdown(timeoutCtx); err != nil {
			app.Log.Error("Error shutting down tracer provider", "error", err)
		}
	}

	if app.MeterProvider != nil {
		if err := app.MeterProvider.Shutdown(timeoutCtx); err != nil {
			app.Log.Error("Error shutting down meter provider", "error", err)
		}
	}

	gracefulShutdown(timeoutCtx, app)
}

func gracefulShutdown(ctx context.Context, app *bootstrap.App) {
	done := make(chan error, 1)
	go func() {
		done <- app.WebServer.Shutdown()
	}()

	select {
	case err := <-done:
		if err != nil {
			app.Log.Error("Error during Fiber shutdown", "error", err)
		} else {
			fmt.Printf("Fiber server shutdown successfully\n")
		}
	case <-ctx.Done():
		app.Log.Warn("Fiber shutdown timeout expired, forcing shutdown")
	}

	// Close database connections
	if app.DB.GDB != nil {
		if gdb, err := app.DB.GDB.DB(); err != nil {
			app.Log.Error("Error getting underlying sql.DB from GORM", "error", err)
		} else {
			if err := gdb.Close(); err != nil {
				app.Log.Error("Error closing GORM connection", "error", err)
			} else {
				fmt.Printf("GORM connection closed successfully\n")
			}
		}
	}

	if app.DB.PgxDB != nil {
		app.DB.PgxDB.Close()
		fmt.Printf("PostgreSQL connection pool closed successfully\n")
	}

	if app.DB.MysqlDB != nil {
		if err := app.DB.MysqlDB.Close(); err != nil {
			app.Log.Error("Error closing MysqlDB", "error", err)
		} else {
			fmt.Printf("MysqlDB connection closed successfully\n")
		}
	}

	if app.Redis != nil {
		if err := app.Redis.Close(); err != nil {
			app.Log.Error("Error closing Redis connection", "error", err)
		} else {
			fmt.Printf("Redis connection closed successfully\n")
		}
	}

	app.GrpcServer.GracefulStop()
	fmt.Printf("gRPC server shutdown successfully\n")

	fmt.Printf("\nServer shutting down gracefully...\n")
}
