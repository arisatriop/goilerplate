package middleware

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	_ "github.com/gofiber/fiber/v2/middleware/basicauth"
	_ "github.com/gofiber/fiber/v2/middleware/compress"
	_ "github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/gofiber/fiber/v2/middleware/csrf"
	_ "github.com/gofiber/fiber/v2/middleware/etag"
	_ "github.com/gofiber/fiber/v2/middleware/filesystem"
	_ "github.com/gofiber/fiber/v2/middleware/limiter"
	_ "github.com/gofiber/fiber/v2/middleware/proxy"
	_ "github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/gofiber/fiber/v2/middleware/redirect"
	_ "github.com/gofiber/fiber/v2/middleware/timeout"
)

func Fiber(app *fiber.App) {
	// RequestID middleware
	app.Use(requestid.New())

	// Logger middleware
	app.Use(logger.New(logger.Config{
		Next:          nil,
		Done:          nil,
		Format:        "${time} | ${locals:requestid} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat:    "2006-01-02 15:04:05.000",
		TimeZone:      "Local",
		TimeInterval:  500 * time.Millisecond,
		Output:        os.Stdout,
		DisableColors: false,
	}))

	file, _ := os.OpenFile("./logs/app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${locals:requestid} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05.000",
		TimeZone:   "Local",
		Output:     file,
	}))

	// app.Use(recover.New())

	// // BasicAuth middleware
	// app.Use(basicauth.New(basicauth.Config{
	// 	Users: map[string]string{
	// 		"user": "pass",
	// 	},
	// }))

	// // Compress middleware
	// app.Use(compress.New(compress.Config{
	// 	Level: compress.LevelBestSpeed, // You can use LevelBestSpeed, LevelBestCompression, etc.
	// }))

	// // CORS middleware
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: "*",
	// 	AllowHeaders: "Origin, Content-Type, Accept",
	// }))

	// // CSRF middleware
	// app.Use(csrf.New())

	// // ETag middleware
	// app.Use(etag.New())

	// // Filesystem middleware
	// app.Use("/static", filesystem.New(filesystem.Config{
	// 	Root: http.Dir("./public"),
	// }))

	// // Limiter middleware
	// app.Use(limiter.New(limiter.Config{
	// 	Max:        10,
	// 	Expiration: 30 * time.Second,
	// }))

	// // Logger middleware
	// app.Use(logger.New())

	// // Proxy middleware
	// app.Use(proxy.New(proxy.Config{
	// 	// For example, redirect all requests to google.com
	// 	Next: nil,
	// }))

	// // Recover middleware
	// app.Use(recover.New())

	// // Redirect middleware
	// app.Use(redirect.New(redirect.Config{
	// 	Rules: map[string]string{
	// 		"/old": "/new",
	// 	},
	// }))

}
