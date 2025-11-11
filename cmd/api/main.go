package main

import (
	"context"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // Подключаем pprof
	"os"
	"os/signal"
	"syscall"
	"time"

	todopb "github.com/akozadaev/go_todo_service/api/proto"
	"github.com/akozadaev/go_todo_service/config"
	"github.com/akozadaev/go_todo_service/internal/database"
	"github.com/akozadaev/go_todo_service/internal/handler"
	"github.com/akozadaev/go_todo_service/internal/logger"
	"github.com/akozadaev/go_todo_service/internal/middleware"
	"github.com/akozadaev/go_todo_service/internal/repository"
	"github.com/akozadaev/go_todo_service/internal/service"
	"github.com/akozadaev/go_todo_service/pkg/trace"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем логгер
	loggerCfg := &logger.Config{
		Level:        cfg.Logger.Level,
		Format:       cfg.Logger.Format,
		Filename:     cfg.Logger.Filename,
		MaxSize:      cfg.Logger.MaxSize,
		MaxAge:       cfg.Logger.MaxAge,
		MaxBackups:   cfg.Logger.MaxBackups,
		Compress:     cfg.Logger.Compress,
		LocalTime:    cfg.Logger.LocalTime,
		RotateDaily:  cfg.Logger.RotateDaily,
		EnableStdout: cfg.Logger.EnableStdout,
	}
	if err := logger.Init(loggerCfg); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Инициализируем специализированные логгеры
	if err := logger.InitSpecializedLoggers("logs"); err != nil {
		if logger.Logger != nil {
			logger.Logger.Fatal("Failed to initialize specialized loggers", zap.Error(err))
		}
		log.Fatalf("Failed to initialize specialized loggers: %v", err)
	}
	defer logger.SyncSpecialized()

	// Инициализируем трассировку
	traceClient, err := trace.NewTraceClient()
	if err != nil {
		if logger.Logger != nil {
			logger.Logger.Fatal("Failed to initialize trace client", zap.Error(err))
		}
		log.Fatalf("Failed to initialize trace client: %v", err)
	}
	defer func() {
		if traceClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = traceClient.Shutdown(ctx)
		}
	}()

	// Логируем запуск приложения
	if logger.Logger != nil {
		logger.Logger.Info("Starting application",
			zap.String("server", cfg.Server.GetServerAddress()),
			zap.String("log_level", cfg.Logger.Level),
			zap.String("pprof_url", "http://"+cfg.Server.GetServerAddress()+"/debug/pprof/"))
	}

	// Подключаемся к базе данных
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		if logger.Logger != nil {
			logger.Logger.Fatal("Failed to connect to database", zap.Error(err))
		}
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Выполняем миграции
	if err = database.AutoMigrate(db); err != nil {
		if logger.Logger != nil {
			logger.Logger.Fatal("Failed to migrate database", zap.Error(err))
		}
		log.Fatalf("F- option go_package = \"github.com/akozadaev/go_todo_servicailed to migrate database: %v", err)
	}

	// Инициализируем слои приложения (Dependency Injection)
	todoRepo := repository.NewTodoRepository(db)
	todoService := service.NewTodoService(todoRepo)
	todoHandler := handler.NewTodoHandler(todoService)
	todoGRPCHandler := handler.NewTodoGRPCHandler(todoService)
	healthHandler := handler.NewHealthHandler(db)

	// Настраиваем Gin
	gin.SetMode(gin.DebugMode) // Используем gin.DebugMode для разработки,  ReleaseMode для релиза
	router := gin.New()

	// Добавляем middleware
	router.Use(gin.Recovery())                   // Встроенный recovery middleware
	router.Use(middleware.RequestIDMiddleware()) // Добавляем request ID
	router.Use(middleware.Logger())              // Кастомный logger
	router.Use(middleware.CORS())                // CORS support

	// Регистрируем health endpoints
	healthHandler.RegisterRoutes(router)

	// Регистрируем API endpoints
	apiGroup := router.Group("/api/v1", middleware.RequireUser())
	if traceClient != nil {
		apiGroup.Use(traceClient.MiddleWareTrace())
	}
	todoHandler.RegisterRoutes(apiGroup)

	// OpenAPI документация доступна через файловый сервер
	// Используйте: make swag-local для просмотра документации

	// Регистрируем pprof endpoints для профилирования
	pprofGroup := router.Group("/debug/pprof")
	pprofGroup.Use(middleware.PProfAuth()) // Защищаем pprof endpoints
	{
		pprofGroup.GET("/", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/cmdline", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/profile", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.POST("/symbol", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/symbol", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/trace", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/allocs", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/block", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/goroutine", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/heap", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/mutex", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		pprofGroup.GET("/threadcreate", gin.WrapF(http.DefaultServeMux.ServeHTTP))
	}

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:         cfg.Server.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Создаем gRPC сервер
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer) // Включаем reflection для grpcurl и других инструментов

	// Регистрируем gRPC сервис
	todopb.RegisterTodoServiceServer(grpcServer, todoGRPCHandler)

	// Запускаем HTTP сервер в отдельной горутине
	go func() {
		if logger.Logger != nil {
			logger.Logger.Info("Starting HTTP server", zap.String("address", server.Addr))
		} else {
			log.Printf("Starting HTTP server on %s", server.Addr)
		}
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if logger.Logger != nil {
				logger.Logger.Fatal("Failed to start HTTP server", zap.Error(err))
			}
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Запускаем gRPC сервер в отдельной горутине
	go func() {
		grpcAddr := cfg.Server.GetGRPCAddress()
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			if logger.Logger != nil {
				logger.Logger.Fatal("Failed to listen on gRPC address", zap.String("address", grpcAddr), zap.Error(err))
			}
			log.Fatalf("Failed to listen on gRPC address %s: %v", grpcAddr, err)
		}

		if logger.Logger != nil {
			logger.Logger.Info("Starting gRPC server", zap.String("address", grpcAddr))
		} else {
			log.Printf("Starting gRPC server on %s", grpcAddr)
		}

		if err = grpcServer.Serve(lis); err != nil {
			if logger.Logger != nil {
				logger.Logger.Fatal("Failed to serve gRPC", zap.Error(err))
			}
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if logger.Logger != nil {
		logger.Logger.Info("Shutting down server...")
	} else {
		log.Println("Shutting down server...")
	}

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Останавливаем HTTP сервер (ждем завершения активных запросов)
	if err = server.Shutdown(ctx); err != nil {
		if logger.Logger != nil {
			logger.Logger.Error("HTTP server forced to shutdown", zap.Error(err))
		} else {
			log.Printf("HTTP server forced to shutdown: %v", err)
		}
	}

	// Останавливаем gRPC сервер
	if logger.Logger != nil {
		logger.Logger.Info("Stopping gRPC server...")
	} else {
		log.Println("Stopping gRPC server...")
	}

	// GracefulStop ждет завершения всех активных запросов
	grpcServer.GracefulStop()

	if logger.Logger != nil {
		logger.Logger.Info("gRPC server stopped")
	} else {
		log.Println("gRPC server stopped")
	}

	// Закрываем соединение с базой данных после остановки сервера
	if logger.Logger != nil {
		logger.Logger.Info("Closing database connection...")
	} else {
		log.Println("Closing database connection...")
	}
	if err := database.Close(db); err != nil {
		if logger.Logger != nil {
			logger.Logger.Error("Error closing database", zap.Error(err))
		} else {
			log.Printf("Error closing database: %v", err)
		}
	}

	if logger.Logger != nil {
		logger.Logger.Info("Server exited gracefully")
	} else {
		log.Println("Server exited gracefully")
	}
}
