package trace

import (
	"context"
	"fmt"

	"github.com/akozadaev/go_todo_service/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Tracer управляет жизненным циклом провайдера трассировки.
// При выключенной трассировке provider равен nil, а Shutdown безопасен.
type Tracer struct {
	provider *sdktrace.TracerProvider
}

// NewTraceClient настраивает OTLP/HTTP-экспорт, ресурс сервиса и W3C propagation.
func NewTraceClient(ctx context.Context, cfg config.TraceConfig) (*Tracer, error) {
	client := &Tracer{}
	if !cfg.Enabled {
		return client, nil
	}

	exporterOptions := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(cfg.Endpoint),
	}
	if cfg.Insecure {
		exporterOptions = append(exporterOptions, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, exporterOptions...)
	if err != nil {
		return nil, fmt.Errorf("создание OTLP-экспортера: %w", err)
	}

	res, err := resource.Merge(resource.Default(), resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		attribute.String("deployment.environment.name", cfg.Environment),
	))
	if err != nil {
		return nil, fmt.Errorf("создание ресурса OpenTelemetry: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	client.provider = provider

	return client, nil
}

// Shutdown отправляет накопленные spans и освобождает ресурсы экспортера.
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t == nil || t.provider == nil {
		return nil
	}
	return t.provider.Shutdown(ctx)
}
