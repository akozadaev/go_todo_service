package main

import (
	"log"
	"os"

	"todo-service/handlers"
	"todo-service/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dsn := "host=" + os.Getenv("DB_HOST") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" port=" + os.Getenv("DB_PORT") +
		" sslmode=disable TimeZone=Europe/Moscow"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Автомиграция
	err = db.AutoMigrate(&models.Todo{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	r := gin.Default()

	handler := &handlers.TodoHandler{DB: db}

	r.GET("/todos", handler.GetTodos)
	r.GET("/todos/:id", handler.GetTodo)
	r.POST("/todos", handler.CreateTodo)
	r.PUT("/todos/:id", handler.UpdateTodo)
	r.DELETE("/todos/:id", handler.DeleteTodo)

	log.Println("Starting server on :8080")
	r.Run(":8080")
}
