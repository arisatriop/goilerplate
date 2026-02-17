package router

import (
	"context"
	"goilerplate/internal/bootstrap"
	"goilerplate/internal/wire"
	"goilerplate/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RouteRegistry struct {
	App   *bootstrap.App
	Wired *wire.ApplicationContainer
}

func NewRouteRegistry(app *bootstrap.App, wired *wire.ApplicationContainer) *RouteRegistry {
	return &RouteRegistry{
		App:   app,
		Wired: wired,
	}
}

func (r *RouteRegistry) index(ctx *fiber.Ctx) error {
	return ctx.SendString("Welcome to Goilerplate!")
}

func (r *RouteRegistry) healthCheck(ctx *fiber.Ctx) error {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": utils.Now().Format(time.RFC3339),
		"service":   r.App.Config.App.Name,
		"version":   r.App.Config.App.Version,
		"checks":    make(map[string]interface{}),
	}

	checks := response["checks"].(map[string]interface{})
	allHealthy := true

	// Check PostgreSQL connection
	if r.App.DB.PgxDB != nil {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := r.App.DB.PgxDB.Ping(timeoutCtx); err != nil {
			checks["postgresql"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			allHealthy = false
		} else {
			checks["postgresql"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	// Check GORM connection
	if r.App.DB.GDB != nil {
		if sqlDB, err := r.App.DB.GDB.DB(); err != nil {
			checks["gorm"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			allHealthy = false
		} else {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := sqlDB.PingContext(timeoutCtx); err != nil {
				checks["gorm"] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				allHealthy = false
			} else {
				checks["gorm"] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}
	}

	// Check Redis connection
	if r.App.Redis != nil {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := r.App.Redis.Ping(timeoutCtx).Err(); err != nil {
			checks["redis"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			allHealthy = false
		} else {
			checks["redis"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	// Set overall status
	if !allHealthy {
		response["status"] = "unhealthy"
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

// Register sets up all the routes and middleware for the application.
func (r *RouteRegistry) Register() {
	http := r.App.WebServer.Use(r.Wired.Middleware.RequestLogger.LogRequest())
	http.Use(r.Wired.Middleware.Recover)

	http.Get("/", r.index)
	http.Get("/health", r.healthCheck)

	(&InternalRouteRegistry{
		App:   r.App,
		Wired: r.Wired,
	}).register(http)

	(&PartnerRouteRegistry{
		App:   r.App,
		Wired: r.Wired,
	}).register(http)

	(&PublicRouteRegistry{
		App:   r.App,
		Wired: r.Wired,
	}).register(http)
}
