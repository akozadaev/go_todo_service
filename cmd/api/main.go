package main

import (
	"context"
	"github.com/akozadaev/go_todo_service/config"
	"github.com/akozadaev/go_todo_service/internal/database"
	"github.com/akozadaev/go_todo_service/internal/handler"
	"github.com/akozadaev/go_todo_service/internal/middleware"
	"github.com/akozadaev/go_todo_service/internal/repository"
	"github.com/akozadaev/go_todo_service/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к базе данных
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Выполняем миграции
	if err = database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Инициализируем слои приложения (Dependency Injection)
	todoRepo := repository.NewTodoRepository(db)
	todoService := service.NewTodoService(todoRepo)
	todoHandler := handler.NewTodoHandler(todoService)
	healthHandler := handler.NewHealthHandler(db)

	// Настраиваем Gin
	gin.SetMode(gin.DebugMode) // Используем gin.DebugMode для разработки,  ReleaseMode для релиза
	router := gin.New()

	// Добавляем middleware
	router.Use(gin.Recovery())      // Встроенный recovery middleware
	router.Use(middleware.Logger()) // Кастомный logger
	router.Use(middleware.CORS())   // CORS support

	// Регистрируем health endpoints
	healthHandler.RegisterRoutes(router)

	// Регистрируем API endpoints
	api := router.Group("/api/v1")
	todoHandler.RegisterRoutes(api)

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:         cfg.Server.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Starting server on %s", server.Addr)
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Останавливаем HTTP сервер (ждем завершения активных запросов)
	if err = server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Закрываем соединение с базой данных после остановки сервера
	log.Println("Closing database connection...")
	if err := database.Close(db); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server exited gracefully")
}
