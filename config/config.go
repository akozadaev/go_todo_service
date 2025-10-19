package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	Trace    TraceConfig
}

type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxIdleConns int
	MaxOpenConns int
}

type LoggerConfig struct {
	Level       string
	Format      string // json или text
	Filename    string
	MaxSize     int
	MaxAge      int
	MaxBackups  int
	Compress    bool
	LocalTime   bool
	RotateDaily bool
	EnableStdout bool // дублировать в stdout для разработки
}

type TraceConfig struct {
	IsTraceEnabled    bool
	Url               string
	ServiceName       string
	IsHttpBodyEnabled bool
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Пытаемся загрузить .env файл (игнорируем если не найден)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "localhost"),
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 5*time.Second),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", ""),
			DBName:       getEnv("DB_NAME", "todo_db"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 100),
		},
		Logger: LoggerConfig{
			Level:       getEnv("LOG_LEVEL", "info"),
			Format:      getEnv("LOG_FORMAT", "json"),
			Filename:    getEnv("LOG_FILENAME", "logs/app.log"),
			MaxSize:     getIntEnv("LOG_MAX_SIZE", 100),
			MaxAge:      getIntEnv("LOG_MAX_AGE", 30),
			MaxBackups:  getIntEnv("LOG_MAX_BACKUPS", 5),
			Compress:    getBoolEnv("LOG_COMPRESS", true),
			LocalTime:   getBoolEnv("LOG_LOCAL_TIME", true),
			RotateDaily: getBoolEnv("LOG_ROTATE_DAILY", false),
			EnableStdout: getBoolEnv("LOG_ENABLE_STDOUT", false),
		},
		Trace: TraceConfig{
			IsTraceEnabled:    getBoolEnv("TRACE_ENABLED", false),
			Url:               getEnv("TRACE_URL", "http://localhost:14268/api/traces"),
			ServiceName:       getEnv("TRACE_SERVICE_NAME", "go-todo-service"),
			IsHttpBodyEnabled: getBoolEnv("TRACE_HTTP_BODY_ENABLED", false),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	return nil
}

// GetDSN возвращает строку подключения к PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Europe/Moscow",
		c.Host, c.User, c.Password, c.DBName, c.Port, c.SSLMode,
	)
}

// GetServerAddress возвращает адрес сервера
func (c *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := strconv.Atoi(value); err == nil {
			return time.Duration(duration) * time.Second
		}
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
