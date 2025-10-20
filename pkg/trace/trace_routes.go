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
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
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

	// Create the OTLP HTTP exporter
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(t.cfg.Url),
		otlptracehttp.WithInsecure(), // для локальной разработки
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(t.cfg.ServiceName),
			semconv.ServiceVersionKey.String("1.0.0"),
			attribute.String("environment", "development"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
		tracesdk.WithSampler(tracesdk.AlwaysSample()), // для разработки - всегда семплируем
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Error().Err(err).Msg("OpenTelemetry error")
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

// MiddleWareTrace создает корневой span для каждого HTTP запроса
func (t *Tracer) MiddleWareTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		if t == nil || !t.cfg.IsTraceEnabled {
			c.Next()
			return
		}

		// Извлекаем контекст из заголовков (для распределенной трассировки)
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Создаем корневой span для HTTP запроса
		spanName := c.Request.Method + " " + c.FullPath()
		ctx, span := t.CreateSpan(ctx, spanName, "http")
		defer span.End()

		// Добавляем атрибуты к span
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.scheme", c.Request.URL.Scheme),
			attribute.String("http.host", c.Request.Host),
		)

		// Добавляем тело запроса если включено
		if t.cfg.IsHttpBodyEnabled {
			if !strings.HasPrefix(c.GetHeader(httpheader.ContentType), httpctype.MIMEDataForm) {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				span.SetAttributes(attribute.String(AttributeReqBody, string(bodyBytes)))
			}
		}

		// Передаем контекст с span дальше
		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// Добавляем информацию о ответе
		span.SetAttributes(
			attribute.Int(AttributeRespHttpCode, c.Writer.Status()),
			attribute.Int("http.response.size", c.Writer.Size()),
		)

		// Обрабатываем ошибки
		if excep := c.Keys["exception"]; excep != nil {
			if v, ok := excep.(error); ok {
				span.SetAttributes(attribute.String(AttributeRespErrMsg, v.Error()))
				span.RecordError(v)
			}
		}

		// Устанавливаем статус span в зависимости от HTTP кода
		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, "HTTP error")
		}
	}
}

// CreateSpan создает новый span с правильной иерархией
func (t *Tracer) CreateSpan(ctx context.Context, name string, fun string) (context.Context, trace2.Span) {
	if t == nil || t.tp == nil {
		return context.Background(), noop.Span{}
	}

	tracer := otel.Tracer(t.ServiceName)
	return tracer.Start(ctx, name)
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
