package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type SpecializedLoggers struct {
	HTTP   *zap.Logger
	Error  *zap.Logger
	Access *zap.Logger
	Audit  *zap.Logger
}

var Loggers *SpecializedLoggers

func InitSpecializedLoggers(baseDir string) error {
	var err error
	Loggers = &SpecializedLoggers{}

	Loggers.HTTP, err = createSpecializedLogger(filepath.Join(baseDir, "http.log"), zapcore.InfoLevel)
	if err != nil {
		return err
	}

	Loggers.Error, err = createSpecializedLogger(filepath.Join(baseDir, "error.log"), zapcore.ErrorLevel)
	if err != nil {
		return err
	}

	Loggers.Access, err = createSpecializedLogger(filepath.Join(baseDir, "access.log"), zapcore.InfoLevel)
	if err != nil {
		return err
	}

	Loggers.Audit, err = createSpecializedLogger(filepath.Join(baseDir, "audit.log"), zapcore.InfoLevel)
	if err != nil {
		return err
	}

	return nil
}

func createSpecializedLogger(filename string, level zapcore.Level) (*zap.Logger, error) {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return nil, err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    50,   // 50 MB для специализированных логов
		MaxAge:     15,   // 15 дней
		MaxBackups: 3,    // 3 резервных файла
		Compress:   true, // сжимать старые файлы
		LocalTime:  true, // использовать локальное время
	}
	writeSyncer := zapcore.AddSync(lumberJackLogger)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		level,
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

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

func LogHTTPRequest(method, path string, statusCode int, latency time.Duration, clientIP, userAgent string, bodySize int) {
	if Loggers != nil && Loggers.HTTP != nil {
		var level zapcore.Level
		switch {
		case statusCode >= 500:
			level = zapcore.ErrorLevel
		case statusCode >= 400:
			level = zapcore.WarnLevel
		default:
			level = zapcore.InfoLevel
		}

		Loggers.HTTP.Log(level, "HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("body_size", bodySize),
		)
	}
}

func LogError(message string, err error, fields ...zap.Field) {
	if Loggers != nil && Loggers.Error != nil {
		allFields := append([]zap.Field{zap.Error(err)}, fields...)
		Loggers.Error.Error(message, allFields...)
	}
}

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
