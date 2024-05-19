package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func Fiber() fiber.Config {
	return fiber.Config{
		Prefork:      false,            // Enable prefork to utilize all CPU cores
		ReadTimeout:  10 * time.Second, // Maximum duration for reading the entire request
		WriteTimeout: 10 * time.Second, // Maximum duration before timing out writes of the response
		IdleTimeout:  5 * time.Second,  // Maximum amount of time to wait for the next request when keep-alives are enabled

		// Enable as needed
		// ServerHeader:              "Fiber",                   // Set a custom server header
		// StrictRouting:             true,                      // Enforce strict routing
		// CaseSensitive:             true,                      // Enforce case sensitivity
		// Immutable: 				  true, 					 // Provide a new copy of request and response contexts
		// UnescapePath:              true,                      // Enable unescape path
		// ETag:                      true,                      // Enable ETag generation
		// BodyLimit:                 4 * 1024 * 1024,           // Set the maximum allowed size for a request body
		// Concurrency:               256 * 1024,                // Maximum number of concurrent connections
		// ReadBufferSize:            4096,                      // Per-connection buffer size for reading requests
		// WriteBufferSize:           4096,                      // Per-connection buffer size for writing responses
		// CompressedFileSuffix:      ".fiber.gz",               // Suffix to add to the file name when serving compressed files
		// ProxyHeader:               "X-Forwarded-For",         // Header key for getting the client's real IP address
		// GETOnly:                   false,                     // Allow only GET method
		// ErrorHandler:              fiber.DefaultErrorHandler, // Custom error handler
		// DisableKeepalive:          false,                     // Disable keep-alive support
		// DisableDefaultDate:        false,                     // Disable default date header
		// DisableDefaultContentType: false,                     // Disable default Content-Type header
		// DisableHeaderNormalizing:  false,                     // Disable header normalization
		// DisableStartupMessage:     false,                     // Disable startup message
		// ReduceMemoryUsage:         false,                     // Reduce memory usage by decreasing buffers and not pre-allocating space
	}
}
