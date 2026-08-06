package database

import (
	"fmt"

	"github.com/akozadaev/go_todo_service/config"
	"github.com/akozadaev/go_todo_service/internal/logger"
	"github.com/akozadaev/go_todo_service/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// NewPostgresDB создает подключение к PostgreSQL
func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	// Получаем доступ к sql.DB для настройки пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Настройка пула соединений
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	if logger.Logger != nil {
		logger.Logger.Info("Successfully connected to database")
	}

	return db, nil
}

// AutoMigrate выполняет автоматическую миграцию моделей
func AutoMigrate(db *gorm.DB) error {
	if logger.Logger != nil {
		logger.Logger.Info("Running auto migration...")
	}

	if err := db.AutoMigrate(
		&model.Todo{},
	); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	if logger.Logger != nil {
		logger.Logger.Info("Auto migration completed successfully")
	}
	return nil
}

// Close закрывает соединение с базой данных
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
