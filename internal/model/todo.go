package model

import (
	"time"
)

// Todo представляет модель задачи
type Todo struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	Title       string    `gorm:"not null;size:255" json:"title" binding:"required,min=1,max=255"`
	Description string    `gorm:"type:text" json:"description" binding:"max=1000"`
	Done        bool      `gorm:"default:false" json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName переопределяет имя таблицы
func (Todo) TableName() string {
	return "todos"
}

// TodoCreateRequest представляет запрос на создание задачи
type TodoCreateRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
	Done        bool   `json:"done"`
}

// TodoUpdateRequest представляет запрос на обновление задачи
type TodoUpdateRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
	Done        bool   `json:"done"`
}

// ToTodo конвертирует CreateRequest в Todo
func (r *TodoCreateRequest) ToTodo() *Todo {
	return &Todo{
		Title:       r.Title,
		Description: r.Description,
		Done:        r.Done,
	}
}
