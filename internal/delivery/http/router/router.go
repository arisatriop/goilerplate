package router

import (
	"context"
	"fmt"
	"strings"
	"time"

	"goilerplate/internal/bootstrap"
	"goilerplate/internal/wire"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	fiberswagger "github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// index answers the root path.
//
// @Summary      Welcome message
// @ID           index
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string
// @Router       / [get]
func (r *RouteRegistry) index(ctx *fiber.Ctx) error {
	return ctx.SendString("Welcome to Goilerplate!")
}

// probeTimeout bounds each dependency check. A probe that hangs is a probe that tells the
// orchestrator nothing, so a slow dependency is reported as not ready rather than waited on.
const probeTimeout = 5 * time.Second

// liveResponse is the body of the liveness probe.
type liveResponse struct {
	Status    string `json:"status" example:"ok"`
	Timestamp string `json:"timestamp" example:"2026-01-02T15:04:05Z"`
}

// readyResponse is the body of the readiness probe.
type readyResponse struct {
	Status    string            `json:"status" example:"ok" enums:"ok,unavailable"`
	Timestamp string            `json:"timestamp" example:"2026-01-02T15:04:05Z"`
	Service   string            `json:"service" example:"Goilerplate"`
	Version   string            `json:"version" example:"1.0.0"`
	Checks    map[string]string `json:"checks"`
}

// live answers the liveness probe. It deliberately checks nothing: liveness asks whether the
// process should be restarted, and a database outage is not fixed by restarting this pod.
//
// @Summary      Liveness probe
// @ID           livez
// @Description  Checks nothing: answers 200 whenever the process can serve.
// @Tags         health
// @Produce      json
// @Success      200  {object}  liveResponse
// @Router       /livez [get]
func (r *RouteRegistry) live(ctx *fiber.Ctx) error {
	return ctx.JSON(liveResponse{Status: "ok", Timestamp: utils.Now().Format(time.RFC3339)})
}

// liveAliasDoc documents the alias of /livez kept for existing probes.
//
// @Summary      Liveness probe (alias of /livez)
// @ID           health
// @Tags         health
// @Produce      json
// @Success      200  {object}  liveResponse
// @Router       /health [get]
//
//nolint:unused // read only by swag
func liveAliasDoc() {}

// readyAliasDoc documents the alias of /readyz kept for existing probes.
//
// @Summary      Readiness probe (alias of /readyz)
// @ID           healthcheck
// @Tags         health
// @Produce      json
// @Success      200  {object}  readyResponse
// @Failure      503  {object}  readyResponse
// @Router       /healthcheck [get]
//
//nolint:unused // read only by swag
func readyAliasDoc() {}

// metricsDoc documents the Prometheus endpoint, which promhttp serves.
//
// @Summary      Prometheus metrics
// @ID           metrics
// @Description  Exists only with OpenTelemetry enabled.
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string
// @Router       /metrics [get]
//
//nolint:unused // read only by swag
func metricsDoc() {}

// ready answers the readiness probe: can this instance serve traffic right now.
//
// The body names each dependency and whether it is up, and nothing else. It used to embed
// err.Error() straight from GORM and Redis, and driver errors routinely carry host names,
// ports, database names and user names — free reconnaissance on an unauthenticated endpoint.
// The detail is logged server-side instead, where the operator who needs it can see it and the
// caller cannot.
//
// @Summary      Readiness probe
// @ID           readyz
// @Description  Pings PostgreSQL and Redis. The status code carries the verdict.
// @Tags         health
// @Produce      json
// @Success      200  {object}  readyResponse
// @Failure      503  {object}  readyResponse
// @Router       /readyz [get]
func (r *RouteRegistry) ready(ctx *fiber.Ctx) error {
	checks := map[string]string{}
	ready := true

	record := func(name string, err error) {
		if err != nil {
			logger.Error(ctx.UserContext(), fmt.Errorf("readiness check %q failed: %w", name, err))
			checks[name] = "unhealthy"
			ready = false
			return
		}
		checks[name] = "healthy"
	}

	if r.App.DB.GDB != nil {
		record("postgresql", r.pingPostgres(ctx.UserContext()))
	}

	if r.App.Redis != nil {
		timeoutCtx, cancel := context.WithTimeout(ctx.UserContext(), probeTimeout)
		defer cancel()

		record("redis", r.App.Redis.Ping(timeoutCtx).Err())
	}

	status := "ok"
	code := fiber.StatusOK
	if !ready {
		status = "unavailable"
		code = fiber.StatusServiceUnavailable
	}

	// The status code carries the verdict, so probes and load balancers can act without
	// parsing the body.
	return ctx.Status(code).JSON(readyResponse{
		Status:    status,
		Timestamp: utils.Now().Format(time.RFC3339),
		Service:   r.App.Config.App.Name,
		Version:   r.App.Config.App.Version,
		Checks:    checks,
	})
}

func (r *RouteRegistry) pingPostgres(ctx context.Context) error {
	sqlDB, err := r.App.DB.GDB.DB()
	if err != nil {
		return fmt.Errorf("getting sql.DB from gorm: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	if err := sqlDB.PingContext(timeoutCtx); err != nil {
		return fmt.Errorf("pinging postgres: %w", err)
	}
	return nil
}

// Register sets up all the routes and middleware for the application.
func (r *RouteRegistry) Register() {
	http := r.App.WebServer.Use(r.Wired.Middleware.Recover)
	http.Get("/", r.index)
	// /livez and /readyz are the names Kubernetes uses; /health and /healthcheck are kept as
	// aliases so existing probes and dashboards keep working.
	http.Get("/livez", r.live)
	http.Get("/health", r.live)
	http.Get("/readyz", r.ready)
	http.Get("/healthcheck", r.ready)
	if r.App.MeterProvider != nil {
		http.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	}

	if strings.ToLower(r.App.Config.App.Env) != "production" {
		http.Static("/swagger-ui", ".swagger")
		http.Get("/swaggerui/*", fiberswagger.New(fiberswagger.Config{
			URL: "/swagger-ui/openapi.json",
		}))
	}
	http.Use(r.Wired.Middleware.RequestLogger.LogRequest())

	(&InternalRouteRegistry{
		App:   r.App,
		Wired: r.Wired,
	}).register(http)

	if bootstrap.PartnerRoutesEnabled(r.App.Config) {
		(&PartnerRouteRegistry{
			App:   r.App,
			Wired: r.Wired,
		}).register(http)
	}

	(&PublicRouteRegistry{
		App:   r.App,
		Wired: r.Wired,
	}).register(http)
}
