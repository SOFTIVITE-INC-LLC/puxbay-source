package logger

import (
	"context"
	"fmt"

	"github.com/softivite/puxbay/internal/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"gorm.io/gorm"
)

// InitTracer initializes an OpenTelemetry tracer provider with a custom DB exporter.
func InitTracer(serviceName string, db *gorm.DB, hub *websocket.Hub) (*sdktrace.TracerProvider, error) {
	// Create a DB exporter (persists traces to database and broadcasts via websocket)
	exporter := NewDBExporter(db, hub)

	// Create a resource describing this application
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create the TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set it as global so that instrumentation libraries can use it
	otel.SetTracerProvider(tp)

	return tp, nil
}
