package logger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDKey ключ для хранения request ID в контексте
type RequestIDKey struct{}

// RequestLogger структура для логгера с контекстом запроса
type RequestLogger struct {
	logger *zap.Logger
	fields []zap.Field
}

// NewRequestLogger создает новый логгер для запроса
func NewRequestLogger(ctx context.Context) *RequestLogger {
	fields := []zap.Field{}

	// Добавляем request ID если есть
	if requestID := ctx.Value(RequestIDKey{}); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	return &RequestLogger{
		logger: Logger,
		fields: fields,
	}
}

// With добавляет поля к логгеру
func (rl *RequestLogger) With(fields ...zap.Field) *RequestLogger {
	return &RequestLogger{
		logger: rl.logger,
		fields: append(rl.fields, fields...),
	}
}

// Debug логирует сообщение уровня debug
func (rl *RequestLogger) Debug(msg string, fields ...zap.Field) {
	allFields := append(rl.fields, fields...)
	rl.logger.Debug(msg, allFields...)
}

// Info логирует сообщение уровня info
func (rl *RequestLogger) Info(msg string, fields ...zap.Field) {
	allFields := append(rl.fields, fields...)
	rl.logger.Info(msg, allFields...)
}

// Warn логирует сообщение уровня warn
func (rl *RequestLogger) Warn(msg string, fields ...zap.Field) {
	allFields := append(rl.fields, fields...)
	rl.logger.Warn(msg, allFields...)
}

// Error логирует сообщение уровня error
func (rl *RequestLogger) Error(msg string, fields ...zap.Field) {
	allFields := append(rl.fields, fields...)
	rl.logger.Error(msg, allFields...)
}

// Fatal логирует сообщение уровня fatal
func (rl *RequestLogger) Fatal(msg string, fields ...zap.Field) {
	allFields := append(rl.fields, fields...)
	rl.logger.Fatal(msg, allFields...)
}

// WithError добавляет ошибку к полям
func (rl *RequestLogger) WithError(err error) *RequestLogger {
	return rl.With(zap.Error(err))
}

// WithDuration добавляет продолжительность к полям
func (rl *RequestLogger) WithDuration(duration time.Duration) *RequestLogger {
	return rl.With(zap.Duration("duration", duration))
}

// WithString добавляет строковое поле
func (rl *RequestLogger) WithString(key, value string) *RequestLogger {
	return rl.With(zap.String(key, value))
}

// WithInt добавляет целочисленное поле
func (rl *RequestLogger) WithInt(key string, value int) *RequestLogger {
	return rl.With(zap.Int(key, value))
}

// WithUint добавляет беззнаковое целочисленное поле
func (rl *RequestLogger) WithUint(key string, value uint) *RequestLogger {
	return rl.With(zap.Uint(key, value))
}

// WithBool добавляет булево поле
func (rl *RequestLogger) WithBool(key string, value bool) *RequestLogger {
	return rl.With(zap.Bool(key, value))
}

// GenerateRequestID генерирует новый request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// AddRequestIDToContext добавляет request ID в контекст
func AddRequestIDToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey{}, requestID)
}

// GetRequestIDFromContext получает request ID из контекста
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey{}).(string)
	return requestID, ok
}

