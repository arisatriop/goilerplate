package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	LogLabel = "outgoing-request-log"
)

type LoggingRoundTripper struct {
	Proxied http.RoundTripper
}

func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := utils.Now()
	startTime := start.Format(time.RFC3339Nano)

	// Ensure request ID
	requestID := req.Header.Get(constants.HeaderRequestID)
	if requestID == "" {
		requestID = uuid.New().String()
		req.Header.Set(constants.HeaderRequestID, requestID)
	}

	// Capture request body
	var requestPayload interface{}
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		requestPayload = parseBody(reqBody, req.Header.Get("Content-Type"))
	}

	// Execute request
	resp, err := lrt.Proxied.RoundTrip(req)

	duration := time.Since(start)
	endTime := utils.Now().Format(time.RFC3339Nano)

	// Capture response details
	var statusCode int
	var responseSize int
	var responseBody interface{}
	var responseHeaders map[string]interface{}
	var responseMessage string

	if resp != nil {
		statusCode = resp.StatusCode
		responseHeaders = make(map[string]interface{})
		for k, v := range resp.Header {
			if len(v) == 1 {
				responseHeaders[k] = v[0]
			} else {
				responseHeaders[k] = v
			}
		}

		if resp.Body != nil {
			respBodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes))
			responseSize = len(respBodyBytes)
			responseBody = parseBody(respBodyBytes, resp.Header.Get("Content-Type"))

			if jsonResp, ok := responseBody.(map[string]interface{}); ok {
				if msg, exists := jsonResp["message"]; exists {
					if msgStr, ok := msg.(string); ok {
						responseMessage = msgStr
					}
				}
			}
		}
	}

	// Prepare request headers for logging
	requestHeaders := make(map[string]interface{})
	for k, v := range req.Header {
		if len(v) == 1 {
			requestHeaders[k] = v[0]
		} else {
			requestHeaders[k] = v
		}
	}

	// Prepare query params
	queryParams := make(map[string]string)
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	// Log attributes
	logAttrs := []slog.Attr{
		slog.String("label", LogLabel),
		slog.String("request_id", requestID),
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.String("path", req.URL.Path),
		slog.String("host", req.URL.Host),
		slog.Any("query_params", queryParams),
		slog.Any("request_headers", requestHeaders),
		slog.Any("request_payload", requestPayload),
		slog.Int("status", statusCode),
		slog.Any("response_headers", responseHeaders),
		slog.Int("response_size", responseSize),
		slog.Any("response_body", responseBody),
		slog.String("response_message", responseMessage),
		slog.String("start_time", startTime),
		slog.String("end_time", endTime),
		slog.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
	}

	if err != nil {
		logAttrs = append(logAttrs, slog.String("error", err.Error()))
	}

	// Always log as INFO
	if strings.ToLower(os.Getenv("APP_ENV")) != "local" {
		slog.LogAttrs(context.Background(), slog.LevelInfo, "Outgoing HTTP request", logAttrs...)
	}

	return resp, err
}

func NewClient(timeout time.Duration) *http.Client {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &http.Client{
		Transport: &LoggingRoundTripper{
			Proxied: http.DefaultTransport,
		},
		Timeout: timeout,
	}
}

// parseBody parses body data for logging with content type awareness (duplicated from request logger for independence)
func parseBody(body []byte, contentType string) interface{} {
	if len(body) == 0 {
		return nil
	}

	// Clean up content type
	mediaType, _, _ := mime.ParseMediaType(contentType)

	switch {
	case mediaType == "application/json":
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			return jsonData
		}
		return string(body)

	case mediaType == "multipart/form-data":
		return map[string]interface{}{
			"content_type": contentType,
			"message":      "multipart/form-data content (binary data not logged)",
			"size_bytes":   len(body),
		}

	case mediaType == "application/octet-stream" ||
		strings.HasPrefix(mediaType, "image/") ||
		strings.HasPrefix(mediaType, "video/") ||
		strings.HasPrefix(mediaType, "audio/"):
		return map[string]interface{}{
			"content_type": contentType,
			"message":      "binary/media content not logged",
			"size_bytes":   len(body),
		}

	case mediaType == "application/x-www-form-urlencoded":
		return string(body)

	default:
		if len(body) > 1000 {
			return map[string]interface{}{
				"content_type": contentType,
				"message":      "content truncated (too large)",
				"size_bytes":   len(body),
				"preview":      string(body[:1000]),
				"truncated_at": 1000,
			}
		}
		return string(body)
	}
}
