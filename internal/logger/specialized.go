package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// SpecializedLoggers содержит специализированные логгеры
type SpecializedLoggers struct {
	HTTP   *zap.Logger
	Error  *zap.Logger
	Access *zap.Logger
	Audit  *zap.Logger
}

// Loggers глобальный экземпляр специализированных логгеров
var Loggers *SpecializedLoggers

// InitSpecializedLoggers инициализирует специализированные логгеры
func InitSpecializedLoggers(baseDir string) error {
	var err error
	Loggers = &SpecializedLoggers{}

	// HTTP логгер для HTTP запросов
	Loggers.HTTP, err = createSpecializedLogger(filepath.Join(baseDir, "http.log"), zapcore.InfoLevel)
	if err != nil {
		return err
	}

	// Error логгер для ошибок
	Loggers.Error, err = createSpecializedLogger(filepath.Join(baseDir, "error.log"), zapcore.ErrorLevel)
	if err != nil {
		return err
	}

	// Access логгер для доступа
	Loggers.Access, err = createSpecializedLogger(filepath.Join(baseDir, "access.log"), zapcore.InfoLevel)
	if err != nil {
		return err
	}

	// Audit логгер для аудита
	Loggers.Audit, err = createSpecializedLogger(filepath.Join(baseDir, "audit.log"), zapcore.InfoLevel)
	if err != nil {
		return err
	}

	return nil
}

// createSpecializedLogger создает специализированный логгер
func createSpecializedLogger(filename string, level zapcore.Level) (*zap.Logger, error) {
	// Создаем директорию для логов если она не существует
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return nil, err
	}

	// Настройка encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Настройка writer с ротацией
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    50,   // 50 MB для специализированных логов
		MaxAge:     15,   // 15 дней
		MaxBackups: 3,    // 3 резервных файла
		Compress:   true, // сжимать старые файлы
		LocalTime:  true, // использовать локальное время
	}
	writeSyncer := zapcore.AddSync(lumberJackLogger)

	// Создаем core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		level,
	)

	// Создаем логгер
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// SyncSpecialized синхронизирует все специализированные логгеры
func SyncSpecialized() {
	if Loggers != nil {
		if Loggers.HTTP != nil {
			_ = Loggers.HTTP.Sync()
		}
		if Loggers.Error != nil {
			_ = Loggers.Error.Sync()
		}
		if Loggers.Access != nil {
			_ = Loggers.Access.Sync()
		}
		if Loggers.Audit != nil {
			_ = Loggers.Audit.Sync()
		}
	}
}

// LogHTTPRequest логирует HTTP запрос в специализированный логгер
func LogHTTPRequest(method, path string, statusCode int, latency time.Duration, clientIP, userAgent string, bodySize int) {
	if Loggers != nil && Loggers.HTTP != nil {
		// Определяем уровень логирования в зависимости от статус кода
		var level zapcore.Level
		switch {
		case statusCode >= 500:
			level = zapcore.ErrorLevel
		case statusCode >= 400:
			level = zapcore.WarnLevel
		default:
			level = zapcore.InfoLevel
		}

		// Логируем HTTP запрос в HTTP логгер
		Loggers.HTTP.Log(level, "HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("body_size", bodySize),
		)

		// Дополнительно логируем ошибки сервера (5xx) в Error логгер
		if statusCode >= 500 && Loggers.Error != nil {
			Loggers.Error.Error("HTTP Server Error",
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("client_ip", clientIP),
				zap.String("user_agent", userAgent),
			)
		}

	}
}

// LogError логирует ошибку в специализированный логгер
func LogError(message string, err error, fields ...zap.Field) {
	if Loggers != nil && Loggers.Error != nil {
		allFields := append([]zap.Field{zap.Error(err)}, fields...)
		Loggers.Error.Error(message, allFields...)
	}
}

// LogAccess логирует событие доступа
func LogAccess(userID, action, resource string, fields ...zap.Field) {
	if Loggers != nil && Loggers.Access != nil {
		allFields := append([]zap.Field{
			zap.String("user_id", userID),
			zap.String("action", action),
			zap.String("resource", resource),
		}, fields...)
		Loggers.Access.Info("Access Event", allFields...)
	}
}

// LogAudit логирует событие аудита
func LogAudit(userID, action, resource string, success bool, fields ...zap.Field) {
	if Loggers != nil && Loggers.Audit != nil {
		allFields := append([]zap.Field{
			zap.String("user_id", userID),
			zap.String("action", action),
			zap.String("resource", resource),
			zap.Bool("success", success),
		}, fields...)
		Loggers.Audit.Info("Audit Event", allFields...)
	}
}
