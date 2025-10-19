package trace

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/akozadaev/go_todo_service/config"
	"github.com/gin-gonic/gin"
	"github.com/gookit/goutil/netutil/httpctype"
	"github.com/gookit/goutil/netutil/httpheader"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	trace2 "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var TraceClient *Tracer

const AttributeReqBody = "request.body"

const (
	AttributeRespHttpCode = "http.status_code"
	AttributeRespErrMsg   = "error.message"
)

type Tracer struct {
	tp          *tracesdk.TracerProvider
	cfg         *config.TraceConfig
	IsEnabled   bool
	ServiceName string
}

// NewTraceClient - создание клиента трассировки
func NewTraceClient() (*Tracer, error) {
	t := &Tracer{}
	// config init
	if err := t.initTraceConfig(); err != nil {
		return nil, err
	}

	if !t.cfg.IsTraceEnabled {
		return t, nil
	}

	// Create the Jaeger exporter
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(t.cfg.Url),
		),
	)

	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		// tracesdk.WithSampler(),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(t.cfg.ServiceName),
			// attribute.String("environment", "development"),
			// attribute.Int64("ID", 1),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		//	error handler
	}))

	t.tp = tp
	TraceClient = t

	return t, nil
}

func (t *Tracer) Shutdown(ctx context.Context) error {
	return t.tp.Shutdown(ctx)
}

// InjectHttpTraceId сгенерированный  метод для тестирования
func (t *Tracer) InjectHttpTraceId(ctx context.Context, req *http.Request) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// MiddleWareTrace сгенерированный  метод для тестирования
func (t *Tracer) MiddleWareTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		if t == nil || !t.cfg.IsTraceEnabled {
			c.Next()

			return
		}

		parentCtx, span := t.CreateSpan(c.Request.Context(), "["+c.Request.Method+"] "+c.FullPath(), "middleware")

		defer span.End()

		if t.cfg.IsHttpBodyEnabled {
			if !strings.HasPrefix(c.GetHeader(httpheader.ContentType), httpctype.MIMEDataForm) {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				span.SetAttributes(attribute.String(AttributeReqBody, string(bodyBytes)))
			}
		}

		c.Request = c.Request.WithContext(parentCtx)
		c.Next()

		// парсинг ошибок
		span.SetAttributes(attribute.Int(AttributeRespHttpCode, c.Writer.Status()))
		{
			excep := c.Keys["exception"]

			if v, ok := excep.(error); ok {
				span.SetAttributes(attribute.String(AttributeRespErrMsg, v.Error()))
			}
		}
	}
}

// CreateSpan сгенерированный  метод для тестирования
func (t *Tracer) CreateSpan(ctx context.Context, name string, fun string) (context.Context, trace2.Span) {
	if t == nil || t.tp == nil {
		return context.Background(), noop.Span{}
	}

	return t.tp.Tracer(t.ServiceName).Start(ctx, name)
}

// initTraceConfig -  инициализирует конфиг трассировки, читает  из файла  .env переменки
func (t *Tracer) initTraceConfig() error {
	serverConfig, err := config.Load()
	if err != nil {
		log.Error().Stack().Err(err)
	}

	traceCfg := &config.TraceConfig{}
	traceCfg.IsTraceEnabled = serverConfig.Trace.IsTraceEnabled
	traceCfg.IsHttpBodyEnabled = serverConfig.Trace.IsHttpBodyEnabled
	traceCfg.Url = serverConfig.Trace.Url
	traceCfg.ServiceName = serverConfig.Trace.ServiceName

	t.cfg = traceCfg
	t.ServiceName = traceCfg.ServiceName
	t.IsEnabled = traceCfg.IsTraceEnabled

	return nil
}
