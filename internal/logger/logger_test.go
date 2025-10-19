package logger

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger(t *testing.T) {
	// Создаем наблюдаемый логгер для тестирования
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Временно заменяем глобальный логгер
	originalLogger := Logger
	Logger = logger
	defer func() { Logger = originalLogger }()

	// Создаем контекст с request ID
	requestID := "test-request-123"
	ctx := AddRequestIDToContext(context.Background(), requestID)

	// Создаем request logger
	reqLogger := NewRequestLogger(ctx)

	// Тестируем логирование
	reqLogger.Info("Test message", zap.String("test_field", "test_value"))

	// Проверяем, что сообщение было записано
	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	if log.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", log.Message)
	}

	// Проверяем, что request ID присутствует в полях
	requestIDFound := false
	testFieldFound := false
	for _, field := range log.Context {
		if field.Key == "request_id" && field.String == requestID {
			requestIDFound = true
		}
		if field.Key == "test_field" && field.String == "test_value" {
			testFieldFound = true
		}
	}

	if !requestIDFound {
		t.Error("Request ID not found in log fields")
	}
	if !testFieldFound {
		t.Error("Test field not found in log fields")
	}
}

func TestRequestLoggerWithFields(t *testing.T) {
	// Создаем наблюдаемый логгер для тестирования
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Временно заменяем глобальный логгер
	originalLogger := Logger
	Logger = logger
	defer func() { Logger = originalLogger }()

	// Создаем request logger без контекста
	reqLogger := NewRequestLogger(context.Background())

	// Тестируем методы With*
	reqLogger.WithError(os.ErrNotExist).
		WithDuration(100*time.Millisecond).
		WithString("operation", "test").
		WithInt("count", 42).
		WithUint("id", 123).
		WithBool("success", true).
		Info("Test with fields")

	// Проверяем, что сообщение было записано
	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	expectedFields := map[string]interface{}{
		"error":     os.ErrNotExist.Error(),
		"duration":  "100ms",
		"operation": "test",
		"count":     int64(42),
		"id":        uint64(123),
		"success":   true,
	}

	for key, expectedValue := range expectedFields {
		found := false
		for _, field := range log.Context {
			if field.Key == key {
				found = true
				// Проверяем значение (упрощенная проверка)
				if field.String != "" && field.String != expectedValue.(string) {
					t.Errorf("Field %s: expected %v, got %s", key, expectedValue, field.String)
				}
				break
			}
		}
		if !found {
			t.Errorf("Field %s not found in log", key)
		}
	}
}

func TestRequestIDGeneration(t *testing.T) {
	// Тестируем генерацию request ID
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == id2 {
		t.Error("Generated request IDs should be unique")
	}

	if len(id1) == 0 {
		t.Error("Request ID should not be empty")
	}
}

func TestRequestIDContext(t *testing.T) {
	// Тестируем работу с контекстом
	requestID := "test-request-456"
	ctx := AddRequestIDToContext(context.Background(), requestID)

	retrievedID, ok := GetRequestIDFromContext(ctx)
	if !ok {
		t.Error("Should be able to retrieve request ID from context")
	}

	if retrievedID != requestID {
		t.Errorf("Expected request ID %s, got %s", requestID, retrievedID)
	}

	// Тестируем случай, когда request ID отсутствует
	emptyCtx := context.Background()
	_, ok = GetRequestIDFromContext(emptyCtx)
	if ok {
		t.Error("Should not be able to retrieve request ID from empty context")
	}
}

func TestLoggerInit(t *testing.T) {
	// Тестируем инициализацию логгера с различными конфигурациями
	testCases := []struct {
		name   string
		config *Config
	}{
		{
			name: "JSON format with file",
			config: &Config{
				Level:        "info",
				Format:       "json",
				Filename:     "test.log",
				MaxSize:      10,
				MaxAge:       7,
				MaxBackups:   3,
				Compress:     true,
				LocalTime:    true,
				RotateDaily:  false,
				EnableStdout: false,
			},
		},
		{
			name: "Text format with stdout",
			config: &Config{
				Level:        "debug",
				Format:       "text",
				Filename:     "",
				EnableStdout: true,
			},
		},
		{
			name: "Both file and stdout",
			config: &Config{
				Level:        "warn",
				Format:       "json",
				Filename:     "test.log",
				EnableStdout: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Сохраняем оригинальный логгер
			originalLogger := Logger

			// Инициализируем новый логгер
			err := Init(tc.config)
			if err != nil {
				t.Errorf("Failed to initialize logger: %v", err)
			}

			// Проверяем, что логгер создан
			if Logger == nil {
				t.Error("Logger should not be nil")
			}

			// Восстанавливаем оригинальный логгер
			Logger = originalLogger

			// Очищаем тестовый файл
			if tc.config.Filename != "" {
				os.Remove(tc.config.Filename)
			}
		})
	}
}

