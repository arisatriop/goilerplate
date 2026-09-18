package bootstrap

import (
	"time"

	"goilerplate/config"
	"goilerplate/internal/delivery/http/middleware"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func NewFiber(cfg *config.Config) *fiber.App {
	var app = fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: NewErrorHandler(),
		Prefork:      cfg.Server.Prefork,
		BodyLimit:    100 * 1024 * 1024, // 100MB limit for file uploads

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
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: "*",
	// 	AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	// 	AllowMethods: "*",
	// }))

	return app
}

func NewErrorHandler() fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		return ctx.Status(code).JSON(fiber.Map{
			"message": err.Error(),
		})
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
