package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goilerplate/config"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Log struct {
	Config        *config.Config
	ElasticClient *elasticsearch.Client
}

func NewLog(cfg *config.Config, es *elasticsearch.Client) *Log {
	return &Log{
		Config:        cfg,
		ElasticClient: es,
	}
}

type IncomingLog struct {
	Timestamp   string            `json:"timestamp"`
	Status      int               `json:"status"`
	Method      string            `json:"method"`
	DurationStr string            `json:"duration_str"`
	DurationMs  float64           `json:"duration_ms"`
	TraceID     string            `json:"trace_id"`
	UserID      string            `json:"user_id,omitempty"`
	IP          string            `json:"ip"`
	Path        string            `json:"path"`
	FullURL     string            `json:"full_url"`
	UserAgent   string            `json:"user_agent"`
	Headers     map[string]string `json:"headers,omitempty"`
}

func (l *Log) IncomingReqestLog(log *logrus.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		// Trace ID (from header or generate new)
		traceID := extractTraceID(c.Get("X-Request-ID"))
		c.Locals("trace_id", traceID)

		// Important headers
		headers := map[string]string{
			"X-Request-ID":    traceID,
			"X-Forwarded-For": c.Get("X-Forwarded-For"),
			"Authorization":   maskToken(c.Get("Authorization")),
		}

		// Continue to handler
		err := c.Next()
		duration := time.Since(start)

		if l.Config.Elasticsearch.Enabled {
			logEntry := IncomingLog{
				Timestamp:   time.Now().Format(time.RFC3339),
				Status:      c.Response().StatusCode(),
				Method:      c.Method(),
				DurationStr: duration.String(), // e.g. "12ms"
				DurationMs:  float64(duration.Microseconds()) / 1000.0,
				TraceID:     traceID,
				UserID:      GetUserID(c),
				IP:          c.IP(),
				Path:        c.Path(),
				FullURL:     fmt.Sprintf("%s://%s%s", c.Protocol(), c.Hostname(), c.OriginalURL()),
				UserAgent:   c.Get("User-Agent"),
				Headers:     headers,
			}
			go sendToElastic(l.Config, l.ElasticClient, logEntry)
		}

		if l.Config.App.Env == "local" {
			log.WithFields(logrus.Fields{
				"status":       c.Response().StatusCode(),
				"method":       c.Method(),
				"duration_str": duration.String(),
				"trace_id":     traceID,
				"ip":           c.IP(),
				"path":         c.Path(),
			}).Info("Incoming request")
			fmt.Println()
		}

		return err
	}
}

func sendToElastic(cfg *config.Config, es *elasticsearch.Client, doc interface{}) {
	data, err := json.Marshal(doc)
	if err != nil {
		logrus.Errorf("failed to marshal log: %v", err)
		return
	}

	res, err := es.Index(cfg.Elasticsearch.ApiIncomingLogIndex, bytes.NewReader(data))
	if err != nil {
		logrus.Errorf("failed to send log to Elasticsearch: %v", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		logrus.Errorf("elasticsearch indexing error: %s", res.String())
	}
}

func extractTraceID(traceID string) string {
	if traceID == "" {
		return uuid.New().String()
	}
	return traceID
}

func maskToken(token string) string {
	if token == "" {
		return ""
	}
	prefix := ""
	if strings.HasPrefix(token, "Bearer ") {
		prefix = "Bearer "
		token = strings.TrimPrefix(token, "Bearer ")
	}
	n := len(token) / 2
	visible := token[n:]
	masked := strings.Repeat("*", len(token)-n)
	return prefix + masked + visible
}
