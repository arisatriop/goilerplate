package bootstrap

import (
	"context"
	"fmt"
	"goilerplate/config"
	"goilerplate/pkg/redact"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// NewMeterProvider registers a Prometheus exporter served at /metrics. Call it only when
// otel.enabled is true.
func NewMeterProvider(cfg *config.Config) (*sdkmetric.MeterProvider, error) {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.App.Name),
		semconv.ServiceVersionKey.String(cfg.App.Version),
		semconv.DeploymentEnvironmentKey.String(cfg.App.Env),
	)

	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("creating Prometheus exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(mp)
	return mp, nil
}

// NewTracerProvider exports traces over OTLP gRPC. Call it only when otel.enabled is true.
func NewTracerProvider(cfg *config.Config) (*sdktrace.TracerProvider, error) {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.App.Name),
		semconv.ServiceVersionKey.String(cfg.App.Version),
		semconv.DeploymentEnvironmentKey.String(cfg.App.Env),
	)

	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTel.Endpoint),
	}
	if cfg.OTel.Insecure {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(context.Background(), exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(redactSpanProcessor{}),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// urlAttributeKeys are span attributes that may carry a query string (old and current HTTP semconv).
var urlAttributeKeys = map[attribute.Key]struct{}{
	"http.url":    {},
	"http.target": {},
	"url.full":    {},
	"url.query":   {},
}

// redactSpanProcessor masks sensitive query parameters in URL attributes when a span starts,
// before any exporter can read them. Attributes set later with the same key are not rewritten.
type redactSpanProcessor struct{}

func (redactSpanProcessor) OnStart(_ context.Context, span sdktrace.ReadWriteSpan) {
	redactor := redact.Default()
	for _, attr := range span.Attributes() {
		if _, ok := urlAttributeKeys[attr.Key]; !ok || attr.Value.Type() != attribute.STRING {
			continue
		}

		value := attr.Value.AsString()
		redacted := redactor.URL(value)
		if attr.Key == "url.query" {
			redacted = redactor.EncodedQuery(value)
		}
		if redacted != value {
			span.SetAttributes(attr.Key.String(redacted))
		}
	}
}

func (redactSpanProcessor) OnEnd(sdktrace.ReadOnlySpan) {}

func (redactSpanProcessor) Shutdown(context.Context) error { return nil }

func (redactSpanProcessor) ForceFlush(context.Context) error { return nil }
