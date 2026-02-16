package middleware

import (
	"context"
	"encoding/json"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"
	"log/slog"
	"mime"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	LogLabel = "incoming-request-log"
)

// RequestLogger provides incoming request logging functionality
type RequestLogger struct {
}

// NewRequestLogger creates a new request logger middleware
func NewRequestLogger() *RequestLogger {
	return &RequestLogger{}
}

// LogRequest returns a Fiber middleware for logging incoming requests
func (rl *RequestLogger) LogRequest() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Generate request ID
		requestID := ctx.Get(strings.ToLower(constants.HeaderRequestID), uuid.New().String())

		userCtx := context.WithValue(ctx.UserContext(), constants.ContextKeyRequestID, requestID)
		ctx.SetUserContext(userCtx)
		ctx.Locals(string(constants.ContextKeyRequestID), requestID)

		start := utils.Now()
		startTime := start.Format(time.RFC3339Nano)

		// Capture request details
		method := ctx.Method()
		url := ctx.OriginalURL()
		path := ctx.Path()
		userAgent := ctx.Get("User-Agent")
		contentType := ctx.Get("Content-Type")
		remoteIP := ctx.IP()
		protocol := ctx.Protocol()
		hostname := ctx.Hostname()

		// Capture all request headers
		headers := make(map[string]interface{})
		for key, values := range ctx.GetReqHeaders() {
			if len(values) == 1 {
				headers[key] = values[0]
			} else {
				headers[key] = values
			}
		}

		// Capture query parameters
		queryParams := make(map[string]string)
		ctx.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			queryParams[string(key)] = string(value)
		})

		// Capture route parameters
		routeParams := make(map[string]string)
		if ctx.Route() != nil {
			for _, param := range ctx.Route().Params {
				routeParams[param] = ctx.Params(param)
			}
		}

		// Capture request body
		var requestPayload interface{}
		if body := ctx.Body(); len(body) > 0 {
			requestPayload = parseBody(body, contentType)
		}

		// Variables for defer
		var panicOccurred bool
		var panicValue interface{}

		// Defer logging after request (this runs AFTER recover middleware)
		defer func() {
			duration := time.Since(start)
			endTime := utils.Now().Format(time.RFC3339Nano)
			statusCode := ctx.Response().StatusCode()

			// Check if panic occurred by looking at context locals set by recover middleware
			if panicFlag := ctx.Locals("panic_occurred"); panicFlag != nil {
				if flag, ok := panicFlag.(bool); ok && flag {
					panicOccurred = true
					if panicVal := ctx.Locals("panic_value"); panicVal != nil {
						panicValue = panicVal
					}
				}
			}

			// Capture response headers
			responseHeaders := make(map[string]interface{})
			ctx.Response().Header.VisitAll(func(key, value []byte) {
				responseHeaders[string(key)] = string(value)
			})

			responseSize := len(ctx.Response().Body())

			// Capture response body (now including panic responses from recover middleware)
			var responseBody interface{}
			var responseMessage string
			if body := ctx.Response().Body(); len(body) > 0 {
				responseContentType := string(ctx.Response().Header.ContentType())
				responseBody = parseBody(body, responseContentType)
				if jsonResp, ok := responseBody.(map[string]interface{}); ok {
					if msg, exists := jsonResp["message"]; exists {
						if msgStr, ok := msg.(string); ok {
							responseMessage = msgStr
						}
					}
				}
			}

			// Simple log attributes - no filtering
			logAttrs := []slog.Attr{
				slog.String("label", LogLabel),
				slog.String("request_id", requestID),
				slog.String("method", method),
				slog.String("url", url),
				slog.String("path", path),
				slog.String("user_agent", userAgent),
				slog.String("content_type", contentType),
				slog.String("remote_ip", remoteIP),
				slog.String("protocol", protocol),
				slog.String("hostname", hostname),
				slog.Any("query_params", queryParams),
				slog.Any("route_params", routeParams),
				slog.Any("request_headers", headers),
				slog.Any("request_payload", requestPayload),
				slog.Int("status", statusCode),
				slog.Any("response_headers", responseHeaders),
				slog.Int("response_size", responseSize),
				slog.Any("response_body", responseBody),
				slog.String("response_message", responseMessage),
				slog.Bool("panic_occurred", panicOccurred),
				slog.String("start_time", startTime),
				slog.String("end_time", endTime),
				slog.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
			}

			// Add panic value if panic occurred
			if panicOccurred {
				logAttrs = append(logAttrs, slog.Any("panic_value", panicValue))
			}

			// Always log as INFO
			if strings.ToLower(os.Getenv("APP_ENV")) != "local" {
				slog.LogAttrs(ctx.Context(), slog.LevelInfo, "Incoming HTTP request", logAttrs...)
			}
		}()

		// Set X-Request-ID response header
		ctx.Set(constants.HeaderRequestID, requestID)

		// Process request
		return ctx.Next()
	}
}

// parseBody parses body data for logging with content type awareness
func parseBody(body []byte, contentType string) interface{} {
	if len(body) == 0 {
		return nil
	}

	// Clean up content type (remove charset, boundary, etc.)
	mediaType, _, _ := mime.ParseMediaType(contentType)

	// Handle different content types
	switch {
	case mediaType == "application/json":
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			return jsonData
		}
		// If JSON parsing fails, return as string
		return string(body)

	case mediaType == "multipart/form-data":
		// For multipart/form-data (file uploads), don't log the binary content
		return map[string]interface{}{
			"content_type": contentType,
			"message":      "multipart/form-data content (binary data not logged)",
			"size_bytes":   len(body),
		}

	case mediaType == "application/octet-stream":
		// For binary data
		return map[string]interface{}{
			"content_type": contentType,
			"message":      "binary content not logged",
			"size_bytes":   len(body),
		}

	case strings.HasPrefix(mediaType, "image/") ||
		strings.HasPrefix(mediaType, "video/") ||
		strings.HasPrefix(mediaType, "audio/"):
		// For media files
		return map[string]interface{}{
			"content_type": contentType,
			"message":      "media content not logged",
			"size_bytes":   len(body),
		}

	case mediaType == "application/x-www-form-urlencoded":
		// Form data is usually safe to log
		return string(body)

	default:
		// For text-based content types or unknown types
		if len(body) > 1000 {
			// Truncate large payloads
			return map[string]interface{}{
				"content_type": contentType,
				"message":      "content truncated (too large)",
				"size_bytes":   len(body),
				"preview":      string(body[:1000]),
				"truncated_at": 1000,
			}
		}

		// Return small payloads as string
		return string(body)
	}
}
