package bootstrap

import (
	"errors"
	"time"

	"goilerplate/config"
	"goilerplate/internal/delivery/http/middleware"
	"goilerplate/pkg/response"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func NewFiber(cfg *config.Config) *fiber.App {
	var app = fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: NewErrorHandler(),
		Prefork:      cfg.Server.Prefork,

		// Fiber buffers the whole body before a handler sees it, so this is the ceiling on
		// memory one request can cost. It was hardcoded at 100MB on a service whose own
		// filesystem.max_file_size defaults to 200KB — the upload validation never got the
		// chance to refuse anything. Defaults to 10MB when unset.
		BodyLimit: cfg.Server.BodyLimitOrDefault(),

		// Without these a stalled client holds a connection indefinitely. They are read from
		// config rather than hardcoded because the right ceiling depends on what the service
		// does; they fall back to a bounded default rather than to "no limit" when unset.
		ReadTimeout:  cfg.Server.ReadTimeoutOrDefault(),
		WriteTimeout: cfg.Server.WriteTimeoutOrDefault(),
		IdleTimeout:  cfg.Server.IdleTimeoutOrDefault(),

		// c.IP() honours the forwarding header only when the request actually came from a
		// configured proxy. With no proxies listed it always reports the peer that connected,
		// so a client cannot pick its own IP by sending X-Forwarded-For — which would let it
		// walk around every per-IP rate limit and poison the audit trail.
		EnableTrustedProxyCheck: true,
		TrustedProxies:          cfg.Server.TrustedProxies,
		ProxyHeader:             cfg.Server.ProxyHeaderOrDefault(),
	})

	app.Use(middleware.Recover())
	app.Use(helmet.New(helmet.Config{
		XSSProtection:         "0", // the legacy filter is itself a vulnerability; modern browsers ignore it
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		HSTSMaxAge:            hstsMaxAge(cfg),
		HSTSExcludeSubdomains: false,
	}))
	if cfg.OTel.Enabled {
		app.Use(otelfiber.Middleware())
	}
	// CORS is off unless asked for. It relaxes a browser default that exists for a reason, and
	// a service with no browser client has no use for it.
	//
	// cors.New panics on an insecure or malformed configuration rather than returning an error,
	// so config.Validate checks the same things first and fails startup with a message that
	// names the key. By the time this runs, the configuration is known to be acceptable.
	if cfg.Server.EnableCORS {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.Server.CORS.AllowOrigin,
			AllowMethods:     cfg.Server.CORS.AllowMethodsOrDefault(),
			AllowHeaders:     cfg.Server.CORS.AllowHeaders,
			AllowCredentials: cfg.Server.CORS.AllowCredentials,
			ExposeHeaders:    cfg.Server.CORS.ExposeHeaders,
			MaxAge:           cfg.Server.CORS.MaxAge,
		}))
	}

	return app
}

func NewErrorHandler() fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		// A fiber.Error is Fiber's own answer — no route (404), wrong method (405), a body over the
		// limit (413) — and its message is safe to show. errors.As rather than a type assertion,
		// so one that was wrapped keeps its status.
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return response.FailStatus(ctx, fiberErr.Code, fiberErr.Message)
		}

		// Anything else is an error a handler returned instead of answering. It used to be sent
		// to the client verbatim as a 500 — driver errors and all. Now a client error keeps its
		// status and code, and everything else is logged and answered generically.
		return response.HandleError(ctx, err)
	}
}

// hstsMaxAge returns the Strict-Transport-Security lifetime, or 0 to omit the header.
// It stays off by default: sending it from a service reached over plain HTTP would tell
// browsers to refuse the only scheme that works.
func hstsMaxAge(cfg *config.Config) int {
	if !cfg.Server.HSTS {
		return 0
	}
	return int((365 * 24 * time.Hour).Seconds())
}
