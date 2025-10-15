package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger глобальный экземпляр логгера
var Logger *zap.Logger

type Config struct {
	Level       string `yaml:"level" json:"level"`               // debug, info, warn, error
	Filename    string `yaml:"filename" json:"filename"`         // путь к файлу лога
	MaxSize     int    `yaml:"max_size" json:"max_size"`         // максимальный размер файла в MB
	MaxAge      int    `yaml:"max_age" json:"max_age"`           // максимальный возраст файла в днях
	MaxBackups  int    `yaml:"max_backups" json:"max_backups"`   // максимальное количество резервных файлов
	Compress    bool   `yaml:"compress" json:"compress"`         // сжимать старые файлы
	LocalTime   bool   `yaml:"local_time" json:"local_time"`     // использовать локальное время для имен файлов
	RotateDaily bool   `yaml:"rotate_daily" json:"rotate_daily"` // ротация по дням
}

func Init(cfg *Config) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zap.InfoLevel
	}

	// Настройка encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var writeSyncer zapcore.WriteSyncer

	if cfg.Filename != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Filename), 0755); err != nil {
			return err
		}

		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			Compress:   cfg.Compress,
			LocalTime:  cfg.LocalTime,
		}
		writeSyncer = zapcore.AddSync(lumberJackLogger)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		level,
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

func InitDefault() error {
	cfg := &Config{
		Level:       "info",
		Filename:    "logs/app.log", // по умолчанию сохраняем в файл
		MaxSize:     100,            // 100 MB
		MaxAge:      30,             // 30 дней
		MaxBackups:  5,              // 5 резервных файлов
		Compress:    true,           // сжимать старые файлы
		LocalTime:   true,           // использовать локальное время
		RotateDaily: false,          // ротация только по размеру
	}
	return Init(cfg)
}

// CreateFileLogger отдельный логгер для записи в файл
func CreateFileLogger(filename string, level zapcore.Level) (*zap.Logger, error) {
	// Создаем директорию для логов если она не существует
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return nil, err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100,  // 100 MB
		MaxAge:     30,   // 30 дней
		MaxBackups: 5,    // 5 резервных файлов
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

// Sync синхронизирует буферы логгера
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
