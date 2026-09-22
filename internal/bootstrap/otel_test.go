package bootstrap

import (
	"context"
	"testing"

	"goilerplate/pkg/redact"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestRedactSpanProcessor_OnStart(t *testing.T) {
	// Arrange
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(redactSpanProcessor{}),
		sdktrace.WithSpanProcessor(recorder),
	)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	// Act
	_, span := tp.Tracer("test").Start(context.Background(), "GET /reset",
		oteltrace.WithAttributes(
			attribute.String("http.url", "http://localhost/reset?token=abc&page=1"),
			attribute.String("http.target", "/reset?token=abc&page=1"),
			attribute.String("http.method", "GET"),
		),
	)
	span.End()

	// Assert
	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	attrs := map[attribute.Key]string{}
	for _, attr := range spans[0].Attributes() {
		attrs[attr.Key] = attr.Value.String()
	}
	assert.Equal(t, "http://localhost/reset?token="+redact.Mask+"&page=1", attrs["http.url"])
	assert.Equal(t, "/reset?token="+redact.Mask+"&page=1", attrs["http.target"])
	assert.Equal(t, "GET", attrs["http.method"])
}
