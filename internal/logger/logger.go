package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger представляет собой глобальный экземпляр логгера
var Logger *zap.Logger

// Config конфигурация для логгера
type Config struct {
	Level       string `yaml:"level" json:"level"`               // debug, info, warn, error, fatal
	Format      string `yaml:"format" json:"format"`             // json или text
	Filename    string `yaml:"filename" json:"filename"`         // путь к файлу лога
	MaxSize     int    `yaml:"max_size" json:"max_size"`         // максимальный размер файла в MB
	MaxAge      int    `yaml:"max_age" json:"max_age"`           // максимальный возраст файла в днях
	MaxBackups  int    `yaml:"max_backups" json:"max_backups"`   // максимальное количество резервных файлов
	Compress    bool   `yaml:"compress" json:"compress"`         // сжимать старые файлы
	LocalTime   bool   `yaml:"local_time" json:"local_time"`     // использовать локальное время для имен файлов
	RotateDaily bool   `yaml:"rotate_daily" json:"rotate_daily"` // ротация по дням
	EnableStdout bool  `yaml:"enable_stdout" json:"enable_stdout"` // дублировать в stdout для разработки
}

// Init инициализирует логгер
func Init(cfg *Config) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zap.InfoLevel
	}

	// Настройка encoder в зависимости от формата
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if cfg.Format == "text" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Настройка writer
	var writeSyncers []zapcore.WriteSyncer

	// Добавляем файловый вывод если указан файл
	if cfg.Filename != "" {
		// Создаем директорию для логов если она не существует
		if err := os.MkdirAll(filepath.Dir(cfg.Filename), 0755); err != nil {
			return err
		}

		// Используем lumberjack для ротации логов
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			Compress:   cfg.Compress,
			LocalTime:  cfg.LocalTime,
		}
		writeSyncers = append(writeSyncers, zapcore.AddSync(lumberJackLogger))
	}

	// Добавляем stdout если включен или если файл не указан
	if cfg.EnableStdout || cfg.Filename == "" {
		writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
	}

	// Если нет ни одного writer'а, используем stdout по умолчанию
	if len(writeSyncers) == 0 {
		writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
	}

	// Объединяем все writer'ы
	writeSyncer := zapcore.NewMultiWriteSyncer(writeSyncers...)

	// Создаем core
	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		level,
	)

	// Создаем логгер
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// InitDefault инициализирует логгер с настройками по умолчанию
func InitDefault() error {
	cfg := &Config{
		Level:       "info",
		Format:      "json",
		Filename:    "logs/app.log", // по умолчанию сохраняем в файл
		MaxSize:     100,            // 100 MB
		MaxAge:      30,             // 30 дней
		MaxBackups:  5,              // 5 резервных файлов
		Compress:    true,           // сжимать старые файлы
		LocalTime:   true,           // использовать локальное время
		RotateDaily: false,          // ротация только по размеру
		EnableStdout: false,        // по умолчанию не дублируем в stdout
	}
	return Init(cfg)
}

// CreateFileLogger создает отдельный логгер для записи в файл
func CreateFileLogger(filename string, level zapcore.Level) (*zap.Logger, error) {
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
