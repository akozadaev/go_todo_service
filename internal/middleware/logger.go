package middleware

import (
	"time"

	"github.com/akozadaev/go_todo_service/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger - middleware для логирования запросов
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Обрабатываем запрос
		c.Next()

		// Логируем после обработки
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if query != "" {
			path = path + "?" + query
		}

		var logLevel zapcore.Level
		switch {
		case statusCode >= 500:
			logLevel = zapcore.ErrorLevel
		case statusCode >= 400:
			logLevel = zapcore.WarnLevel
		default:
			logLevel = zapcore.InfoLevel
		}

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("body_size", bodySize),
		}

		if logger.Logger != nil {
			logger.Logger.Log(logLevel, "HTTP Request", fields...)
		}

		logger.LogHTTPRequest(method, path, statusCode, latency, clientIP, userAgent, bodySize)
	}
}
